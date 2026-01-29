package api

import (
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HeartbeatRequest struct {
	NodeIP         string                       `json:"node_ip"`
	Hostname       string                       `json:"hostname"`
	Status         string                       `json:"status"`
	CPUUsage       float64                      `json:"cpu_usage"`
	MemoryUsage    float64                      `json:"memory_usage"`
	DockerStatus   string                       `json:"docker_status"`
	DeploymentTime string                       `json:"deployment_time"`
	Services       []global.DockerServiceStatus `json:"services"`
}

// ReceiveHeartbeat handles the POST request from the agent
func ReceiveHeartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service.HandleHeartbeat(req.NodeIP, req.Hostname, req.Status, req.CPUUsage, req.MemoryUsage, req.DockerStatus, req.DeploymentTime, req.Services)

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
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"status": "offline", "message": "Agent not yet seen"})
		return
	}

	// Simple check: if last seen > 30 seconds ago, consider offline
	// (Though the map logic in service just returns the raw struct, let's refine logic here if needed)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}
