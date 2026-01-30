package service

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

type ServiceStatus struct {
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Status  string  `json:"status"` // Running, Stopped, Error
	Health  string  `json:"health"` // Healthy, Unhealthy
	Uptime  string  `json:"uptime"`
	CPU     float64 `json:"cpu"`    // 进程 CPU 占用
	Memory  uint64  `json:"memory"` // 进程内存占用 (Bytes)
	PID     int32   `json:"pid"`
	Message string  `json:"message"`
	NodeIP  string  `json:"node_ip,omitempty"`
}

var startTime = time.Now()

func GetServicesHealth() []ServiceStatus {
	results := make([]ServiceStatus, 0)

	// 1. Backend Core Service Status
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	uptime := time.Since(startTime).Round(time.Second).String()

	results = append(results, ServiceStatus{
		Name:    "Anyadmin-Backend",
		Type:    "Core",
		Status:  "Running",
		Health:  "Healthy",
		Uptime:  uptime,
		CPU:     0.5, // Mock CPU as we don't have gopsutil
		Memory:  m.Alloc,
		PID:     int32(os.Getpid()),
		Message: "Backend service is operational",
	})

	// 2. Services from Agents (Heartbeats)
	agents := GetAllAgents()
	for _, agent := range agents {
		// Add agent itself as a service
		agentStatus := "Running"
		agentHealth := "Healthy"
		if time.Since(agent.LastSeen) > 30*time.Second {
			agentStatus = "Offline"
			agentHealth = "Unhealthy"
		}

		results = append(results, ServiceStatus{
			Name:    fmt.Sprintf("Agent (%s)", agent.Hostname),
			Type:    "Agent",
			Status:  agentStatus,
			Health:  agentHealth,
			Uptime:  agent.DeploymentTime,
			CPU:     agent.CPUUsage,
			Memory:  uint64(agent.MemoryUsage * 1024 * 1024), // Assuming mem usage in MB
			Message: fmt.Sprintf("Node: %s, OS: %s, GPU: %s", agent.NodeIP, agent.OSSpec, agent.GPUStatus),
			NodeIP:  agent.NodeIP,
		})

		// Add docker services managed by this agent
		for _, svc := range agent.Services {
			svcStatus := "Stopped"
			svcHealth := "Unhealthy"
			if svc.State == "running" {
				svcStatus = "Running"
				svcHealth = "Healthy"
			}

			results = append(results, ServiceStatus{
				Name:    svc.Name,
				Type:    "Container",
				Status:  svcStatus,
				Health:  svcHealth,
				Uptime:  svc.Uptime,
				Message: fmt.Sprintf("Image: %s, Node: %s", svc.Image, agent.NodeIP),
				NodeIP:  agent.NodeIP,
			})
		}
	}

	return results
}
