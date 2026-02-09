package api

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

func DeployService(c *gin.Context) {
	var req global.DeploymentConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map DeploymentConfig to InferenceConfig for compatibility
	inferenceConfig := global.InferenceConfig{
		Name:      "vllm", // Standardize name to match container reported by agent
		IP:        req.InferenceHost,
		Port:      req.InferencePort,
		ModelName: req.ModelName,
		// model_path: req.ModelName,
	}

	// Map Engine based on Platform
	switch req.Platform {
	case "nvidia":
		inferenceConfig.Engine = "vLLM"
	case "ascend":
		inferenceConfig.Engine = "MindIE"
		inferenceConfig.Name = "mindie" // Standard name for MindIE container
	default:
		inferenceConfig.Engine = "Unknown"
	}

	// Calculate default balanced config
	// We assume a default GPU memory if not known (e.g. 24GB) or we could query agent status if node exists.
	// For new deployment, we might not have status yet. Use safe defaults or 24GB.
	// In a real scenario, we should probably fetch this from the agent if possible.
	// Let's check if we have agent status for this IP.
	var gpuMem float64 = 8.0 // Default fall back

	calcParams := utils.CalculateConfigParams{
		ModelNameOrPath: req.ModelName,
		GPUMemoryGB:     gpuMem,
		Mode:            "balanced",
		GPUUtilization:  0.85,
	}

	vllmCfg, _, err := utils.CalculateVLLMConfig(calcParams)
	if err == nil {
		inferenceConfig.Mode = calcParams.Mode
		inferenceConfig.GPUMemoryGB = calcParams.GPUMemoryGB
		inferenceConfig.GPUUtilization = calcParams.GPUUtilization
		inferenceConfig.MaxModelLen = vllmCfg.MaxModelLen
		inferenceConfig.MaxNumSeqs = vllmCfg.MaxNumSeqs
		inferenceConfig.MaxNumBatchedTokens = vllmCfg.MaxNumBatchedTokens
		inferenceConfig.GpuMemoryUtilization = vllmCfg.GPUMemoryUtil

	} else {
		log.Printf("Failed to calculate default vllm config: %v", err)
		// Set some safe defaults
		inferenceConfig.Mode = "max_token"
		inferenceConfig.MaxModelLen = 4096
		inferenceConfig.MaxNumSeqs = 20
		inferenceConfig.MaxNumBatchedTokens = 8192
		inferenceConfig.GpuMemoryUtilization = 0.85
		inferenceConfig.GPUMemoryGB = 0
		inferenceConfig.GPUUtilization = 0
	}

	utils.Mu.Lock()

	// Handle Target Nodes Update/Merge
	if req.TargetNodes != "" && req.MgmtHost != "" && req.MgmtPort != "" {
		nodes := strings.Split(req.TargetNodes, "\n")

		// Create a map of existing nodes for easy lookup
		existingNodes := make(map[string]global.DeploymentNode)
		for _, node := range utils.DeploymentNodes {
			existingNodes[node.NodeIP] = node
		}

		var updatedNodes []global.DeploymentNode

		for _, nodeIP := range nodes {
			nodeIP = strings.TrimSpace(nodeIP)
			if nodeIP == "" {
				continue
			}

			// Standardize IP (strip port)
			host, _, err := net.SplitHostPort(nodeIP)
			if err != nil {
				host = nodeIP
			}

			// Async deployment of agent
			go service.DeployAgent(host, req.MgmtHost, req.MgmtPort, req.Mode)

			// Preserve existing config or create new node
			if existing, ok := existingNodes[host]; ok {
				updatedNodes = append(updatedNodes, existing)
				delete(existingNodes, host) // Remove so we know what's left
			} else {
				updatedNodes = append(updatedNodes, global.DeploymentNode{
					NodeIP:        host,
					Hostname:      host,
					InferenceCfgs: []global.InferenceConfig{},
					RagAppCfgs:    []global.RagAppConfig{},
				})
			}
		}

		// Append remaining nodes that weren't in the request?
		// If the user provided a list of "Target Nodes" for this deployment,
		// should we remove others?
		// The wizard seems to define the cluster. Let's keep it sync with the list provided.
		// But be careful not to lose data if the user just omitted one.
		// For safety in this "Add/Deploy" context, we might just append new ones and ensure existing ones are updated.
		// But req.TargetNodes usually comes from the textarea which lists ALL nodes.
		utils.DeploymentNodes = updatedNodes

		utils.MgmtHost = req.MgmtHost
		utils.MgmtPort = req.MgmtPort
	}
	utils.Mu.Unlock() // Unlock briefly as DeployAgent is async and we are done with nodes list structure update for now

	// Mode specific logging or additional actions
	if req.Mode == "new_deployment" {
		log.Printf("[全新部署] 正在初始化节点并拉起服务: %s", req.ModelName)
	} else {
		log.Printf("[接入服务] 正在登记现有服务: %s (%s:%s)", req.ModelName, req.InferenceHost, req.InferencePort)
	}

	utils.Mu.Lock()
	defer utils.Mu.Unlock()

	// Helper to add/update config in a node
	addOrUpdateInferenceCfg := func(nodeIP string, newCfg global.InferenceConfig) {
		for i, node := range utils.DeploymentNodes {
			if node.NodeIP == nodeIP {
				// Check if exists
				found := false
				for j, cfg := range node.InferenceCfgs {
					if cfg.Name == newCfg.Name {
						utils.DeploymentNodes[i].InferenceCfgs[j] = newCfg
						found = true
						break
					}
				}
				if !found {
					newCfg.CreatedAt = time.Now()
					newCfg.UpdatedAt = time.Now()
					utils.DeploymentNodes[i].InferenceCfgs = append(utils.DeploymentNodes[i].InferenceCfgs, newCfg)
				}
				return
			}
		}
	}

	addOrUpdateRagCfg := func(nodeIP string, newCfg global.RagAppConfig) {
		// Apply defaults from .env-anythingllm if missing
		if newCfg.StorageDir == "" {
			newCfg.StorageDir = "/app/server/storage"
		}
		if newCfg.LLMProvider == "" {
			newCfg.LLMProvider = "generic-openai"
		}
		if newCfg.GenericOpenAIBasePath == "" {
			newCfg.GenericOpenAIBasePath = "http://host.docker.internal:8000/v1"
		}
		if newCfg.GenericOpenAIModelPref == "" {
			newCfg.GenericOpenAIModelPref = "Qwen3-1.7B"
		}
		if newCfg.GenericOpenAIModelTokenLimit == 0 {
			newCfg.GenericOpenAIModelTokenLimit = 4098
		}
		if newCfg.GenericOpenAIMaxTokens == 0 {
			newCfg.GenericOpenAIMaxTokens = 2048
		}
		// Hardcoded key from env
		if newCfg.GenericOpenAIKey == "" {
			newCfg.GenericOpenAIKey = "REPLACE_THIS_WITH_YOUR_ACTUAL_KEY"
		}
		if newCfg.VectorDB == "" {
			newCfg.VectorDB = "lancedb"
		}

		for i, node := range utils.DeploymentNodes {
			if node.NodeIP == nodeIP {
				// Check if exists
				found := false
				for j, cfg := range node.RagAppCfgs {
					if cfg.Name == newCfg.Name {
						utils.DeploymentNodes[i].RagAppCfgs[j] = newCfg
						found = true
						break
					}
				}
				if !found {
					newCfg.CreatedAt = time.Now()
					newCfg.UpdatedAt = time.Now()
					utils.DeploymentNodes[i].RagAppCfgs = append(utils.DeploymentNodes[i].RagAppCfgs, newCfg)
				}
				return
			}
		}
	}

	// Add Inference Config
	if req.InferenceHost != "" {
		addOrUpdateInferenceCfg(req.InferenceHost, inferenceConfig)
	}

	if req.EnableRAG && req.RAGHost != "" {
		// addOrUpdateInferenceCfg(req.RAGHost, global.InferenceConfig{
		// 	Name:   "anythingllm",
		// 	Engine: "RAG App",
		// 	IP:     req.RAGHost,
		// 	Port:   req.RAGPort,
		// })
		addOrUpdateRagCfg(req.RAGHost, global.RagAppConfig{
			Name:     "anythingllm",
			Host:     req.RAGHost,
			Port:     req.RAGPort,
			VectorDB: req.VectorDBType, // Assuming linked
		})
	}

	if req.EnableVectorDB && req.VectorDBHost != "" {
		addOrUpdateInferenceCfg(req.VectorDBHost, global.InferenceConfig{
			Name:   strings.ToLower(req.VectorDBType),
			Engine: "Vector DB",
			IP:     req.VectorDBHost,
			Port:   req.VectorDBPort,
		})
	}

	if req.EnableParser && req.ParserHost != "" {
		addOrUpdateInferenceCfg(req.ParserHost, global.InferenceConfig{
			Name:   "mineru",
			Engine: "Parser",
			IP:     req.ParserHost,
			Port:   req.ParserPort,
		})
	}

	// Persist to file
	utils.Mu.Unlock() // avoid double lock in SaveToFile
	utils.SaveToFile()
	utils.Mu.Lock() // re-lock for defer unlock? Actually defer will unlock. We should just not lock around SaveToFile if it locks internally.
	// SaveToFile locks internally. So we should UNLOCK before calling it.
	// I unlocked above. But defer is still scheduled.
	// To avoid panic on defer Unlock of unlocked mutex, I should remove defer or handle carefully.
	// Let's restructure.

	// 记录审计日志
	action := "服务部署"
	detail := "发起部署任务: " + inferenceConfig.Name
	if req.Mode == "integrate_existing" {
		action = "服务接入"
		detail = "接入现有服务: " + inferenceConfig.Name + " (" + inferenceConfig.IP + ":" + inferenceConfig.Port + ")"
	}
	service.RecordLog(c.GetString("username"), action, detail, "Info")

	c.JSON(http.StatusOK, gin.H{
		"message":      "Deployment Started",
		"container_id": "pending",
		"artifacts": gin.H{ // Mock artifacts for frontend display
			"deploy_script.sh": "#!/bin/bash\n# Deployment Script\n# Deployment is now handled automatically by the backend via SSH.\n# You can check the server logs for progress.",
			"config.yaml":      fmt.Sprintf("model: %s\nengine: %s\nhost: %s\nport: %s", inferenceConfig.Name, inferenceConfig.Engine, inferenceConfig.IP, inferenceConfig.Port),
		},
	})
}

