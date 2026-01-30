package service

import (
	"fmt"
	"log"
	"time"
)

func ControlContainer(containerName string, action string, nodeIP string) error {
	log.Printf("[Container] Action: %s on container: %s (Node: %s)", action, containerName, nodeIP)
	
	if nodeIP == "" || nodeIP == "localhost" || nodeIP == "127.0.0.1" {
		// Mock local action
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	// Remote action via SSH
	client, err := GetSSHClient(nodeIP, "22")
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %w", nodeIP, err)
	}
	defer client.Close()

	cmd := fmt.Sprintf("docker %s %s", action, containerName)
	output, err := ExecuteCommand(client, cmd)
	if err != nil {
		return fmt.Errorf("remote command failed: %w, output: %s", err, output)
	}

	return nil
}
