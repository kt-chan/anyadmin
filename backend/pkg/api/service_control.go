package api

import (
	"net/http"
	"time"

	"anyadmin-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

type ServiceControlRequest struct {
	Name   string `json:"name"`
	NodeIP string `json:"node_ip"`
	Type   string `json:"type"` // Agent, Container, Core
}

// RestartService handles service restart requests
func RestartService(c *gin.Context) {
	var req ServiceControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	
	// Simulate restart delay
	// In a real scenario, this would call the agent or docker client
	service.RecordLog(username.(string), "重启服务", "重启了服务 "+req.Name+" ("+req.Type+")", "Info")

	// Trigger actual logic based on type
	if req.Type == "Container" {
		// Call container service
		if err := service.ControlContainer(req.Name, "restart", req.NodeIP); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restart container: " + err.Error()})
			return
		}
	} else if req.Type == "Agent" {
		// Logic to restart agent (might need SSH or self-restart command)
		// For now, we just log it as successful
		time.Sleep(1 * time.Second) 
	}

	c.JSON(http.StatusOK, gin.H{"status": "restarting", "message": "Service restart initiated"})
}

// StopService handles service stop requests
func StopService(c *gin.Context) {
	var req ServiceControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	service.RecordLog(username.(string), "停止服务", "停止了服务 "+req.Name, "Warning")

	if req.Type == "Container" {
		if err := service.ControlContainer(req.Name, "stop", req.NodeIP); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop container: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "stopped", "message": "Service stop initiated"})
}
