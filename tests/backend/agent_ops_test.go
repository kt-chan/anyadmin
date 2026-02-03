package backend

import (
	"anyadmin-backend/pkg/service"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgentDeploymentAndControl(t *testing.T) {
	nodeIP := "172.20.0.10"
	mgmtHost := "172.20.0.1" 
	mgmtPort := "8080"

	// 2. Deploy Agent (simulates "send to target host" and "start")
	// "integrate" mode bypasses Go installation, assuming we just update the agent.
	// NOTE: DeployAgent now handles stopping existing agent first.
	t.Log("Deploying agent...")
	
	// Use a channel to detect timeout of the synchronous DeployAgent call if it hangs
	done := make(chan bool)
	go func() {
		service.DeployAgent(nodeIP, mgmtHost, mgmtPort, "integrate")
		done <- true
	}()

	select {
	case <-done:
		t.Log("Agent deployment returned.")
	case <-time.After(60 * time.Second):
		t.Fatal("DeployAgent timed out after 60s")
	}
	
	// Give it some time to start up
	t.Log("Waiting for agent startup (5s)...")
	time.Sleep(5 * time.Second)

	// 3. Check Health
	healthURL := fmt.Sprintf("http://%s:9090/health", nodeIP)
	t.Logf("Checking health at %s", healthURL)
	
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get(healthURL)
	if err != nil {
		t.Fatalf("Failed to contact agent health endpoint: %v", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Agent should be healthy")

	// 4. Test Control Container (Restart anythingllm)
	t.Log("Testing ControlContainer restart for anythingllm...")
	err = service.ControlContainer("anythingllm", "restart", nodeIP)
	if err != nil {
		t.Errorf("ControlContainer failed: %v", err)
	} else {
		t.Log("ControlContainer restart success")
	}

	// 5. Test Control Container (Restart vllm)
	t.Log("Testing ControlContainer restart for vllm...")
	err = service.ControlContainer("vllm", "restart", nodeIP)
	if err != nil {
		t.Errorf("ControlContainer vllm restart failed: %v", err)
	} else {
		t.Log("ControlContainer vllm restart success")
	}
}
