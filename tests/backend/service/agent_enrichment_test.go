package service_test

import (
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"
	"testing"
	"os"
)

func TestAgentEnrichment(t *testing.T) {
	// Initialize utils
	utils.InitData()
	utils.MgmtHost = "172.20.0.1"
	utils.MgmtPort = "8080"
	utils.SaveToFile()

	targetIP := "172.20.0.10"
	
	t.Run("RebuildAgent", func(t *testing.T) {
		err := service.RebuildAgent()
		if err != nil {
			t.Fatalf("RebuildAgent failed: %v", err)
		}
		// The binary should be in backend/dist/anyadmin-agent relative to root
		// From this test's perspective (tests/backend/service), it's ../../../backend/dist/anyadmin-agent
		if _, err := os.Stat("../../../backend/dist/anyadmin-agent"); os.IsNotExist(err) {
			t.Fatal("anyadmin-agent binary not found after rebuild at ../../../backend/dist/anyadmin-agent")
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
