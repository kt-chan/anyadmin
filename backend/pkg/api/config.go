package api

import (
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/mockdata"
	"anyadmin-backend/pkg/service"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SaveInferenceConfig(c *gin.Context) {
	var config global.InferenceConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		fmt.Printf("[DEBUG] SaveInferenceConfig Bind Error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mockdata.Mu.Lock()
	found := false
	for i, cfg := range mockdata.InferenceCfgs {
		match := false
		if config.Name != "" && cfg.Name == config.Name {
			match = true
		} else if config.IP != "" && cfg.IP == config.IP {
			match = true
		} else if config.ModelName != "" && (cfg.ModelName == config.ModelName || cfg.ModelPath == config.ModelName) {
			match = true
		}

		if match {
			// Update existing fields
			if config.Engine != "" {
				mockdata.InferenceCfgs[i].Engine = config.Engine
			}
			if config.ModelName != "" {
				mockdata.InferenceCfgs[i].ModelName = config.ModelName
			}
			if config.ModelPath != "" {
				mockdata.InferenceCfgs[i].ModelPath = config.ModelPath
			}
			if config.IP != "" {
				mockdata.InferenceCfgs[i].IP = config.IP
			}
			if config.Port != "" {
				mockdata.InferenceCfgs[i].Port = config.Port
			}
			if config.Mode != "" {
				mockdata.InferenceCfgs[i].Mode = config.Mode
			}

			if config.MaxModelLen > 0 {
				mockdata.InferenceCfgs[i].MaxModelLen = config.MaxModelLen
			}
			if config.MaxNumSeqs > 0 {
				mockdata.InferenceCfgs[i].MaxNumSeqs = config.MaxNumSeqs
			}
			if config.MaxNumBatchedTokens > 0 {
				mockdata.InferenceCfgs[i].MaxNumBatchedTokens = config.MaxNumBatchedTokens
			}
			if config.GpuMemoryUtilization > 0 {
				mockdata.InferenceCfgs[i].GpuMemoryUtilization = config.GpuMemoryUtilization
			}

			mockdata.InferenceCfgs[i].UpdatedAt = time.Now()
			config = mockdata.InferenceCfgs[i] // Use the updated one for subsequent logic
			found = true
			break
		}
	}
	if !found {
		config.CreatedAt = time.Now()
		config.UpdatedAt = time.Now()
		mockdata.InferenceCfgs = append(mockdata.InferenceCfgs, config)
	}
	mockdata.Mu.Unlock()

	if err := mockdata.SaveToFile(); err != nil {
		fmt.Printf("[DEBUG] SaveToFile Error: %v\n", err)
	}

	// Trigger Agent update if applicable
	go func() {
		// Find agent IP
		nodeIP := config.IP // Use config.IP if available
		if nodeIP == "" {
			// Try to find from known agents
			agents := service.GetAllAgents()
			if len(agents) > 0 {
				nodeIP = agents[0].NodeIP // Fallback to first agent
			}
		}

		if nodeIP != "" {
			agentConfig := make(map[string]string)
			if config.MaxModelLen > 0 {
				agentConfig["VLLM_MAX_MODEL_LEN"] = fmt.Sprintf("%d", config.MaxModelLen)
			}
			if config.MaxNumSeqs > 0 {
				agentConfig["VLLM_MAX_NUM_SEQS"] = fmt.Sprintf("%d", config.MaxNumSeqs)
			}
			if config.MaxNumBatchedTokens > 0 {
				agentConfig["VLLM_MAX_NUM_BATCHED_TOKENS"] = fmt.Sprintf("%d", config.MaxNumBatchedTokens)
			}
			if config.GpuMemoryUtilization > 0 {
				agentConfig["VLLM_GPU_MEMORY_UTILIZATION"] = fmt.Sprintf("%.2f", config.GpuMemoryUtilization)
			}

			if len(agentConfig) > 0 {
				service.UpdateVLLMConfig(nodeIP, agentConfig, true)
			}
		}
	}()

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "修改配置", "保存了模型 "+config.Name+" 的推理参数", "Info")

	c.JSON(http.StatusOK, config)
}

func GetInferenceConfigs(c *gin.Context) {
	mockdata.Mu.Lock()
	defer mockdata.Mu.Unlock()
	c.JSON(http.StatusOK, mockdata.InferenceCfgs)
}

func DeleteInferenceConfig(c *gin.Context) {
	// Mock delete
	// In real logic we would delete from DB and stop container
	username, _ := c.Get("username")
	service.RecordLog(username.(string), "删除服务", "彻底移除了模型配置及其关联容器", "Warning")

	c.JSON(http.StatusOK, gin.H{"message": "服务已彻底删除"})
}
