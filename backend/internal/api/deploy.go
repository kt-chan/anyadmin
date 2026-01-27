package api

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"anyadmin-backend/internal/global"
	"anyadmin-backend/internal/mockdata"
	"anyadmin-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type DeployRequest struct {
	global.InferenceConfig
	DeployMode string `json:"deploy_mode"`
}

func DeployService(c *gin.Context) {
	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var containerID string
	var path string
	var err error

	// 仅当选择“全新部署”时才拉起 Docker 容器
	if req.DeployMode == "new" {
		path, containerID, err = service.GenerateAndStart(req.InferenceConfig)
		if err != nil {
			log.Printf("[部署失败] 模型: %s, 错误: %v", req.Name, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "path": path})
			return
		}
	} else {
		log.Printf("[接入服务] 正在登记现有服务: %s (%s:%s)", req.Name, req.IP, req.Port)
	}

	// 将配置持久化到数据库 (Mock Store)
	config := req.InferenceConfig
	
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

	// 记录审计日志
	action := "服务部署"
	detail := "成功部署并启动了新容器: " + config.Name
	if req.DeployMode == "existing" {
		action = "服务接入"
		detail = "成功接入现有外部服务: " + config.Name + " (" + config.IP + ":" + config.Port + ")"
	} else if containerID != "" {
		detail = "容器已拉起, 名称: " + config.Name + ", ID: " + containerID[:12]
	}
	service.RecordLog(c.GetString("username"), action, detail, "Info")

	c.JSON(http.StatusOK, gin.H{
		"message":      "Success",
		"container_id": containerID,
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
	
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "Connection failed: " + err.Error()})
			return
		}
		conn.Close()
	
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Connection established successfully"})
	}
	