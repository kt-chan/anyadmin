package api

import (
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/mockdata"
	"anyadmin-backend/pkg/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HeartbeatRequest struct {
	NodeIP         string                       `json:"node_ip"`
	Hostname       string                       `json:"hostname"`
	Status         string                       `json:"status"`
	CPUUsage       float64                      `json:"cpu_usage"`
	CPUCapacity    string                       `json:"cpu_capacity"`
	MemoryUsage    float64                      `json:"memory_usage"`
	MemoryCapacity string                       `json:"memory_capacity"`
	DockerStatus   string                       `json:"docker_status"`
	DeploymentTime string                       `json:"deployment_time"`
	OSSpec         string                       `json:"os_spec"`
	GPUStatus      string                       `json:"gpu_status"`
	Services       []global.DockerServiceStatus `json:"services"`
}

// ReceiveHeartbeat handles the POST request from the agent
func ReceiveHeartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service.HandleHeartbeat(req.NodeIP, req.Hostname, req.Status, req.CPUUsage, req.CPUCapacity, req.MemoryUsage, req.MemoryCapacity, req.DockerStatus, req.DeploymentTime, req.OSSpec, req.GPUStatus, req.Services)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// CheckAgentStatus allows the frontend to poll for agent status
func CheckAgentStatus(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP address required"})
		return
	}

	status, exists := service.GetAgentStatus(ip)
	
	// Load configured services for this IP from mockdata
	configuredServices := []global.DockerServiceStatus{}
	hostname := ip
	mockdata.Mu.Lock()
	for _, node := range mockdata.DeploymentNodes {
		if node.NodeIP == ip {
			hostname = node.Hostname
			for _, cfg := range node.InferenceCfgs {
				configuredServices = append(configuredServices, global.DockerServiceStatus{
					Name:   cfg.Name,
					Image:  cfg.Engine,
					Status: "Configured (Stopped)",
					State:  "stopped",
				})
			}
			for _, cfg := range node.RagAppCfgs {
				configuredServices = append(configuredServices, global.DockerServiceStatus{
					Name:   cfg.Name,
					Image:  "RAG Application",
					Status: "Configured (Stopped)",
					State:  "stopped",
				})
			}
			break
		}
	}
	mockdata.Mu.Unlock()

	if !exists {
		// Return placeholder status if agent never seen
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"node_ip":       ip,
				"hostname":      hostname,
				"status":        "offline",
				"services":      configuredServices,
				"docker_status": "unknown",
			},
		})
		return
	}

	// If exists, merge configured services if they are not in the reported heartbeat
	for _, cfgSvc := range configuredServices {
		found := false
		for _, hbSvc := range status.Services {
			if hbSvc.Name == cfgSvc.Name {
				found = true
				break
			}
		}
		if !found {
			status.Services = append(status.Services, cfgSvc)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}
