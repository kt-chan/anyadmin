package service

import (
	"fmt"
	"log"
	"strings"
	"time"
	"regexp"
	"strconv"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/mockdata"
	"anyadmin-backend/pkg/utils"

	"golang.org/x/crypto/ssh"
)

func ControlContainer(containerName string, action string, nodeIP string) error {
	containerNameLower := strings.ToLower(containerName)
	log.Printf("[Container] Action: %s on container: %s (Node: %s)", action, containerName, nodeIP)
	
	if nodeIP == "" || nodeIP == "localhost" || nodeIP == "127.0.0.1" {
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	// Connect SSH
	client, err := GetSSHClient(nodeIP, "22")
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %w", nodeIP, err)
	}
	defer client.Close()

	// Special handling for vLLM Restart to apply configuration
	if containerNameLower == "vllm" && action == "restart" {
		return restartVLLMWithConfig(client, nodeIP, containerName)
	}

	// Default action
	cmd := fmt.Sprintf("docker %s %s", action, containerName)
	output, err := ExecuteCommand(client, cmd)
	if err != nil {
		return fmt.Errorf("remote command failed: %w, output: %s", err, output)
	}

	return nil
}

func restartVLLMWithConfig(client *ssh.Client, nodeIP, containerName string) error {
	// 1. Get Config
	var config global.InferenceConfig
	mockdata.Mu.Lock()
	found := false
	for _, cfg := range mockdata.InferenceCfgs {
		if cfg.Name == "vllm" || cfg.Name == "default" || cfg.Engine == "vLLM" { // Simplified matching
			config = cfg
			found = true
			break
		}
	}
	mockdata.Mu.Unlock()

	if !found {
		log.Println("[Container] No vLLM config found, using simple restart")
		ExecuteCommand(client, "docker restart "+containerName)
		return nil
	}

	// 2. Get Node GPU Info
	agentStatus, exists := GetAgentStatus(nodeIP)
	gpuMemGB := 24.0 // Default fallback
	if exists {
		// Parse GPUStatus string: "1 x NVIDIA ... | Mem: 7683/8188 MB"
		re := regexp.MustCompile(`Mem: \d+/(\d+) MB`)
		matches := re.FindStringSubmatch(agentStatus.GPUStatus)
		if len(matches) > 1 {
			totalMB, _ := strconv.ParseFloat(matches[1], 64)
			gpuMemGB = totalMB / 1024.0
		}
	}

	// 3. Calculate Config
	modelConfig := utils.EstimateModelConfigFromName(config.ModelPath)
	if modelConfig.Name == "" {
		modelConfig.Name = config.ModelPath
	}
	
	gpuConfig := utils.GPUConfig{
		MemoryGB:    gpuMemGB,
		Utilization: 0.9,
		ReservedGB:  1.0,
	}

	var vllmConfig utils.VLLMConfig
	switch config.Mode {
	case "max_token":
		vllmConfig = utils.CalculateMaxTokenConfig(modelConfig, gpuConfig)
	case "max_concurrency":
		vllmConfig = utils.CalculateMaxConcurrencyConfig(modelConfig, gpuConfig)
	case "balanced":
		fallthrough
	default:
		vllmConfig = utils.CalculateBalancedConfig(modelConfig, gpuConfig)
	}
	
	// Ensure SwapSpace is set if enabled (hardcoded to true for safety in this context)
	if gpuMemGB < 16 {
		vllmConfig.SwapSpaceGB = 8
	}
	vllmConfig.EnablePrefixCaching = true

	// 4. Construct Command
	modelArg := config.ModelPath
	
	// Base Args
	args := []string{
		"--model", modelArg,
		"--max-model-len", fmt.Sprintf("%d", vllmConfig.MaxModelLen),
		"--max-num-seqs", fmt.Sprintf("%d", vllmConfig.MaxNumSeqs),
		"--max-num-batched-tokens", fmt.Sprintf("%d", vllmConfig.MaxNumBatchedTokens),
		"--gpu-memory-utilization", fmt.Sprintf("%.2f", vllmConfig.GPUMemoryUtil),
	}
	if vllmConfig.SwapSpaceGB > 0 {
		args = append(args, "--swap-space", fmt.Sprintf("%d", vllmConfig.SwapSpaceGB))
	}
	if vllmConfig.EnablePrefixCaching {
		args = append(args, "--enable-prefix-caching")
	}

	cmdArgs := strings.Join(args, " ")
	
	log.Printf("[Container] Restarting vLLM with args: %s", cmdArgs)

	// 5. Execute Docker Commands
	ExecuteCommand(client, "docker stop "+containerName)
	ExecuteCommand(client, "docker rm "+containerName)

	runCmd := fmt.Sprintf(`docker run -d --name %s --restart unless-stopped --gpus all --ipc=host -p 8000:8000 -v /root/.cache/huggingface:/root/.cache/huggingface vllm/vllm-openai:latest %s`, 
		containerName, cmdArgs)
	
	output, err := ExecuteCommand(client, runCmd)
	if err != nil {
		return fmt.Errorf("failed to start vllm: %w, output: %s", err, output)
	}

	return nil
}