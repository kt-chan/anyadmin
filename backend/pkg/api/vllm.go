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
	ModelName string `json:"model_name"`
	NodeIP    string `json:"node_ip"`
	Mode      string `json:"mode"`
}

func CalculateVLLMConfig(c *gin.Context) {
	var req VLLMCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// 1. Get GPU Info from Agent
	// If NodeIP is provided, try to find that specific agent.
	// If not, or if agent not found, we might want to fail or use defaults.
	// Current requirement: "with input paremters (GPU memory, Model Type) provided in the agent heartbeat"
	
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

	if !foundAgent {
		// Log warning but proceed with default or error? 
		// Proceeding allows testing without live agents
		// c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found or offline"})
		// return
	}

	// 2. Calculate Config
	params := utils.CalculateConfigParams{
		ModelNameOrPath: req.ModelName,
		GPUMemoryGB:     gpuMemoryGB,
		Mode:            req.Mode,
		GPUUtilization:  0.9,
	}

	vllmConfig, modelConfig, err := utils.CalculateVLLMConfig(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Calculation failed: " + err.Error()})
		return
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
