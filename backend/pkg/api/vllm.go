package api

import (
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type VLLMCalculateRequest struct {
	ModelName      string  `json:"model_name"`
	NodeIP         string  `json:"node_ip"`
	Mode           string  `json:"mode"`
	GPUMemorySize  float64 `json:"gpu_memory_size"`
	GPUUtilization float64 `json:"gpu_utilization"`
}

func CalculateVLLMConfig(c *gin.Context) {
	// Toggle debug mode based on query param
	if c.Query("debug") == "true" {
		utils.DebugMode = true
		defer func() { utils.DebugMode = false }()
	}

	var req VLLMCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// 1. Get GPU Info from Agent or Request
	var gpuMemoryGB float64 = 8.0 // Default fallback
	if req.GPUMemorySize > 0 {
		gpuMemoryGB = req.GPUMemorySize
	} else if req.NodeIP != "" {
		if agent, exists := service.GetAgentStatus(req.NodeIP); exists {
			gpuMemoryGB = parseGPUMemory(agent.GPUStatus)
		}
	} else {
		// Try to find any running agent if IP not specified (optional convenience)
		agents := service.GetAllAgents()
		for _, agent := range agents {
			if agent.Status == "Running" || agent.Status == "Healthy" {
				gpuMemoryGB = parseGPUMemory(agent.GPUStatus)
				req.NodeIP = agent.NodeIP // Update for response context if needed
				break
			}
		}
	}

	gpuUtilization := 0.85
	if req.GPUUtilization > 0 {
		gpuUtilization = req.GPUUtilization
	}

	// 1.5 Lookup Real Model Path from utils if available
	// The frontend might send generic container names like "vllm" or "qwen"
	// We want the actual model path from our deployment config
	real_model_path := req.ModelName
	if req.NodeIP != "" {
		utils.ExecuteRead(func() {
			for _, node := range utils.DeploymentNodes {
				if node.NodeIP == req.NodeIP {
					for _, cfg := range node.InferenceCfgs {
						if cfg.ModelPath != "" && (cfg.Engine == "vLLM" || cfg.Name == "vllm" || strings.Contains(strings.ToLower(cfg.Name), "vllm")) {
							real_model_path = cfg.ModelPath
							break
						}
					}
					break
				}
			}
		})
	}

	// 2. Calculate Config
	params := utils.CalculateConfigParams{
		ModelNameOrPath: real_model_path,
		GPUMemoryGB:     gpuMemoryGB,
		Mode:            req.Mode,
		GPUUtilization:  gpuUtilization,
	}

	vllmConfig, modelConfig, err := utils.CalculateVLLMConfig(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Calculation failed: " + err.Error()})
		return
	}

	// 3. Enrich utils and Save to File
	if req.NodeIP != "" {
		utils.ExecuteWrite(func() {
			for i, node := range utils.DeploymentNodes {
				if node.NodeIP == req.NodeIP {
					for j, cfg := range node.InferenceCfgs {
						if cfg.Engine == "vLLM" || cfg.Name == "vllm" {
							if utils.DeploymentNodes[i].InferenceCfgs[j].MaxModelLen == 0 {
								utils.DeploymentNodes[i].InferenceCfgs[j].MaxModelLen = vllmConfig.MaxModelLen
							}
							if utils.DeploymentNodes[i].InferenceCfgs[j].MaxNumSeqs == 0 {
								utils.DeploymentNodes[i].InferenceCfgs[j].MaxNumSeqs = vllmConfig.MaxNumSeqs
							}
							if utils.DeploymentNodes[i].InferenceCfgs[j].MaxNumBatchedTokens == 0 {
								utils.DeploymentNodes[i].InferenceCfgs[j].MaxNumBatchedTokens = vllmConfig.MaxNumBatchedTokens
							}
							if utils.DeploymentNodes[i].InferenceCfgs[j].GpuMemoryUtilization == 0 {
								utils.DeploymentNodes[i].InferenceCfgs[j].GpuMemoryUtilization = vllmConfig.GPUMemoryUtil
							}
							if utils.DeploymentNodes[i].InferenceCfgs[j].ModelName == "" {
								utils.DeploymentNodes[i].InferenceCfgs[j].ModelName = modelConfig.Name
							}
							break
						}
					}
					break
				}
			}
		}, true)
	}

	c.JSON(http.StatusOK, gin.H{
		"vllm_config":  vllmConfig,
		"model_config": modelConfig,
		"gpu_memory":   gpuMemoryGB,
		"node_ip":      req.NodeIP,
	})
}

// Helper to parse GPU memory string from agent status
// e.g., "NVIDIA GeForce RTX 4090 (24GB)" or "Tesla V100 32GB"
func parseGPUMemory(gpuStatus string) float64 {
	// Look for explicit pattern like "24GB", "24 GB", "24G"
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([GM]B)`)
	matches := re.FindStringSubmatch(strings.ToUpper(gpuStatus))
	if len(matches) == 3 {
		val, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			if matches[2] == "MB" {
				return val / 1024
			}
			return val
		}
	}

	// Fallback/Heuristics based on name if no explicit size found
	lower := strings.ToLower(gpuStatus)
	if strings.Contains(lower, "4090") {
		return 24
	}
	if strings.Contains(lower, "3090") {
		return 24
	}
	if strings.Contains(lower, "a100") {
		if strings.Contains(lower, "80g") {
			return 80
		}
		return 40
	}
	if strings.Contains(lower, "v100") {
		return 32 // or 16
	}
	if strings.Contains(lower, "t4") {
		return 16
	}

	return 24.0 // Default
}
