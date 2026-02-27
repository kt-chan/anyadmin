package api

import (
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"
	"fmt"
	"net/http"
	"strings"
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

	found := false
	var updatedConfig global.InferenceConfig
	// Capture affected configs for notification
	var affectedConfigs []global.InferenceConfig

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
		if newCfg.GPUUtilization > 0 {
			cfg.GPUUtilization = newCfg.GPUUtilization
		}
		if newCfg.GPUMemoryGB > 0 {
			cfg.GPUMemoryGB = newCfg.GPUMemoryGB
		}
		cfg.UpdatedAt = time.Now()
	}

	saveErr := utils.ExecuteWrite(func() {
		for i, node := range utils.DeploymentNodes {
			// If targetIP is provided, skip nodes that don't match.
			// If targetIP is empty, we update matching services on ALL nodes (Service Level)
			if targetIP != "" && node.NodeIP != targetIP {
				continue
			}

			for j, cfg := range node.InferenceCfgs {
				// Match by ModelName (if provided) or Name
				match := (config.ModelName != "" && cfg.ModelName == config.ModelName) || 
				         (config.Name != "" && cfg.Name == config.Name)

				if match {
					updateFields(&utils.DeploymentNodes[i].InferenceCfgs[j], config)
					updatedConfig = utils.DeploymentNodes[i].InferenceCfgs[j]
					found = true
					// If we are updating a specific IP, we can break after finding it.
					// But if we are updating Service Level (targetIP == ""), we continue to next nodes.
					if targetIP != "" {
						break
					}
				}
			}
		}

		if found {
			if targetIP == "" {
				// Global update: we need to find all instances of this service name
				for _, node := range utils.DeploymentNodes {
					for _, cfg := range node.InferenceCfgs {
						if cfg.Name == config.Name {
							affectedConfigs = append(affectedConfigs, cfg)
						}
					}
				}
			} else {
				affectedConfigs = append(affectedConfigs, updatedConfig)
			}
		} else if targetIP != "" {
			// If not found but IP is provided, try to add to the specific node or create new node
			nodeFound := false
			for i, node := range utils.DeploymentNodes {
				if node.NodeIP == targetIP {
					// Create new config
					newCfg := config
					newCfg.CreatedAt = time.Now()
					newCfg.UpdatedAt = time.Now()
					utils.DeploymentNodes[i].InferenceCfgs = append(utils.DeploymentNodes[i].InferenceCfgs, newCfg)
					
					updatedConfig = newCfg
					affectedConfigs = append(affectedConfigs, newCfg)
					found = true
					nodeFound = true
					break
				}
			}
			
			if !nodeFound {
				// Create new node
				newCfg := config
				newCfg.CreatedAt = time.Now()
				newCfg.UpdatedAt = time.Now()
				
				newNode := global.DeploymentNode{
					NodeIP:        targetIP,
					Hostname:      targetIP, // Default hostname
					InferenceCfgs: []global.InferenceConfig{newCfg},
					RagAppCfgs:    []global.RagAppConfig{},
				}
				utils.DeploymentNodes = append(utils.DeploymentNodes, newNode)
				
				updatedConfig = newCfg
				affectedConfigs = append(affectedConfigs, newCfg)
				found = true
			}
		}
	}, true)

	// If not found, add to the node matching config.IP (only if IP was provided and not found)
	if !found {
		fmt.Printf("[DEBUG] Target service not found Error at SaveInferenceConfig")
		c.JSON(http.StatusNotFound, gin.H{"error": "Target service not found"})
		return
	}

	if saveErr != nil {
		fmt.Printf("[DEBUG] SaveToFile Error: %v\n", saveErr)
	}

	// Trigger Agent update for ALL affected nodes
	go func() {
		for _, cfg := range affectedConfigs {
			// Find agent IP
			nodeIP := cfg.IP
			if nodeIP != "" {
				agentConfig := make(map[string]string)
				if cfg.ModelName != "" {
					agentConfig["model_name"] = cfg.ModelName
				}
				if cfg.MaxModelLen > 0 {
					agentConfig["max_model_len"] = fmt.Sprintf("%d", cfg.MaxModelLen)
				}
				if cfg.MaxNumSeqs > 0 {
					agentConfig["max_num_seqs"] = fmt.Sprintf("%d", cfg.MaxNumSeqs)
				}
				if cfg.MaxNumBatchedTokens > 0 {
					agentConfig["max_num_batched_tokens"] = fmt.Sprintf("%d", cfg.MaxNumBatchedTokens)
				}
				if cfg.GpuMemoryUtilization > 0 {
					agentConfig["gpu_memory_utilization"] = fmt.Sprintf("%.2f", cfg.GpuMemoryUtilization)
				}
				if cfg.GPUMemoryGB > 0 {
					agentConfig["gpu_memory_size"] = fmt.Sprintf("%.1f", cfg.GPUMemoryGB)
				}
				if cfg.Mode != "" {
					agentConfig["mode"] = cfg.Mode
				}

				if len(agentConfig) > 0 {
					fmt.Printf("[DEBUG] Triggering VLLM Config Update for Node: %s, Instance: %s\n", nodeIP, cfg.Name)
					service.UpdateVLLMConfig(nodeIP, cfg.Name, agentConfig, true)
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
	var allConfigs []global.InferenceConfig
	utils.ExecuteRead(func() {
		// Flatten configs for frontend compatibility
		for _, node := range utils.DeploymentNodes {
			allConfigs = append(allConfigs, node.InferenceCfgs...)
		}
	})

	c.JSON(http.StatusOK, allConfigs)
}

func DeleteInferenceConfig(c *gin.Context) {
	name := c.Param("id")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing service name"})
		return
	}

	type deletionTarget struct {
		NodeIP string
		Name   string
	}
	var targets []deletionTarget

	found := false
	err := utils.ExecuteWrite(func() {
		for i, node := range utils.DeploymentNodes {
			var newInferenceCfgs []global.InferenceConfig
			for _, cfg := range node.InferenceCfgs {
				// Match by full name or prefix (project name)
				if cfg.Name == name || strings.HasPrefix(cfg.Name, name+":") {
					found = true
					targets = append(targets, deletionTarget{NodeIP: node.NodeIP, Name: cfg.Name})
					// Record that we found it, but don't add to new list (thus deleting it)
					fmt.Printf("[DEBUG] Deleting Inference Config: %s from Node: %s\n", cfg.Name, node.NodeIP)
					continue
				}
				newInferenceCfgs = append(newInferenceCfgs, cfg)
			}
			utils.DeploymentNodes[i].InferenceCfgs = newInferenceCfgs
		}
	}, true)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration after deletion"})
		return
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	// Trigger container stop/down on agents
	go func() {
		for _, target := range targets {
			fmt.Printf("[DEBUG] Triggering Container Down for Node: %s, Instance: %s\n", target.NodeIP, target.Name)
			if err := service.ControlContainer(target.Name, "down", target.NodeIP); err != nil {
				fmt.Printf("[ERROR] Failed to stop container %s on node %s: %v\n", target.Name, target.NodeIP, err)
			}
		}
	}()

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "删除服务", "彻底移除了推理服务配置及关联容器: "+name, "Warning")

	c.JSON(http.StatusOK, gin.H{"message": "推理服务及其关联容器已彻底删除"})
}

func DeleteRagAppConfig(c *gin.Context) {
	name := c.Param("id")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing app name"})
		return
	}

	type deletionTarget struct {
		NodeIP string
		Name   string
	}
	var targets []deletionTarget

	found := false
	err := utils.ExecuteWrite(func() {
		for i, node := range utils.DeploymentNodes {
			var newRagAppCfgs []global.RagAppConfig
			for _, cfg := range node.RagAppCfgs {
				if cfg.Name == name || strings.HasPrefix(cfg.Name, name+":") {
					found = true
					targets = append(targets, deletionTarget{NodeIP: node.NodeIP, Name: cfg.Name})
					fmt.Printf("[DEBUG] Deleting RAG App Config: %s from Node: %s\n", cfg.Name, node.NodeIP)
					continue
				}
				newRagAppCfgs = append(newRagAppCfgs, cfg)
			}
			utils.DeploymentNodes[i].RagAppCfgs = newRagAppCfgs
		}
	}, true)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration after deletion"})
		return
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "RAG App not found"})
		return
	}

	// Trigger container stop/down on agents
	go func() {
		for _, target := range targets {
			fmt.Printf("[DEBUG] Triggering Container Down for Node: %s, Instance: %s\n", target.NodeIP, target.Name)
			if err := service.ControlContainer(target.Name, "down", target.NodeIP); err != nil {
				fmt.Printf("[ERROR] Failed to stop container %s on node %s: %v\n", target.Name, target.NodeIP, err)
			}
		}
	}()

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "删除服务", "彻底移除了 RAG 应用配置及关联容器: "+name, "Warning")

	c.JSON(http.StatusOK, gin.H{"message": "RAG 应用及其关联容器已彻底删除"})
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

	err := utils.ExecuteWrite(func() {
		utils.MgmtHost = req.MgmtHost
		utils.MgmtPort = req.MgmtPort
	}, true)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "System config saved"})
}

