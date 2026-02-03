package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/mockdata"
	"anyadmin-backend/pkg/utils"
)

// AgentStatus represents the state of a remote agent
type AgentStatus struct {
	NodeIP         string                       `json:"node_ip"`
	Hostname       string                       `json:"hostname"`
	LastSeen       time.Time                    `json:"last_seen"`
	Status         string                       `json:"status"`
	CPUUsage       float64                      `json:"cpu_usage"`
	CPUCapacity    string                       `json:"cpu_capacity"`
	MemoryUsage    float64                      `json:"memory_usage"`
	MemoryCapacity string                       `json:"memory_capacity"`
	DockerStatus   string                       `json:"docker_status"`
	DeploymentTime string                       `json:"deployment_time"`
	OSSpec         string                       `json:"os_spec"`
	GPUStatus      string                       `json:"gpu_status"`
	Services       []global.DockerServiceStatus `json:"services"`
}

var (
	// In-memory store for agent status
	agentStatusMap = make(map[string]AgentStatus)
	statusMutex    sync.RWMutex
)

// HandleHeartbeat updates the status of an agent
func HandleHeartbeat(ip, hostname, status string, cpu float64, cpuCap string, mem float64, memCap string, dockerStatus, deploymentTime, osSpec, gpuStatus string, services []global.DockerServiceStatus) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	agentStatusMap[ip] = AgentStatus{
		NodeIP:         ip,
		Hostname:       hostname,
		LastSeen:       time.Now(),
		Status:         status,
		CPUUsage:       cpu,
		CPUCapacity:    cpuCap,
		MemoryUsage:    mem,
		MemoryCapacity: memCap,
		DockerStatus:   dockerStatus,
		DeploymentTime: deploymentTime,
		OSSpec:         osSpec,
		GPUStatus:      gpuStatus,
		Services:       services,
	}
}

// GetAgentStatus retrieves the status of a specific agent
func GetAgentStatus(ip string) (AgentStatus, bool) {
	statusMutex.RLock()
	defer statusMutex.RUnlock()

	status, exists := agentStatusMap[ip]
	return status, exists
}

// GetAllAgents retrieves all known agents
func GetAllAgents() []AgentStatus {
	statusMutex.RLock()
	defer statusMutex.RUnlock()

	agents := make([]AgentStatus, 0, len(agentStatusMap))
	for _, agent := range agentStatusMap {
		agents = append(agents, agent)
	}
	return agents
}

// --- Container Control Logic (Merged from container.go) ---

func ControlContainer(containerName string, action string, nodeIP string) error {
	containerNameLower := strings.ToLower(containerName)
	log.Printf("[Agent] ControlContainer: %s %s on %s", action, containerName, nodeIP)

	agentPort := "9090" // Default agent port
	agentURL := fmt.Sprintf("http://%s:%s/container/control", nodeIP, agentPort)

	// Special handling for vLLM Restart to apply configuration
	if containerNameLower == "vllm" && action == "restart" {
		return restartVLLMViaAgent(nodeIP, containerName, agentURL)
	}

	// Default action
	payload := map[string]string{
		"container_name": containerName,
		"action":         action,
	}
	return sendAgentRequest(agentURL, payload)
}

func sendAgentRequest(url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Use custom client to bypass proxy for internal agent calls
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: nil,
		},
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to connect to agent at %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("agent error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func restartVLLMViaAgent(nodeIP, containerName, agentURL string) error {
	// 1. Get Config
	var config global.InferenceConfig
	mockdata.Mu.Lock()
	found := false
	for _, cfg := range mockdata.InferenceCfgs {
		if cfg.Name == "vllm" || cfg.Name == "default" || cfg.Engine == "vLLM" {
			config = cfg
			found = true
			break
		}
	}
	mockdata.Mu.Unlock()

	if !found {
		log.Println("[Container] No vLLM config found, using simple restart")
		return sendAgentRequest(agentURL, map[string]string{
			"container_name": containerName,
			"action":         "restart",
		})
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

	utilization := 0.9
	if gpuMemGB < 8 {
		utilization = 0.85
	}

	gpuConfig := utils.GPUConfig{
		MemoryGB:    gpuMemGB,
		Utilization: utilization,
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

	log.Printf("[Container] Restarting vLLM via Agent with args: %s", cmdArgs)

	// Construct the full shell command to recreate the container
	// Note: We use ';' to ensure subsequent commands run even if stop/rm fail (e.g. if container doesn't exist)
	// OR use '|| true' to suppress errors for stop/rm.
	// docker stop name || true; docker rm name || true; docker run ...
	
	runCmd := fmt.Sprintf(`docker run -d --name %s --restart unless-stopped --gpus all --ipc=host -p 8000:8000 -v /root/.cache/huggingface:/root/.cache/huggingface vllm/vllm-openai:latest %s`,
		containerName, cmdArgs)

	fullCmd := fmt.Sprintf("docker stop %s || true; docker rm %s || true; %s", containerName, containerName, runCmd)

	payload := map[string]string{
		"container_name": containerName,
		"action":         "recreate_vllm",
		"args":           fullCmd,
	}

	return sendAgentRequest(agentURL, payload)
}