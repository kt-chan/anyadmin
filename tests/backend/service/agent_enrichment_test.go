package service_test

import (
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/mockdata"
	"testing"
	"os"
)

func TestAgentEnrichment(t *testing.T) {
	// Initialize mockdata
	mockdata.InitData()
	mockdata.MgmtHost = "172.20.0.1"
	mockdata.MgmtPort = "8080"
	mockdata.SaveToFile()

	targetIP := "172.20.0.10"
	
	t.Run("RebuildAgent", func(t *testing.T) {
		err := service.RebuildAgent()
		if err != nil {
			t.Fatalf("RebuildAgent failed: %v", err)
		}
		// The binary should be in backend/dist/agent_linux relative to root
		// From this test's perspective (tests/backend/service), it's ../../../backend/dist/agent_linux
		if _, err := os.Stat("../../../backend/dist/agent_linux"); os.IsNotExist(err) {
			t.Fatal("agent_linux binary not found after rebuild at ../../../backend/dist/agent_linux")
		}
	})

	t.Run("ControlAgent_Start", func(t *testing.T) {
		// This will test SSH connection, config generation, transfer, and execution
		err := service.ControlAgent(targetIP, "start")
		if err != nil {
			t.Errorf("ControlAgent Start failed: %v", err)
		}
	})

	t.Run("ControlAgent_Stop", func(t *testing.T) {
		err := service.ControlAgent(targetIP, "stop")
		if err != nil {
			t.Errorf("ControlAgent Stop failed: %v", err)
		}
	})
    
    t.Run("ControlAgent_Restart", func(t *testing.T) {
		err := service.ControlAgent(targetIP, "restart")
		if err != nil {
			t.Errorf("ControlAgent Restart failed: %v", err)
		}
	})
}
