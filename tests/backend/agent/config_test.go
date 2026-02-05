package agent

import (
	"anyadmin-backend/pkg/service"
	"testing"
	"time"
)

func TestUpdateVLLMConfig(t *testing.T) {
	nodeIP := "172.20.0.10"
	config := map[string]string{
		"VLLM_MAX_MODEL_LEN":          "8192",
		"VLLM_GPU_MEMORY_UTILIZATION": "0.85",
	}

    // Wait for agent to start (simple sleep)
    time.Sleep(2 * time.Second)

	err := service.UpdateVLLMConfig(nodeIP, config, true)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}
}
