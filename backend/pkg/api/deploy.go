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
	"anyadmin-backend/pkg/mockdata"
	"anyadmin-backend/pkg/service"

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
		Name:      req.ModelName,
		IP:        req.InferenceHost,
		Port:      req.InferencePort,
		ModelPath: req.ModelName, // Assuming model name implies path/id
	}

	// Map Engine based on Platform
	switch req.Platform {
	case "nvidia":
		inferenceConfig.Engine = "vLLM"
	case "ascend":
		inferenceConfig.Engine = "MindIE"
	default:
		inferenceConfig.Engine = "Unknown"
	}

	// Always deploy agent to target nodes to ensure control/telemetry
	if req.TargetNodes != "" {
		nodes := strings.Split(req.TargetNodes, "\n")
		for _, node := range nodes {
			node = strings.TrimSpace(node)
			if node == "" {
				continue
			}
			// Async deployment
			go service.DeployAgent(node, req.MgmtHost, req.MgmtPort, req.Mode)
		}
	}

	// Mode specific logging or additional actions
	if req.Mode == "new_deployment" {
		log.Printf("[全新部署] 正在初始化节点并拉起服务: %s", req.ModelName)
	} else {
		log.Printf("[接入服务] 正在登记现有服务: %s (%s:%s)", req.ModelName, req.InferenceHost, req.InferencePort)
	}

	// 将配置持久化到数据库 (Mock Store)
	config := inferenceConfig
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	mockdata.Mu.Lock()
	found := false
	for i, cfg := range mockdata.InferenceCfgs {
		if cfg.Name == config.Name {
			mockdata.InferenceCfgs[i] = config
			found = true
			break
		}
	}
	if !found {
		mockdata.InferenceCfgs = append(mockdata.InferenceCfgs, config)
	}
	mockdata.Mu.Unlock()
	
	// Persist to file
	mockdata.SaveToFile()

	// 记录审计日志
	action := "服务部署"
	detail := "发起部署任务: " + config.Name
	if req.Mode == "integrate_existing" {
		action = "服务接入"
		detail = "接入现有服务: " + config.Name + " (" + config.IP + ":" + config.Port + ")"
	}
	service.RecordLog(c.GetString("username"), action, detail, "Info")

	c.JSON(http.StatusOK, gin.H{
		"message":      "Deployment Started",
		"container_id": "pending",
		"artifacts": gin.H{ // Mock artifacts for frontend display
			"deploy_script.sh": "#!/bin/bash\n# Deployment Script\n# Deployment is now handled automatically by the backend via SSH.\n# You can check the server logs for progress.",
			"config.yaml":      fmt.Sprintf("model: %s\nengine: %s\nhost: %s\nport: %s", config.Name, config.Engine, config.IP, config.Port),
		},
	})
}

// GetNodes returns the list of target nodes
func GetNodes(c *gin.Context) {
	mockdata.Mu.Lock()
	defer mockdata.Mu.Unlock()
	c.JSON(http.StatusOK, gin.H{"nodes": mockdata.DeploymentNodes})
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

	mockdata.Mu.Lock()
	mockdata.DeploymentNodes = req.Nodes
	mockdata.Mu.Unlock()
	
	mockdata.SaveToFile()

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

	url := fmt.Sprintf("http://%s:%s/v1/models", req.Host, req.Port)

	// Use custom client to bypass proxy for internal service calls
	tr := &http.Transport{
		Proxy: nil,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to vLLM service: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Proxy the JSON response directly
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

	timeout := 2 * time.Second

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
		// Bypass proxy for internal checks
		tr := &http.Transport{
			Proxy: nil,
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   timeout,
		}
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Successfully connected to vLLM service"})
			return
		}
		// Fallback to TCP if HTTP fails or for generic check
	}

	// For AnythingLLM (RAG App), check HTTP root
	if req.Type == "rag_app" {
		url := fmt.Sprintf("http://%s", address)
		tr := &http.Transport{Proxy: nil}
		client := &http.Client{Transport: tr, Timeout: timeout}
		resp, err := client.Get(url)
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
