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
		// If targetIP is provided, skip nodes that don't match.
		// If targetIP is empty, we update matching services on ALL nodes (Service Level)
		if targetIP != "" && node.NodeIP != targetIP {
			continue
		}

		for j, cfg := range node.InferenceCfgs {
			match := config.ModelName != "" && cfg.ModelName == config.ModelName

			if match {
				updateFields(&mockdata.DeploymentNodes[i].InferenceCfgs[j], config)
				updatedConfig = mockdata.DeploymentNodes[i].InferenceCfgs[j]
				found = true
				// If we are updating a specific IP, we can break after finding it.
				// But if we are updating Service Level (targetIP == ""), we continue to next nodes.
				if targetIP != "" {
					break
				}
			}
		}
	}

	// If not found, add to the node matching config.IP (only if IP was provided and not found)
	if !found {
		mockdata.Mu.Unlock()
		fmt.Printf("[DEBUG] Target service not found Error at SaveInferenceConfig")
		c.JSON(http.StatusNotFound, gin.H{"error": "Target service not found"})
		return
	}

	// Capture updated configs for notification before unlocking
	var affectedConfigs []global.InferenceConfig
	if targetIP == "" {
		// Global update: we need to find all instances of this service name
		for _, node := range mockdata.DeploymentNodes {
			for _, cfg := range node.InferenceCfgs {
				if cfg.Name == config.Name {
					affectedConfigs = append(affectedConfigs, cfg)
				}
			}
		}
	} else {
		affectedConfigs = append(affectedConfigs, updatedConfig)
	}

	mockdata.Mu.Unlock()

	if err := mockdata.SaveToFile(); err != nil {
		fmt.Printf("[DEBUG] SaveToFile Error: %v\n", err)
	}

	// Trigger Agent update for ALL affected nodes
	go func() {
		for _, cfg := range affectedConfigs {
			// Find agent IP
			nodeIP := cfg.IP
			if nodeIP != "" {
				agentConfig := make(map[string]string)
				if cfg.MaxModelLen > 0 {
					agentConfig["VLLM_MAX_MODEL_LEN"] = fmt.Sprintf("%d", cfg.MaxModelLen)
				}
				if cfg.MaxNumSeqs > 0 {
					agentConfig["VLLM_MAX_NUM_SEQS"] = fmt.Sprintf("%d", cfg.MaxNumSeqs)
				}
				if cfg.MaxNumBatchedTokens > 0 {
					agentConfig["VLLM_MAX_NUM_BATCHED_TOKENS"] = fmt.Sprintf("%d", cfg.MaxNumBatchedTokens)
				}
				if cfg.GpuMemoryUtilization > 0 {
					agentConfig["VLLM_GPU_MEMORY_UTILIZATION"] = fmt.Sprintf("%.2f", cfg.GpuMemoryUtilization)
				}

				if len(agentConfig) > 0 {
					fmt.Printf("[DEBUG] Triggering VLLM Config Update for Node: %s\n", nodeIP)
					service.UpdateVLLMConfig(nodeIP, agentConfig, true)
				}
			}
		}
	}()

	username, _ := c.Get("username")
	targetStr := "全局"
	if targetIP != "" {
		targetStr = targetIP
	}
	service.RecordLog(username.(string), "修改配置", "保存了模型 "+config.Name+" 的推理参数 (目标: "+targetStr+")", "Info")

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

// System Config
type SystemConfig struct {
	MgmtHost string `json:"mgmt_host"`
	MgmtPort string `json:"mgmt_port"`
}

func SaveSystemConfig(c *gin.Context) {
	var req SystemConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mockdata.Mu.Lock()
	mockdata.MgmtHost = req.MgmtHost
	mockdata.MgmtPort = req.MgmtPort
	mockdata.Mu.Unlock()

	if err := mockdata.SaveToFile(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "System config saved"})
}

func GetServicesConfig(c *gin.Context) {
	mockdata.Mu.Lock()
	defer mockdata.Mu.Unlock()

	// Create a grouped view of services
	// Map[ServiceName] -> []ServiceInstance
	type ServiceInstance struct {
		NodeIP string      `json:"node_ip"`
		Type   string      `json:"type"` // "vLLM", "AnythingLLM", etc.
		Port   string      `json:"port"`
		Config interface{} `json:"config"`
	}
	groupedServices := make(map[string][]ServiceInstance)

	for _, node := range mockdata.DeploymentNodes {
		// Inference Services
		for _, cfg := range node.InferenceCfgs {
			instance := ServiceInstance{
				NodeIP: node.NodeIP,
				Type:   "vLLM",
				Port:   cfg.Port,
				Config: cfg,
			}
			groupedServices[cfg.Name] = append(groupedServices[cfg.Name], instance)
		}
		// RAG Apps
		for _, cfg := range node.RagAppCfgs {
			instance := ServiceInstance{
				NodeIP: node.NodeIP,
				Type:   "AnythingLLM",
				Port:   cfg.Port,
				Config: cfg,
			}
			groupedServices[cfg.Name] = append(groupedServices[cfg.Name], instance)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"mgmt_host":        mockdata.MgmtHost,
		"mgmt_port":        mockdata.MgmtPort,
		"nodes":            mockdata.DeploymentNodes,
		"grouped_services": groupedServices,
	})
}

