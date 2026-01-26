package service

import (
	"fmt"
	"time"

	"anyadmin-backend/internal/global"
)

func GenerateAndStart(config global.InferenceConfig) (string, string, error) {
	// Mock deployment
	time.Sleep(2 * time.Second)
	containerID := fmt.Sprintf("mock-container-%d", time.Now().Unix())
	path := fmt.Sprintf("/deployments/%s/docker-compose.yml", config.Name)
	return path, containerID, nil
}
