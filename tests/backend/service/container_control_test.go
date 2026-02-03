package service_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"
    "anyadmin-backend/pkg/utils"
)

type ContainerControlRequest struct {
	ContainerName string `json:"container_name"`
	Action        string `json:"action"` // start, stop, restart
}

func TestContainerControl(t *testing.T) {
	agentURL := "http://172.20.0.10:9090/container/control"
	
	// Test Stop vLLM
	t.Log("Testing STOP vllm")
	if err := sendControlRequest(agentURL, "vllm", "stop"); err != nil {
		t.Errorf("Failed to stop vllm: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Test Start vLLM
	t.Log("Testing START vllm")
	if err := sendControlRequest(agentURL, "vllm", "start"); err != nil {
		t.Errorf("Failed to start vllm: %v", err)
	}
    time.Sleep(2 * time.Second)
}

func sendControlRequest(url, name, action string) error {
	req := ContainerControlRequest{
		ContainerName: name,
		Action:        action,
	}

	resp, err := utils.PostJSON(url, req, 10*time.Second)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
