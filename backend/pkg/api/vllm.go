package api

import (
	"anyadmin-backend/pkg/mockdata"
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type VLLMCalculateRequest struct {
	ModelName string `json:"model_name"`
	NodeIP    string `json:"node_ip"`
	Mode      string `json:"mode"`
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

	// 1. Get GPU Info from Agent
	var gpuMemoryGB float64 = 24.0 // Default fallback
	var foundAgent bool = false

	if req.NodeIP != "" {
		if agent, exists := service.GetAgentStatus(req.NodeIP); exists {
			gpuMemoryGB = parseGPUMemory(agent.GPUStatus)
			foundAgent = true
		}
	} else {
		// Try to find any running agent if IP not specified (optional convenience)
		agents := service.GetAllAgents()
		for _, agent := range agents {
			if agent.Status == "Running" || agent.Status == "Healthy" {
				gpuMemoryGB = parseGPUMemory(agent.GPUStatus)
				req.NodeIP = agent.NodeIP // Update for response context if needed
				foundAgent = true
				break
			}
		}
	}

	// 1.5 Lookup Real Model Path from MockData if available
	// The frontend might send generic container names like "vllm" or "qwen"
	// We want the actual model path from our deployment config
	realModelPath := req.ModelName
	if req.NodeIP != "" {
		mockdata.Mu.Lock()
		for _, cfg := range mockdata.InferenceCfgs {
			if cfg.IP == req.NodeIP {
				// If we found a config for this node, use its model path
				// This handles cases where container name is "vllm" but real path is "Qwen/Qwen2.5-7B"
				if cfg.ModelPath != "" {
					realModelPath = cfg.ModelPath
				}
				break
			}
		}
		mockdata.Mu.Unlock()
	}

	if !foundAgent {
		// Log warning but proceed with default or error? 
		// Proceeding allows testing without live agents
		// c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found or offline"})
		// return
	}

	// 2. Calculate Config
	params := utils.CalculateConfigParams{
		ModelNameOrPath: realModelPath,
		GPUMemoryGB:     gpuMemoryGB,
		Mode:            req.Mode,
		GPUUtilization:  0.9,
	}

	vllmConfig, modelConfig, err := utils.CalculateVLLMConfig(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Calculation failed: " + err.Error()})
		return
	}

	// 3. Enrich MockData and Save to File
	if req.NodeIP != "" {
		mockdata.Mu.Lock()
		updated := false
		for i, cfg := range mockdata.InferenceCfgs {
			if cfg.IP == req.NodeIP {
				// Only update if current values are zero or "missing"
				if mockdata.InferenceCfgs[i].MaxModelLen == 0 {
					mockdata.InferenceCfgs[i].MaxModelLen = vllmConfig.MaxModelLen
				}
				if mockdata.InferenceCfgs[i].MaxNumSeqs == 0 {
					mockdata.InferenceCfgs[i].MaxNumSeqs = vllmConfig.MaxNumSeqs
				}
				if mockdata.InferenceCfgs[i].MaxNumBatchedTokens == 0 {
					mockdata.InferenceCfgs[i].MaxNumBatchedTokens = vllmConfig.MaxNumBatchedTokens
				}
				if mockdata.InferenceCfgs[i].GpuMemoryUtilization == 0 {
					mockdata.InferenceCfgs[i].GpuMemoryUtilization = vllmConfig.GPUMemoryUtil
				}
				if mockdata.InferenceCfgs[i].ModelName == "" {
					mockdata.InferenceCfgs[i].ModelName = modelConfig.Name
				}
				updated = true
				break
			}
		}
		mockdata.Mu.Unlock()
		if updated {
			mockdata.SaveToFile()
		}
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
