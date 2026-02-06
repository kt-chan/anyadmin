package service_test

import (
	"anyadmin-backend/pkg/service"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
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
	healthURL := fmt.Sprintf("http://%s:8082/health", nodeIP)
	t.Logf("Checking health at %s", healthURL)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	var lastErr error
	// Poll for health (up to 60s)
	for i := 0; i < 12; i++ {
		// Try local GET first
		resp, err := client.Get(healthURL)
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				t.Log("Agent is healthy (via external HTTP)")
				lastErr = nil
				break
			}
			resp.Body.Close()
			lastErr = fmt.Errorf("status code %d", resp.StatusCode)
		} else {
			// Fallback: Check via SSH if external access is blocked
			sshClient, sshErr := service.GetSSHClient(nodeIP, "22")
			if sshErr == nil {
				// Check using curl on localhost
				out, cmdErr := service.ExecuteCommand(sshClient, "curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8082/health")
				sshClient.Close()
				if cmdErr == nil && strings.TrimSpace(out) == "200" {
					t.Log("Agent is healthy (via internal curl)")
					lastErr = nil
					break
				}
			}
			lastErr = err
		}
		t.Logf("Health check failed (attempt %d/12): %v. Retrying...", i+1, lastErr)
		time.Sleep(5 * time.Second)
	}

	if lastErr != nil {
		t.Fatalf("Agent failed to become healthy: %v", lastErr)
	}
	// assert.Equal is removed as we checked it in the loop

	// 4. Test Control Container (Restart anythingllm)
	t.Log("Testing ControlContainer restart for anythingllm...")
	err := service.ControlContainer("anythingllm", "restart", nodeIP)
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
