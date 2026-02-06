package service

import (
	"anyadmin-backend/pkg/mockdata"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

type ServiceStatus struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	ModelName string  `json:"model_name,omitempty"`
	Status    string  `json:"status"` // Running, Stopped, Error
	Health    string  `json:"health"` // Healthy, Unhealthy
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
		CPU:     0.5,
		Memory:  m.Alloc,
		PID:     int32(os.Getpid()),
		Message: "Backend service is operational",
	})

	// 2. Agents & Services Status (Base on Deployment Configuration)
	agents := GetAllAgents()
	agentMap := make(map[string]AgentStatus)
	for _, agent := range agents {
		agentMap[agent.NodeIP] = agent
	}

	mockdata.Mu.Lock()
	// Copy nodes to avoid holding lock during processing
	nodes := make([]struct{
		IP string
		Hostname string
		Configs []struct{
			Name string
			Engine string
			ModelName string
			IP string
		}
	}, len(mockdata.DeploymentNodes))
	
	for i, n := range mockdata.DeploymentNodes {
		nodes[i].IP = n.NodeIP
		nodes[i].Hostname = n.Hostname
		// Collect all configs (Inference + RAG) into a generic list for checking
		for _, cfg := range n.InferenceCfgs {
			nodes[i].Configs = append(nodes[i].Configs, struct{Name, Engine, ModelName, IP string}{
				Name: cfg.Name, Engine: cfg.Engine, ModelName: cfg.ModelName, IP: cfg.IP,
			})
		}
		for _, cfg := range n.RagAppCfgs {
			nodes[i].Configs = append(nodes[i].Configs, struct{Name, Engine, ModelName, IP string}{
				Name: cfg.Name, Engine: "RAG App", ModelName: "", IP: cfg.Host,
			})
		}
	}
	mockdata.Mu.Unlock()

	for _, node := range nodes {
		nodeIP := node.IP
		// Handle port in IP if present (legacy safety)
		if strings.Contains(nodeIP, ":") {
			nodeIP = strings.Split(nodeIP, ":")[0]
		}

		agentStatus := "Offline"
		agentHealth := "Unhealthy"
		hostname := node.Hostname
		if hostname == "" { hostname = "Unknown" }
		osSpec := "-"
		gpuStatus := "-"
		uptime := "-"
		var cpu float64
		var mem uint64

		if agent, ok := agentMap[nodeIP]; ok {
			hostname = agent.Hostname
			osSpec = agent.OSSpec
			gpuStatus = agent.GPUStatus
			uptime = agent.DeploymentTime
			cpu = agent.CPUUsage
			mem = uint64(agent.MemoryUsage * 1024 * 1024)

			if time.Since(agent.LastSeen) <= 30*time.Second {
				agentStatus = "Running"
				agentHealth = "Healthy"
			}
		}

		results = append(results, ServiceStatus{
			Name:    fmt.Sprintf("Agent (%s)", hostname),
			Type:    "Agent",
			Status:  agentStatus,
			Health:  agentHealth,
			Uptime:  uptime,
			CPU:     cpu,
			Memory:  mem,
			Message: fmt.Sprintf("Node: %s, OS: %s, GPU: %s", nodeIP, osSpec, gpuStatus),
			NodeIP:  nodeIP,
		})

		// 3. Services from Configuration for this Node
		for _, cfg := range node.Configs {
			svcStatus := "Offline"
			svcHealth := "Unhealthy"
			svcUptime := "-"
			svcMsg := fmt.Sprintf("Node %s has not checked in", nodeIP)
			
			if agent, ok := agentMap[nodeIP]; ok {
				agentOnline := time.Since(agent.LastSeen) <= 30*time.Second
				
				if !agentOnline {
					svcStatus = "Offline"
					svcMsg = fmt.Sprintf("Node %s is offline", nodeIP)
				} else {
					svcStatus = "Stopped"
					svcMsg = fmt.Sprintf("Container not found on node %s", nodeIP)
					
					for _, dockerSvc := range agent.Services {
						lcDockerName := strings.ToLower(dockerSvc.Name)
						lcCfgName := strings.ToLower(cfg.Name)
						lcEngine := strings.ToLower(cfg.Engine)

						// Match if names are similar OR if it's a known engine container
						isMatch := strings.Contains(lcDockerName, lcCfgName) || strings.Contains(lcCfgName, lcDockerName)
						
						// Special case: if engine is vLLM, it might be named just "vllm"
						if !isMatch && (lcEngine == "vllm" || lcEngine == "nvidia") {
							isMatch = strings.Contains(lcDockerName, "vllm")
						}
						// Special case: if engine is MindIE, it might be named "mindie"
						if !isMatch && (lcEngine == "mindie" || lcEngine == "ascend") {
							isMatch = strings.Contains(lcDockerName, "mindie")
						}

						if isMatch {
							if dockerSvc.State == "running" {
								svcStatus = "Running"
								svcHealth = "Healthy"
							}
							svcUptime = dockerSvc.Uptime
							svcMsg = fmt.Sprintf("Image: %s, Node: %s", dockerSvc.Image, agent.NodeIP)
							break
						}
					}
				}
			}

			results = append(results, ServiceStatus{
				Name:      strings.ToLower(cfg.Name),
				Type:      "Container",
				ModelName: cfg.ModelName,
				Status:    svcStatus,
				Health:    svcHealth,
				Uptime:    svcUptime,
				Message:   svcMsg,
				NodeIP:    nodeIP,
			})
		}
	}

	return results
}