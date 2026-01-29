package agent_test

import (
	"anyadmin-backend/pkg/agent"
	"testing"
)

func TestParseProcStat(t *testing.T) {
	// Example line from /proc/stat
	// cpu  user nice system idle iowait irq softirq steal guest guest_nice
	// cpu  2248 34 2290 226255 106 1 102 0 0 0
	content := "cpu  2248 34 2290 226255 106 1 102 0 0 0\ncpu0 123..."
	
	idle, total, err := agent.ParseProcStat(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected values:
	// idle = 226255
	// total = 2248 + 34 + 2290 + 226255 + 106 + 1 + 102 = 231036
	expectedIdle := uint64(226255)
	expectedTotal := uint64(231036)

	if idle != expectedIdle {
		t.Errorf("expected idle %d, got %d", expectedIdle, idle)
	}
	if total != expectedTotal {
		t.Errorf("expected total %d, got %d", expectedTotal, total)
	}
}

func TestCalculateCPUPercent(t *testing.T) {
	// Case 1: 50% Usage
	usage := agent.CalculateCPUPercent(50, 100, 100, 200)
	if usage != 50.0 {
		t.Errorf("expected 50.0, got %f", usage)
	}

	// Case 2: 100% Usage (No idle increase)
	usage = agent.CalculateCPUPercent(50, 100, 50, 200)
	if usage != 100.0 {
		t.Errorf("expected 100.0, got %f", usage)
	}
	
	// Case 3: 0% Usage (Only idle increase)
	usage = agent.CalculateCPUPercent(50, 100, 150, 200)
	if usage != 0.0 {
		t.Errorf("expected 0.0, got %f", usage)
	}
}

func TestParseMemInfo(t *testing.T) {
	content := `
MemTotal:       16384000 kB
MemFree:         1000000 kB
MemAvailable:    8192000 kB
Buffers:          200000 kB
Cached:          3000000 kB
`
	usage, err := agent.ParseMemInfo(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// usage = (Total - Available) / Total
	// (16384000 - 8192000) / 16384000 = 0.5 -> 50.0%
	if usage != 50.0 {
		t.Errorf("expected 50.0, got %f", usage)
	}
}

func TestParseDockerPsOutput(t *testing.T) {
	output := `1234567890ab|vllm-server|vllm/vllm-openai:latest|Up 2 hours|running|2 hours ago
abcdef123456|anythingllm|mintplexlabs/anythingllm:latest|Up 5 minutes|running|5 minutes ago
7890abcdef12|nginx|nginx:latest|Up 10 days|running|10 days ago`

	services := agent.ParseDockerPsOutput(output)

	// Should only find vllm and anythingllm, not nginx
	if len(services) != 2 {
		t.Errorf("expected 2 services, got %d", len(services))
	}

	foundVllm := false
	foundAnythingLLM := false

	for _, s := range services {
		if s.Name == "vllm-server" {
			foundVllm = true
		}
		if s.Name == "anythingllm" {
			foundAnythingLLM = true
		}
	}

	if !foundVllm {
		t.Errorf("vllm-server not found in results")
	}
	if !foundAnythingLLM {
		t.Errorf("anythingllm not found in results")
	}
}