func GetServicesConfig(c *gin.Context) {
	// Create a grouped view of services
	// Map[ServiceName] -> []ServiceInstance
	type ServiceInstance struct {
		NodeIP    string      `json:"node_ip"`
		Type      string      `json:"type"` // "vLLM", "AnythingLLM", etc.
		Port      string      `json:"port"`
		IsManaged bool        `json:"is_managed"`
		Config    interface{} `json:"config"`
	}
	groupedServices := make(map[string][]ServiceInstance)

	utils.ExecuteRead(func() {
		for _, node := range utils.DeploymentNodes {
			// Inference Services
			for _, cfg := range node.InferenceCfgs {
				displayName := cfg.Name
				if parts := strings.Split(cfg.Name, ":"); len(parts) == 2 {
					displayName = parts[0] // Use instance/project name for grouping
				}
				
				instance := ServiceInstance{
					NodeIP:    node.NodeIP,
					Type:      "vLLM",
					Port:      cfg.Port,
					IsManaged: cfg.IsManaged,
					Config:    cfg,
				}
				groupedServices[displayName] = append(groupedServices[displayName], instance)
			}
			// RAG Apps
			for _, cfg := range node.RagAppCfgs {
				displayName := cfg.Name
				if parts := strings.Split(cfg.Name, ":"); len(parts) == 2 {
					displayName = parts[0]
				}
				
				instance := ServiceInstance{
					NodeIP:    node.NodeIP,
					Type:      "AnythingLLM",
					Port:      cfg.Port,
					IsManaged: cfg.IsManaged,
					Config:    cfg,
				}
				groupedServices[displayName] = append(groupedServices[displayName], instance)
			}
		}
	})

	c.JSON(http.StatusOK, gin.H{
		"mgmt_host":        utils.MgmtHost,
		"mgmt_port":        utils.MgmtPort,
		"nodes":            utils.DeploymentNodes,
		"grouped_services": groupedServices,
	})
}

