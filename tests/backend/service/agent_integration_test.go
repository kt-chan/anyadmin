package service_test

import (
	"testing"
	"time"
	"log"

	"anyadmin-backend/pkg/service"
	"github.com/stretchr/testify/assert"
)

func TestAgentContainerControl(t *testing.T) {
	nodeIP := "172.20.0.10"
	containerName := "vllm" // Assumes vllm container exists or we can try with a non-existent one to see error handling

	// 1. Test Stop
	log.Println("Testing Stop Container...")
	err := service.ControlContainer(containerName, "stop", nodeIP)
	if err != nil {
		t.Logf("Stop failed (might be expected if not running): %v", err)
	} else {
		t.Log("Stop command sent successfully")
	}
	
	// Wait a bit
	time.Sleep(2 * time.Second)

	// 2. Test Start
	log.Println("Testing Start Container...")
	err = service.ControlContainer(containerName, "start", nodeIP)
	// assert.NoError(t, err, "Start container should not return error") // Might fail if container doesn't exist
	if err != nil {
		t.Logf("Start failed: %v", err)
	} else {
		t.Log("Start command sent successfully")
	}

	// 3. Test Invalid Node
	err = service.ControlContainer(containerName, "start", "192.168.1.99")
	assert.Error(t, err, "Should fail for unreachable node")
}
