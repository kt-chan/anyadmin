package utils

import (
	"anyadmin-backend/pkg/utils"
	"testing"
)

func TestCalculateVLLMConfig(t *testing.T) {
	tests := []struct {
		name        string
		modelName   string
		gpuMemory   float64
		mode        string
		utilization float64
	}{
		{
			name:        "Qwen3-1.7B on 8GB Balanced",
			modelName:   "Qwen3-1.7B",
			gpuMemory:   8.0,
			mode:        "balanced",
			utilization: 0.85,
		},
		{
			name:        "Qwen3-1.7B on 8GB MaxToken",
			modelName:   "Qwen3-1.7B",
			gpuMemory:   8.0,
			mode:        "max_token",
			utilization: 0.85,
		},
		{
			name:        "Qwen3-1.7B on 24GB MaxConcurrency",
			modelName:   "Qwen3-1.7B",
			gpuMemory:   24.0,
			mode:        "max_concurrency",
			utilization: 0.90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := utils.CalculateConfigParams{
				ModelNameOrPath: tt.modelName,
				GPUMemoryGB:     tt.gpuMemory,
				Mode:            tt.mode,
				GPUUtilization:  tt.utilization,
			}

			vllmCfg, modelCfg, err := utils.CalculateVLLMConfig(params)
			if err != nil {
				t.Fatalf("Calculation failed: %v", err)
			}

			t.Logf("Mode: %s, MaxModelLen: %d, MaxNumSeqs: %d, MaxNumBatchedTokens: %d, GPUMemoryUtil: %.2f",
				tt.mode, vllmCfg.MaxModelLen, vllmCfg.MaxNumSeqs, vllmCfg.MaxNumBatchedTokens, vllmCfg.GPUMemoryUtil)

			// Validation logic
			if vllmCfg.MaxModelLen <= 0 {
				t.Errorf("MaxModelLen should be positive, got %d", vllmCfg.MaxModelLen)
			}
			if vllmCfg.MaxNumSeqs <= 0 {
				t.Errorf("MaxNumSeqs should be positive, got %d", vllmCfg.MaxNumSeqs)
			}
			if vllmCfg.MaxNumBatchedTokens < vllmCfg.MaxModelLen {
				t.Errorf("MaxNumBatchedTokens (%d) should be at least MaxModelLen (%d)", vllmCfg.MaxNumBatchedTokens, vllmCfg.MaxModelLen)
			}

			// Model Config Validation
			if modelCfg.Name == "" {
				t.Errorf("Model Name should not be empty")
			}
		})
	}
}

func TestMemoryParsing(t *testing.T) {
    // Test some internal helpers if they were exported, 
    // but we can test via main CalculateVLLMConfig if needed.
}