func SaveRagAppConfig(c *gin.Context) {
	var config global.RagAppConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	found := false
	var updatedConfig global.RagAppConfig
	var affectedConfigs []global.RagAppConfig

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

	err := utils.ExecuteWrite(func() {
		for i, node := range utils.DeploymentNodes {
			// If Host matches NodeIP, or search all if Host is empty (Service Level)
			if config.Host != "" && node.NodeIP != config.Host {
				continue
			}

			for j, cfg := range node.RagAppCfgs {
				if cfg.Name == config.Name {
					// Update fields selectively
					updateRagFields(&utils.DeploymentNodes[i].RagAppCfgs[j], config)
					updatedConfig = utils.DeploymentNodes[i].RagAppCfgs[j]
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

		if found {
			if config.Host == "" {
				for _, node := range utils.DeploymentNodes {
					for _, cfg := range node.RagAppCfgs {
						if cfg.Name == config.Name {
							affectedConfigs = append(affectedConfigs, cfg)
						}
					}
				}
			} else {
				affectedConfigs = append(affectedConfigs, updatedConfig)
			}
		}
	}, true)

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	if err != nil {
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
					fmt.Printf("[DEBUG] Triggering AnythingLLM Config Update for Node: %s, Instance: %s\n", nodeIP, cfg.Name)
					service.UpdateAnythingLLMConfig(nodeIP, cfg.Name, agentConfig, true)
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

	found := false
	err := utils.ExecuteWrite(func() {
		for i, node := range utils.DeploymentNodes {
			if node.NodeIP == req.TargetNodeIP {
				utils.DeploymentNodes[i].AgentConfig = req.Config
				found = true
				break
			}
		}
	}, true)

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Agent config saved"})
}
