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

	// Standardize name based on ModelType and unique identifier
	// If ServiceName is provided, use it, otherwise generate one
	baseInstanceName := req.ServiceName
	if baseInstanceName == "" {
		baseInstanceName = fmt.Sprintf("%s-%s", strings.ToLower(req.ModelType), time.Now().Format("01021504"))
	}
	baseInstanceName = strings.ToLower(baseInstanceName)

	// Prefix with type to avoid port/env conflicts
	instanceName := "inf-" + baseInstanceName

	// Map DeploymentConfig to InferenceConfig for compatibility
	inferenceConfig := global.InferenceConfig{
		Name:      instanceName, // This is the unique Instance/Project Name
		ModelType: req.ModelType,
		IsManaged: req.Mode == "new_deployment",
		APIKey:    req.APIKey,
		BaseURL:   req.BaseURL,
		IP:        req.InferenceHost,
		Port:      req.InferencePort,
		ModelName: req.ModelName,
	}

	// Determine Compose Service Name
	composeService := "vllm-llm"
	switch req.ModelType {
	case "embedding":
		composeService = "vllm-embedding"
	case "reranker":
		composeService = "vllm-embedding" // Assuming same service handles both
	case "ocr", "vlm":
		composeService = "vllm-mineru"
	}

	// For managed services, we use the "project:service" convention
	if req.Mode == "new_deployment" {
		inferenceConfig.Engine = "vLLM"
		// The Name stored in DB will be "instance:compose_service" 
		// so agent knows what to do
		inferenceConfig.Name = instanceName + ":" + composeService
	} else {
		inferenceConfig.Engine = "External"
	}

	// ... (Calculation logic remains similar, but ensure it uses inferenceConfig.Name where appropriate)
	var gpuMem float64 = 8.0 
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
		inferenceConfig.MaxModelLen = 4096
		inferenceConfig.MaxNumSeqs = 20
		inferenceConfig.MaxNumBatchedTokens = 8192
		inferenceConfig.GpuMemoryUtilization = 0.85
	}

	utils.ExecuteWrite(func() {
		// Handle Target Nodes Update/Merge
		if req.TargetNodes != "" && req.MgmtHost != "" && req.MgmtPort != "" {
			nodes := strings.Split(req.TargetNodes, "\n")
			existingNodes := make(map[string]global.DeploymentNode)
			for _, node := range utils.DeploymentNodes {
				existingNodes[node.NodeIP] = node
			}

			var updatedNodes []global.DeploymentNode
			for _, nodeIP := range nodes {
				nodeIP = strings.TrimSpace(nodeIP)
				if nodeIP == "" { continue }
				host, _, err := net.SplitHostPort(nodeIP)
				if err != nil { host = nodeIP }

				// Async deployment of agent
				go service.DeployAgent(host, req.MgmtHost, req.MgmtPort, req.Mode)

				if existing, ok := existingNodes[host]; ok {
					updatedNodes = append(updatedNodes, existing)
					delete(existingNodes, host)
				} else {
					updatedNodes = append(updatedNodes, global.DeploymentNode{
						NodeIP:        host,
						Hostname:      host,
						InferenceCfgs: []global.InferenceConfig{},
						RagAppCfgs:    []global.RagAppConfig{},
					})
				}
			}
			utils.DeploymentNodes = updatedNodes
			utils.MgmtHost = req.MgmtHost
			utils.MgmtPort = req.MgmtPort
		}
	}, true)

	// Mode specific logging
	if req.Mode == "new_deployment" {
		log.Printf("[全新部署] 正在初始化节点并拉起新服务: %s (%s)", instanceName, req.ModelName)
	} else {
		log.Printf("[接入服务] 正在登记现有服务: %s (%s:%s)", instanceName, req.InferenceHost, req.InferencePort)
	}

	// Helper to add config (Always ADD now)
	addInferenceCfg := func(nodeIP string, newCfg global.InferenceConfig) {
		for i, node := range utils.DeploymentNodes {
			if node.NodeIP == nodeIP {
				newCfg.CreatedAt = time.Now()
				newCfg.UpdatedAt = time.Now()
				utils.DeploymentNodes[i].InferenceCfgs = append(utils.DeploymentNodes[i].InferenceCfgs, newCfg)
				return
			}
		}
	}

	addRagCfg := func(nodeIP string, newCfg global.RagAppConfig) {
        // ... (Default settings)
		if newCfg.StorageDir == "" { newCfg.StorageDir = "/app/server/storage" }
		if newCfg.LLMProvider == "" { newCfg.LLMProvider = "generic-openai" }
		if newCfg.GenericOpenAIBasePath == "" { newCfg.GenericOpenAIBasePath = "http://host.docker.internal:8000/v1" }
		if newCfg.GenericOpenAIModelPref == "" { newCfg.GenericOpenAIModelPref = "Qwen3-1.7B" }
		if newCfg.GenericOpenAIKey == "" { newCfg.GenericOpenAIKey = "REPLACE_THIS_WITH_YOUR_ACTUAL_KEY" }
		if newCfg.VectorDB == "" { newCfg.VectorDB = "lancedb" }

		for i, node := range utils.DeploymentNodes {
			if node.NodeIP == nodeIP {
				newCfg.CreatedAt = time.Now()
				newCfg.UpdatedAt = time.Now()
				utils.DeploymentNodes[i].RagAppCfgs = append(utils.DeploymentNodes[i].RagAppCfgs, newCfg)
				return
			}
		}
	}

	utils.ExecuteWrite(func() {
		if req.InferenceHost != "" {
			addInferenceCfg(req.InferenceHost, inferenceConfig)
		}
		if req.EnableRAG && req.RAGHost != "" {
			ragInstanceName := "rag-" + baseInstanceName
			addRagCfg(req.RAGHost, global.RagAppConfig{
				Name:      ragInstanceName + ":anythingllm", // Unique project name
				IsManaged: req.Mode == "new_deployment",
				Host:      req.RAGHost,
				Port:      req.RAGPort,
				VectorDB:  req.VectorDBType,
			})
		}

		if req.EnableVectorDB && req.VectorDBHost != "" {
			vdbName := strings.ToLower(req.VectorDBType)
			vdbInstanceName := "vdb-" + baseInstanceName
			name := vdbInstanceName + "-" + vdbName
			if req.Mode == "new_deployment" {
				name = vdbInstanceName + ":" + vdbName
			}
			addInferenceCfg(req.VectorDBHost, global.InferenceConfig{
				Name:      name,
				ModelType: "vectordb",
				IsManaged: req.Mode == "new_deployment",
				Engine:    "Vector DB",
				IP:        req.VectorDBHost,
				Port:      req.VectorDBPort,
			})
		}

		if req.EnableParser && req.ParserHost != "" {
			psrInstanceName := "psr-" + baseInstanceName
			name := psrInstanceName + "-parser"
			if req.Mode == "new_deployment" {
				name = psrInstanceName + ":mineru-api"
			}
			addInferenceCfg(req.ParserHost, global.InferenceConfig{
				Name:      name,
				ModelType: "parser",
				IsManaged: req.Mode == "new_deployment",
				Engine:    "Parser",
				IP:        req.ParserHost,
				Port:      req.ParserPort,
			})
		}
	}, true)

	// Trigger immediate start if managed
	if req.Mode == "new_deployment" {
		go func() {
			// Give agent time to start if it was just deployed
			time.Sleep(10 * time.Second)
			
			if req.InferenceHost != "" {
				configMap := map[string]string{
					"model_name": req.ModelName,
					"port":       req.InferencePort,
				}
				service.UpdateVLLMConfig(req.InferenceHost, inferenceConfig.Name, configMap, true)
			}

			if req.EnableRAG && req.RAGHost != "" {
				ragName := "rag-" + baseInstanceName + ":anythingllm"
				configMap := map[string]string{
					"port": req.RAGPort,
				}
				service.UpdateAnythingLLMConfig(req.RAGHost, ragName, configMap, true)
			}

			if req.EnableVectorDB && req.VectorDBHost != "" {
				vdbName := strings.ToLower(req.VectorDBType)
				service.ControlContainer("vdb-"+baseInstanceName+":"+vdbName, "start", req.VectorDBHost)
			}

			if req.EnableParser && req.ParserHost != "" {
				service.ControlContainer("psr-"+baseInstanceName+":mineru-api", "start", req.ParserHost)
			}
		}()
	}

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
	var nodeIPs []string
	utils.ExecuteRead(func() {
		for _, node := range utils.DeploymentNodes {
			nodeIPs = append(nodeIPs, node.NodeIP)
		}
	})

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

	err := utils.ExecuteWrite(func() {
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
	}, true)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save nodes"})
		return
	}

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
