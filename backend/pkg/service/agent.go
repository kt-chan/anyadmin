package service

import (
	"anyadmin-backend/pkg/global"
	"sync"
	"time"
)

// AgentStatus represents the state of a remote agent
type AgentStatus struct {
	NodeIP         string                       `json:"node_ip"`
	Hostname       string                       `json:"hostname"`
	LastSeen       time.Time                    `json:"last_seen"`
	Status         string                       `json:"status"`
	CPUUsage       float64                      `json:"cpu_usage"`
	MemoryUsage    float64                      `json:"memory_usage"`
	DockerStatus   string                       `json:"docker_status"`
	DeploymentTime string                       `json:"deployment_time"`
	OSSpec         string                       `json:"os_spec"`
	GPUStatus      string                       `json:"gpu_status"`
	Services       []global.DockerServiceStatus `json:"services"`
}

var (
	// In-memory store for agent status
	agentStatusMap = make(map[string]AgentStatus)
	statusMutex    sync.RWMutex
)

// HandleHeartbeat updates the status of an agent
func HandleHeartbeat(ip, hostname, status string, cpu, mem float64, dockerStatus, deploymentTime, osSpec, gpuStatus string, services []global.DockerServiceStatus) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	agentStatusMap[ip] = AgentStatus{
		NodeIP:         ip,
		Hostname:       hostname,
		LastSeen:       time.Now(),
		Status:         status,
		CPUUsage:       cpu,
		MemoryUsage:    mem,
		DockerStatus:   dockerStatus,
		DeploymentTime: deploymentTime,
		OSSpec:         osSpec,
		GPUStatus:      gpuStatus,
		Services:       services,
	}
}

// GetAgentStatus retrieves the status of a specific agent
func GetAgentStatus(ip string) (AgentStatus, bool) {
	statusMutex.RLock()
	defer statusMutex.RUnlock()

	status, exists := agentStatusMap[ip]
	return status, exists
}

// GetAllAgents retrieves all known agents
func GetAllAgents() []AgentStatus {
	statusMutex.RLock()
	defer statusMutex.RUnlock()

	agents := make([]AgentStatus, 0, len(agentStatusMap))
	for _, agent := range agentStatusMap {
		agents = append(agents, agent)
	}
	return agents
}
