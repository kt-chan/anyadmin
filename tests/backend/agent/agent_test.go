package agent_test

import (
	"anyadmin-backend/pkg/agent"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendHeartbeat(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agent/heartbeat" {
			t.Errorf("Expected path /api/v1/agent/heartbeat, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		var req agent.HeartbeatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode body: %v", err)
		}

		if req.NodeIP != "127.0.0.1" {
			t.Errorf("Expected NodeIP 127.0.0.1, got %s", req.NodeIP)
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Call the function
	err := agent.SendHeartbeat(server.URL, "127.0.0.1", "test-host", "24m")
	if err != nil {
		t.Errorf("SendHeartbeat failed: %v", err)
	}
}
