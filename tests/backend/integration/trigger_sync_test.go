package integration

import (
	"anyadmin-backend/pkg/service"
	"anyadmin-backend/pkg/utils"
	"testing"
)

func TestTriggerSync(t *testing.T) {
	// Load data from file to populate utils.DeploymentNodes
	utils.LoadFromFile()

	// This will regenerate litellm_config.yaml and push it to all nodes
	err := service.SyncLiteLLMConfig("")
	if err != nil {
		t.Fatalf("Failed to sync LiteLLM: %v", err)
	}
	t.Log("LiteLLM Sync triggered successfully")
}
