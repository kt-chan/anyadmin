package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/utils"
)

// AgentStatus represents the state of a remote agent
type AgentStatus struct {
	NodeIP         string                       `json:"node_ip"`
	Hostname       string                       `json:"hostname"`
	LastSeen       time.Time                    `json:"last_seen"`
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

var (
	// In-memory store for agent status
	agentStatusMap = make(map[string]AgentStatus)
	statusMutex    sync.RWMutex
)

// HandleHeartbeat updates the status of an agent
func HandleHeartbeat(ip, hostname, status string, cpu float64, cpuCap string, mem float64, memCap string, dockerStatus, deploymentTime, osSpec, gpuStatus string, services []global.DockerServiceStatus) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	agentStatusMap[ip] = AgentStatus{
		NodeIP:         ip,
		Hostname:       hostname,
		LastSeen:       time.Now(),
		Status:         status,
		CPUUsage:       cpu,
		CPUCapacity:    cpuCap,
		MemoryUsage:    mem,
		MemoryCapacity: memCap,
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

// --- Container Control Logic (Merged from container.go) ---

func ControlContainer(containerName string, action string, nodeIP string) error {
	containerNameLower := strings.ToLower(containerName)
	log.Printf("[Agent] ControlContainer: %s %s on %s", action, containerName, nodeIP)

	agentPort := "8082" // Default agent port
	agentURL := fmt.Sprintf("http://%s:%s/container/control", nodeIP, agentPort)

	// Default action
	payload := map[string]string{
		"container_name": containerNameLower,
		"action":         action,
	}
	return sendAgentRequest(agentURL, payload)
}

func sendAgentRequest(url string, payload interface{}) error {
	resp, err := utils.PostJSON(url, payload, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to agent at %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("agent error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode agent response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("agent command failed: %s", result.Message)
	}

	return nil
}

// UpdateVLLMConfig updates the configuration for vLLM on a remote agent
func UpdateVLLMConfig(nodeIP string, config map[string]string, restart bool) error {
	log.Printf("[Agent] UpdateVLLMConfig on %s", nodeIP)

	agentPort := "8082"
	agentURL := fmt.Sprintf("http://%s:%s/config/update", nodeIP, agentPort)

	payload := map[string]interface{}{
		"container_name": "vllm",
		"config":         config,
		"restart":        restart,
	}

	return sendAgentRequest(agentURL, payload)
}