func SaveRagAppConfig(c *gin.Context) {
	var config global.RagAppConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mockdata.Mu.Lock()
	found := false
	var updatedConfig global.RagAppConfig

	// Helper to update fields selectively
	updateRagFields := func(cfg *global.RagAppConfig, newCfg global.RagAppConfig) {
		if newCfg.StorageDir != "" {
			cfg.StorageDir = newCfg.StorageDir
		}
		if newCfg.LLMProvider != "" {
			cfg.LLMProvider = newCfg.LLMProvider
		}
		if newCfg.VectorDB != "" {
			cfg.VectorDB = newCfg.VectorDB
		}
		if newCfg.GenericOpenAIBasePath != "" {
			cfg.GenericOpenAIBasePath = newCfg.GenericOpenAIBasePath
		}
		if newCfg.GenericOpenAIModelPref != "" {
			cfg.GenericOpenAIModelPref = newCfg.GenericOpenAIModelPref
		}
		if newCfg.GenericOpenAIKey != "" {
			cfg.GenericOpenAIKey = newCfg.GenericOpenAIKey
		}
		if newCfg.GenericOpenAIModelTokenLimit > 0 {
			cfg.GenericOpenAIModelTokenLimit = newCfg.GenericOpenAIModelTokenLimit
		}
		if newCfg.GenericOpenAIMaxTokens > 0 {
			cfg.GenericOpenAIMaxTokens = newCfg.GenericOpenAIMaxTokens
		}
		cfg.UpdatedAt = time.Now()
	}

	for i, node := range mockdata.DeploymentNodes {
		// If Host matches NodeIP, or search all if Host is empty (Service Level)
		if config.Host != "" && node.NodeIP != config.Host {
			continue
		}

		for j, cfg := range node.RagAppCfgs {
			if cfg.Name == config.Name {
				// Update fields selectively
				updateRagFields(&mockdata.DeploymentNodes[i].RagAppCfgs[j], config)
				updatedConfig = mockdata.DeploymentNodes[i].RagAppCfgs[j]
				found = true
				// For this node, we found the service. Break inner loop to move to next node (or finish if specific).
				break
			}
		}

		// If we targeted a specific host and found it, we can stop searching entirely.
		if config.Host != "" && found {
			break
		}
	}

	if !found {
		mockdata.Mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	// Capture affected configs
	var affectedConfigs []global.RagAppConfig
	if config.Host == "" {
		for _, node := range mockdata.DeploymentNodes {
			for _, cfg := range node.RagAppCfgs {
				if cfg.Name == config.Name {
					affectedConfigs = append(affectedConfigs, cfg)
				}
			}
		}
	} else {
		affectedConfigs = append(affectedConfigs, updatedConfig)
	}

	mockdata.Mu.Unlock()

	if err := mockdata.SaveToFile(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Trigger Agent update for ALL affected nodes
	go func() {
		for _, cfg := range affectedConfigs {
			nodeIP := cfg.Host
			if nodeIP != "" {
				agentConfig := make(map[string]string)
				if cfg.LLMProvider != "" {
					agentConfig["LLM_PROVIDER"] = cfg.LLMProvider
				}
				if cfg.VectorDB != "" {
					agentConfig["VECTOR_DB"] = cfg.VectorDB
				}
				if cfg.GenericOpenAIBasePath != "" {
					agentConfig["GENERIC_OPEN_AI_BASE_PATH"] = cfg.GenericOpenAIBasePath
				}
				if cfg.GenericOpenAIModelPref != "" {
					agentConfig["GENERIC_OPEN_AI_MODEL_PREF"] = cfg.GenericOpenAIModelPref
				}
				if cfg.GenericOpenAIKey != "" {
					agentConfig["GENERIC_OPEN_AI_API_KEY"] = cfg.GenericOpenAIKey
				}
				if cfg.GenericOpenAIModelTokenLimit > 0 {
					agentConfig["GENERIC_OPEN_AI_MODEL_TOKEN_LIMIT"] = fmt.Sprintf("%d", cfg.GenericOpenAIModelTokenLimit)
				}
				if cfg.GenericOpenAIMaxTokens > 0 {
					agentConfig["GENERIC_OPEN_AI_MAX_TOKENS"] = fmt.Sprintf("%d", cfg.GenericOpenAIMaxTokens)
				}

				if len(agentConfig) > 0 {
					fmt.Printf("[DEBUG] Triggering AnythingLLM Config Update for Node: %s\n", nodeIP)
					service.UpdateAnythingLLMConfig(nodeIP, agentConfig, true)
				}
			}
		}
	}()

	username, _ := c.Get("username")
	targetStr := "全局"
	if config.Host != "" {
		targetStr = config.Host
	}
	service.RecordLog(username.(string), "修改配置", "保存了应用 "+config.Name+" 的配置 (目标: "+targetStr+")", "Info")

	c.JSON(http.StatusOK, config)
}

func SaveAgentConfig(c *gin.Context) {
	var req struct {
		TargetNodeIP string             `json:"target_node_ip"`
		Config       global.AgentConfig `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mockdata.Mu.Lock()
	found := false
	for i, node := range mockdata.DeploymentNodes {
		if node.NodeIP == req.TargetNodeIP {
			mockdata.DeploymentNodes[i].AgentConfig = req.Config
			found = true
			break
		}
	}
	mockdata.Mu.Unlock()

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	if err := mockdata.SaveToFile(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Agent config saved"})
}
