package api

import (
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
	foundInConfig := false
	utils.ExecuteRead(func() {
		for _, node := range utils.DeploymentNodes {
			if node.NodeIP == ip {
				hostname = node.Hostname
				foundInConfig = true
				for _, cfg := range node.InferenceCfgs {
					configuredServices = append(configuredServices, global.DockerServiceStatus{
						Name:      cfg.Name,
						Image:     cfg.Engine,
						Status:    "Configured (Stopped)",
						State:     "stopped",
						IsManaged: true,
					})
				}
				for _, cfg := range node.RagAppCfgs {
					configuredServices = append(configuredServices, global.DockerServiceStatus{
						Name:      cfg.Name,
						Image:     "RAG Application",
						Status:    "Configured (Stopped)",
						State:     "stopped",
						IsManaged: true,
					})
				}
				break
			}
		}
	})

	if !exists && !foundInConfig {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent status not found for IP: " + ip})
		return
	}

	if !exists {
		// Return placeholder status if agent never seen but exists in config
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
		lcCfgName := strings.ToLower(cfgSvc.Name)
		// Handle projectName:serviceName format by converting to common docker styles
		dockerBase := strings.ReplaceAll(lcCfgName, ":", "-")
		dockerAlt := strings.ReplaceAll(dockerBase, ".", "_")

		for i, hbSvc := range status.Services {
			lcHbName := strings.ToLower(hbSvc.Name)

			// Match if names are equal OR if heartbeat name contains/is contained by sanitized config name
			if lcHbName == lcCfgName ||
				strings.Contains(lcHbName, dockerBase) ||
				strings.Contains(lcHbName, dockerAlt) ||
				strings.Contains(dockerBase, lcHbName) {
				found = true
				// Ensure it's marked as managed if it matches config
				status.Services[i].IsManaged = true
				// Use the prettier config name for display if it's a match
				status.Services[i].Name = cfgSvc.Name
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