// GetNodes returns the list of target nodes (IPs)
func GetNodes(c *gin.Context) {
	utils.Mu.Lock()
	defer utils.Mu.Unlock()

	var nodeIPs []string
	for _, node := range utils.DeploymentNodes {
		nodeIPs = append(nodeIPs, node.NodeIP)
	}

	c.JSON(http.StatusOK, gin.H{"nodes": nodeIPs})
}

type SaveNodesRequest struct {
	Nodes []string `json:"nodes"`
}

// SaveNodes updates the list of target nodes
func SaveNodes(c *gin.Context) {
	var req SaveNodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	utils.Mu.Lock()
	// Merge logic similar to DeployService
	existingNodes := make(map[string]global.DeploymentNode)
	for _, node := range utils.DeploymentNodes {
		existingNodes[node.NodeIP] = node
	}

	var updatedNodes []global.DeploymentNode
	for _, ip := range req.Nodes {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		// Clean IP (remove port if present in list for some reason, though frontend usually separates)
		// But here we expect "IP:Port" or "IP" from textarea.
		// If it has port, we might want to store it?
		// DeploymentNode struct has NodeIP.
		// Let's strip port for storage key, but maybe keep original string if needed?
		// Standardize: Store IP in NodeIP.

		host, _, err := net.SplitHostPort(ip)
		if err != nil {
			host = ip // assume just IP
		}

		if existing, ok := existingNodes[host]; ok {
			updatedNodes = append(updatedNodes, existing)
		} else {
			updatedNodes = append(updatedNodes, global.DeploymentNode{
				NodeIP:        host,
				Hostname:      host,
				InferenceCfgs: []global.InferenceConfig{},
			})
		}
	}
	utils.DeploymentNodes = updatedNodes
	utils.Mu.Unlock()

	utils.SaveToFile()

	c.JSON(http.StatusOK, gin.H{"message": "Success", "nodes": req.Nodes})
}

