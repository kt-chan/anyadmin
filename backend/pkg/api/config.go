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

	targetIP := config.IP
	found := false
	var updatedConfig global.InferenceConfig
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
		if newCfg.APIKey != "" {
			// Encrypt if not already encrypted (naive check: RSA base64 is long)
			if len(newCfg.APIKey) < 100 {
				if enc, err := utils.EncryptPassword(newCfg.APIKey); err == nil {
					cfg.APIKey = enc
				} else {
					cfg.APIKey = newCfg.APIKey
				}
			} else {
				cfg.APIKey = newCfg.APIKey
			}
		}
		if newCfg.BaseURL != "" {
			cfg.BaseURL = newCfg.BaseURL
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
			if targetIP != "" && node.NodeIP != targetIP {
				continue
			}

			for j, cfg := range node.InferenceCfgs {
				if cfg.Name == config.Name {
					updateFields(&utils.DeploymentNodes[i].InferenceCfgs[j], config)
					updatedConfig = utils.DeploymentNodes[i].InferenceCfgs[j]
					found = true
					break
				}
			}
			if found && targetIP != "" { break }
		}

		if found {
			affectedConfigs = append(affectedConfigs, updatedConfig)
		} else if targetIP != "" {
			for i, node := range utils.DeploymentNodes {
				if node.NodeIP == targetIP {
					newCfg := config
					newCfg.CreatedAt = time.Now()
					newCfg.UpdatedAt = time.Now()
					utils.DeploymentNodes[i].InferenceCfgs = append(utils.DeploymentNodes[i].InferenceCfgs, newCfg)
					updatedConfig = newCfg
					affectedConfigs = append(affectedConfigs, newCfg)
					found = true
					break
				}
			}
		}
	}, true)

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target service not found"})
		return
	}

	if saveErr != nil {
		fmt.Printf("[DEBUG] SaveToFile Error: %v\n", saveErr)
	}

	// Trigger LiteLLM Update and Agent update
	go func() {
		service.SyncLiteLLMConfig("") 
		for _, cfg := range affectedConfigs {
			if cfg.IsManaged && cfg.Engine == "vLLM" {
				agentConfig := make(map[string]string)
				if cfg.ModelName != "" { agentConfig["model_name"] = cfg.ModelName }
				service.UpdateVLLMConfig(cfg.IP, cfg.Name, agentConfig, true)
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
				if cfg.Name == name || strings.HasPrefix(cfg.Name, name+":") {
					found = true
					targets = append(targets, deletionTarget{NodeIP: node.NodeIP, Name: cfg.Name})
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

	go func() {
		for _, target := range targets {
			service.ControlContainer(target.Name, "down", target.NodeIP)
		}
		service.SyncLiteLLMConfig("")
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

	go func() {
		for _, target := range targets {
			service.ControlContainer(target.Name, "down", target.NodeIP)
		}
		service.SyncLiteLLMConfig("")
	}()

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "删除服务", "彻底移除了 RAG 应用配置及关联容器: "+name, "Warning")
	c.JSON(http.StatusOK, gin.H{"message": "RAG 应用及其关联容器已彻底删除"})
}

func SaveSystemConfig(c *gin.Context) {
	var req SystemConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	utils.ExecuteWrite(func() {
		utils.MgmtHost = req.MgmtHost
		utils.MgmtPort = req.MgmtPort
	}, true)
	c.JSON(http.StatusOK, gin.H{"message": "System config saved"})
}

type SystemConfig struct {
	MgmtHost string `json:"mgmt_host"`
	MgmtPort string `json:"mgmt_port"`
}

func GetServicesConfig(c *gin.Context) {
	type ServiceInstance struct {
		NodeIP    string      `json:"node_ip"`
		Type      string      `json:"type"` 
		Port      string      `json:"port"`
		IsManaged bool        `json:"is_managed"`
		Config    interface{} `json:"config"`
	}
	groupedServices := make(map[string][]ServiceInstance)

	utils.ExecuteRead(func() {
		for _, node := range utils.DeploymentNodes {
			for _, cfg := range node.InferenceCfgs {
				displayName := cfg.Name
				if parts := strings.Split(cfg.Name, ":"); len(parts) == 2 {
					displayName = parts[0]
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

	updateRagFields := func(cfg *global.RagAppConfig, newCfg global.RagAppConfig) {
		if newCfg.StorageDir != "" { cfg.StorageDir = newCfg.StorageDir }
		if newCfg.LLMProvider != "" { cfg.LLMProvider = newCfg.LLMProvider }
		if newCfg.VectorDB != "" { cfg.VectorDB = newCfg.VectorDB }
		if newCfg.GenericOpenAIBasePath != "" { cfg.GenericOpenAIBasePath = newCfg.GenericOpenAIBasePath }
		if newCfg.GenericOpenAIModelPref != "" { cfg.GenericOpenAIModelPref = newCfg.GenericOpenAIModelPref }
		if newCfg.GenericOpenAIKey != "" {
			if len(newCfg.GenericOpenAIKey) < 100 {
				if enc, err := utils.EncryptPassword(newCfg.GenericOpenAIKey); err == nil {
					cfg.GenericOpenAIKey = enc
				} else {
					cfg.GenericOpenAIKey = newCfg.GenericOpenAIKey
				}
			} else {
				cfg.GenericOpenAIKey = newCfg.GenericOpenAIKey
			}
		}
		if newCfg.GenericOpenAIModelTokenLimit > 0 { cfg.GenericOpenAIModelTokenLimit = newCfg.GenericOpenAIModelTokenLimit }
		if newCfg.GenericOpenAIMaxTokens > 0 { cfg.GenericOpenAIMaxTokens = newCfg.GenericOpenAIMaxTokens }
		cfg.UpdatedAt = time.Now()
	}

	utils.ExecuteWrite(func() {
		for i, node := range utils.DeploymentNodes {
			if config.Host != "" && node.NodeIP != config.Host {
				continue
			}
			for j, cfg := range node.RagAppCfgs {
				if cfg.Name == config.Name {
					updateRagFields(&utils.DeploymentNodes[i].RagAppCfgs[j], config)
					updatedConfig = utils.DeploymentNodes[i].RagAppCfgs[j]
					found = true
					break
				}
			}
			if config.Host != "" && found { break }
		}
		if found {
			affectedConfigs = append(affectedConfigs, updatedConfig)
		}
	}, true)

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	go func() {
		for _, cfg := range affectedConfigs {
			agentConfig := make(map[string]string)
			if cfg.LLMProvider != "" { agentConfig["LLM_PROVIDER"] = cfg.LLMProvider }
			if cfg.GenericOpenAIBasePath != "" { agentConfig["GENERIC_OPEN_AI_BASE_PATH"] = cfg.GenericOpenAIBasePath }
			if cfg.GenericOpenAIModelPref != "" { agentConfig["GENERIC_OPEN_AI_MODEL_PREF"] = cfg.GenericOpenAIModelPref }
			if cfg.GenericOpenAIKey != "" {
				decryptedKey := cfg.GenericOpenAIKey
				if dec, err := utils.DecryptPassword(cfg.GenericOpenAIKey); err == nil {
					decryptedKey = dec
				}
				agentConfig["GENERIC_OPEN_AI_API_KEY"] = decryptedKey
			}
			service.UpdateAnythingLLMConfig(cfg.Host, cfg.Name, agentConfig, true)
		}
	}()

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "修改配置", "保存了应用 "+config.Name+" 的配置", "Info")
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
	utils.ExecuteWrite(func() {
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
	c.JSON(http.StatusOK, gin.H{"message": "Agent config saved"})
}
