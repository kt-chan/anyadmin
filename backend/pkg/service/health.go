package service

import (
	"anyadmin-backend/pkg/mockdata"
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
}

func GetServicesHealth() []ServiceStatus {
	mockdata.Mu.Lock()
	defer mockdata.Mu.Unlock()

	results := make([]ServiceStatus, 0)

	// Core Service (Mock)
	results = append(results, ServiceStatus{
		Name:   "AnythingLLM-admin-core",
		Type:   "Core",
		Status: "Running",
		Health: "Healthy",
		Uptime: "24.5 小时",
		CPU:    1.2,
		Memory: 1024 * 1024 * 50, // 50MB
		PID:    1234,
	})

	// Configured Services
	for _, cfg := range mockdata.InferenceCfgs {
		status := "Running"
		health := "Healthy"
		msg := "Service connected"
		uptime := "12.0 小时"

		// Simulate stopped state for one service randomly or specifically
		if cfg.Name == "milvus-standalone" {
			status = "Stopped"
			health = "Unhealthy"
			msg = "Container exited"
			uptime = "-"
		}

		results = append(results, ServiceStatus{
			Name:    cfg.Name,
			Type:    "Inference",
			Status:  status,
			Health:  health,
			Message: msg,
			Uptime:  uptime,
		})
	}

	return results
}
