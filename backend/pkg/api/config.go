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

	// We assume we are saving config for a specific node based on config.IP
	targetIP := config.IP
	// If config.IP is empty (which shouldn't happen for existing config), we might need to search by name?
	// But let's assume IP is provided or we search everything.

	mockdata.Mu.Lock()
	found := false
	var updatedConfig global.InferenceConfig

	// Helper to update fields
	updateFields := func(cfg *global.InferenceConfig, newCfg global.InferenceConfig) {
		if newCfg.Engine != "" {
			cfg.Engine = newCfg.Engine
		}
		if newCfg.ModelName != "" {
			cfg.ModelName = newCfg.ModelName
		}
		if newCfg.ModelPath != "" {
			cfg.ModelPath = newCfg.ModelPath
		}
		if newCfg.IP != "" {
			cfg.IP = newCfg.IP
		}
		if newCfg.Port != "" {
			cfg.Port = newCfg.Port
		}
		if newCfg.Mode != "" {
			cfg.Mode = newCfg.Mode
		}
		if newCfg.MaxModelLen > 0 {
			cfg.MaxModelLen = newCfg.MaxModelLen
		}
		if newCfg.MaxNumSeqs > 0 {
			cfg.MaxNumSeqs = newCfg.MaxNumSeqs
		}
		if newCfg.MaxNumBatchedTokens > 0 {
			cfg.MaxNumBatchedTokens = newCfg.MaxNumBatchedTokens
		}
		if newCfg.GpuMemoryUtilization > 0 {
			cfg.GpuMemoryUtilization = newCfg.GpuMemoryUtilization
		}
		cfg.UpdatedAt = time.Now()
	}

	for i, node := range mockdata.DeploymentNodes {
		// Check if this node matches the target IP if provided
		if targetIP != "" && node.NodeIP != targetIP {
			continue
		}

		for j, cfg := range node.InferenceCfgs {
			match := false
			if config.Name != "" && cfg.Name == config.Name {
				match = true
			} else if config.IP != "" && cfg.IP == config.IP && cfg.Port == config.Port { // More strict matching if possible
				match = true
			} else if config.ModelName == cfg.ModelName {
				match = true
			}

			if match {
				updateFields(&mockdata.DeploymentNodes[i].InferenceCfgs[j], config)
				updatedConfig = mockdata.DeploymentNodes[i].InferenceCfgs[j]
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	// If not found, add to the node matching config.IP
	if !found {
		// Throw Error
		mockdata.Mu.Unlock()
		fmt.Printf("[DEBUG] Target Model not found Error at SaveInferenceConfig")
		c.JSON(http.StatusInternalServerError, config)
	}

	mockdata.Mu.Unlock()

	if err := mockdata.SaveToFile(); err != nil {
		fmt.Printf("[DEBUG] SaveToFile Error: %v\n", err)
	}

	// Trigger Agent update if applicable
	go func() {
		// Find agent IP
		nodeIP := updatedConfig.IP // Use config.IP if available
		if nodeIP != "" {
			agentConfig := make(map[string]string)
			if updatedConfig.MaxModelLen > 0 {
				agentConfig["VLLM_MAX_MODEL_LEN"] = fmt.Sprintf("%d", updatedConfig.MaxModelLen)
			}
			if updatedConfig.MaxNumSeqs > 0 {
				agentConfig["VLLM_MAX_NUM_SEQS"] = fmt.Sprintf("%d", updatedConfig.MaxNumSeqs)
			}
			if updatedConfig.MaxNumBatchedTokens > 0 {
				agentConfig["VLLM_MAX_NUM_BATCHED_TOKENS"] = fmt.Sprintf("%d", updatedConfig.MaxNumBatchedTokens)
			}
			if updatedConfig.GpuMemoryUtilization > 0 {
				agentConfig["VLLM_GPU_MEMORY_UTILIZATION"] = fmt.Sprintf("%.2f", updatedConfig.GpuMemoryUtilization)
			}

			if len(agentConfig) > 0 {
				service.UpdateVLLMConfig(nodeIP, agentConfig, true)
			}
		}
	}()

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "修改配置", "保存了模型 "+config.Name+" 的推理参数", "Info")

	c.JSON(http.StatusOK, updatedConfig)
}

func GetInferenceConfigs(c *gin.Context) {
	mockdata.Mu.Lock()
	defer mockdata.Mu.Unlock()

	// Flatten configs for frontend compatibility
	var allConfigs []global.InferenceConfig
	for _, node := range mockdata.DeploymentNodes {
		allConfigs = append(allConfigs, node.InferenceCfgs...)
	}

	c.JSON(http.StatusOK, allConfigs)
}

func DeleteInferenceConfig(c *gin.Context) {
	// Mock delete
	// In real logic we would delete from DB and stop container
	username, _ := c.Get("username")
	service.RecordLog(username.(string), "删除服务", "彻底移除了模型配置及其关联容器", "Warning")

	// TODO: Implement actual deletion from nested structure if needed
	// For now just return success as per original mock

	c.JSON(http.StatusOK, gin.H{"message": "服务已彻底删除"})
}
