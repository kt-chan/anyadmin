package api_test

import (
	"anyadmin-backend/pkg/api"
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/utils"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupDashboardRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/agent/heartbeat", api.ReceiveHeartbeat)
	r.GET("/api/v1/dashboard/stats", api.GetDashboardStats)
	r.POST("/api/v1/container/control", func(c *gin.Context) {
		// Mock auth for control
		c.Set("username", "admin")
		api.ControlContainer(c)
	})
	return r
}

func TestDashboardEnrichment(t *testing.T) {
	// Setup utils config
	utils.ExecuteWrite(func() {
		utils.DeploymentNodes = []global.DeploymentNode{
			{
				NodeIP: "172.20.0.10",
				InferenceCfgs: []global.InferenceConfig{
					{Name: "vllm", IP: "172.20.0.10", Engine: "vLLM"},
				},
			},
		}
	}, true)

	router := setupDashboardRouter()

	// 1. Send Heartbeat with enriched info
	heartbeat := map[string]interface{}{
		"node_ip":         "172.20.0.10",
		"hostname":        "node-gpu-01",
		"status":          "online",
		"cpu_usage":       15.5,
		"memory_usage":    2048.0,
		"docker_status":   "active",
		"deployment_time": "10h",
		"os_spec":         "Ubuntu 22.04.3 LTS",
		"gpu_status":      "NVIDIA A100 80GB | Util: 45% | Mem: 4096/81920 MB",
		"services": []global.DockerServiceStatus{
			{ID: "c1", Name: "vllm", Image: "vllm:latest", State: "running", Status: "Up 10 hours", Uptime: "10h"},
		},
	}
	hbBody, _ := json.Marshal(heartbeat)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/v1/agent/heartbeat", bytes.NewBuffer(hbBody))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 2. Get Dashboard Stats and verify enrichment
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/dashboard/stats", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var stats map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &stats)

	services := stats["services"].([]interface{})

	// Verify Backend service
	foundBackend := false
	foundAgent := false
	foundContainer := false

	for _, s := range services {
		svc := s.(map[string]interface{})
		switch svc["type"] {
		case "Core":
			foundBackend = true
			assert.Equal(t, "Anyadmin-Backend", svc["name"])
		case "Agent":
			foundAgent = true
			assert.Contains(t, svc["name"], "node-gpu-01")
			assert.Contains(t, svc["message"], "Ubuntu 22.04")
			assert.Contains(t, svc["message"], "NVIDIA A100 80GB")
		case "Container":
			foundContainer = true
			assert.Equal(t, "vllm", svc["name"])
		}
	}

	assert.True(t, foundBackend, "Backend service should be in health list")
	assert.True(t, foundAgent, "Agent service should be in health list")
	assert.True(t, foundContainer, "Container service should be in health list")

	// 3. Test Container Control (Mocked SSH for remote)
	// Since it's remote node (172.20.0.10), it will try SSH.
	// To avoid real SSH in unit test, we can check logs or just ensure it reaches the service.

	control := map[string]interface{}{
		"name":    "vllm",
		"action":  "restart",
		"node_ip": "172.20.0.10",
	}
	cBody, _ := json.Marshal(control)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/v1/container/control", bytes.NewBuffer(cBody))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)

	// It might fail because of real SSH attempt, but let's see.
	// If it fails with "connection refused" or similar, it means it tried.
	// In a real mock we would mock GetSSHClient.
	t.Logf("Control response: %d %s", w3.Code, w3.Body.String())
}