type VLLMRequest struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

func FetchVLLMModels(c *gin.Context) {
	var req VLLMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Try vLLM service first (usually port 8000)
	vllmUrl := fmt.Sprintf("http://%s:%s/v1/models", req.Host, req.Port)
	resp, err := utils.Get(vllmUrl, 5*time.Second)
	if err == nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			c.Data(resp.StatusCode, "application/json", body)
			return
		}
	}

	// Fallback: Try Agent discovery (usually port 8082)
	agentUrl := fmt.Sprintf("http://%s:8082/models/discover", req.Host)
	resp, err = utils.Get(agentUrl, 5*time.Second)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to both vLLM and Agent service: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response from agent"})
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}

type ConnectionTestRequest struct {
	Type string `json:"type"`
	Host string `json:"host"`
	Port string `json:"port"` // Can be int or string in JSON, but binding as string is safer usually if we convert. Actually let's use string/int interface or just string.
}

func TestServiceConnection(c *gin.Context) {
	var req ConnectionTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeout := 5 * time.Second

	// Handle SSH (Multiple nodes from textarea)
	if req.Type == "ssh" {
		nodes := strings.Split(req.Host, "\n")
		var failedNodes []string
		successCount := 0

		for _, node := range nodes {
			node = strings.TrimSpace(node)
			if node == "" {
				continue
			}

			host, port, err := net.SplitHostPort(node)
			if err != nil {
				// Assume node is just Host, use default port
				host = node
				port = req.Port
				if port == "" {
					port = "22"
				}
			}

			address := net.JoinHostPort(host, port)
			conn, err := net.DialTimeout("tcp", address, timeout)
			if err != nil {
				failedNodes = append(failedNodes, fmt.Sprintf("%s: %v", node, err))
			} else {
				conn.Close()
				successCount++
			}
		}

		if len(failedNodes) > 0 {
			msg := fmt.Sprintf("Failed to connect to %d nodes: %v", len(failedNodes), failedNodes)
			c.JSON(http.StatusOK, gin.H{"status": "error", "message": msg}) // Return 200 with error status so frontend handles it gracefully
		} else if successCount == 0 {
			c.JSON(http.StatusOK, gin.H{"status": "error", "message": "No valid nodes provided"})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "success", "message": fmt.Sprintf("Successfully connected to all %d nodes", successCount)})
		}
		return
	}

	address := fmt.Sprintf("%s:%s", req.Host, req.Port)

	// For vLLM (inference), we might want to check HTTP explicitly
	if req.Type == "inference" {
		url := fmt.Sprintf("http://%s/health", address)
		resp, err := utils.Get(url, timeout)
		if err == nil && resp.StatusCode == http.StatusOK {
			c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Successfully connected to vLLM service"})
			return
		}
		// Fallback to TCP if HTTP fails or for generic check
	}

	// For AnythingLLM (RAG App), check HTTP root
	if req.Type == "rag_app" {
		url := fmt.Sprintf("http://%s", address)
		resp, err := utils.Get(url, timeout)
		if err == nil && resp.StatusCode < 500 {
			c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Successfully connected to AnythingLLM service"})
			return
		}
	}

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "Connection failed: " + err.Error()})
		return
	}
	conn.Close()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Connection established successfully"})
}

type AgentControlRequest struct {
	IP     string `json:"ip" binding:"required"`
	Action string `json:"action" binding:"required"` // start, stop, restart
}

func ControlAgent(c *gin.Context) {
	var req AgentControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.ControlAgent(req.IP, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	actionMsg := req.Action + "ed"
	if req.Action == "stop" {
		actionMsg = "stopped"
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Agent %s %s", req.IP, actionMsg)})
}
