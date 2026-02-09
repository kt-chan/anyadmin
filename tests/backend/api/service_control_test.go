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

func TestRemoteServiceControlAndHealth(t *testing.T) {
	// Setup utils config
	utils.Mu.Lock()
	utils.DeploymentNodes = []global.DeploymentNode{
		{
			NodeIP: "172.20.0.10",
			InferenceCfgs: []global.InferenceConfig{
				{Name: "vllm", IP: "172.20.0.10", Engine: "vLLM"},
				{Name: "anythingllm", IP: "172.20.0.10", Engine: "RAG App"},
			},
		},
	}
	utils.Mu.Unlock()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/agent/heartbeat", api.ReceiveHeartbeat)
	router.GET("/api/v1/dashboard/stats", api.GetDashboardStats)
	router.POST("/api/v1/container/control", func(c *gin.Context) {
		c.Set("username", "admin")
		api.ControlContainer(c)
	})

	// 1. Simulate Heartbeat from the known remote target 172.20.0.10
	heartbeat := map[string]interface{}{
		"node_ip":         "172.20.0.10",
		"hostname":        "DESKTOP-OSLI7Q7",
		"status":          "online",
		"cpu_usage":       5.0,
		"memory_usage":    1024.0,
		"docker_status":   "active",
		"deployment_time": "34m",
		"os_spec":         "Ubuntu 22.04.3 LTS",
		"gpu_status":      "NVIDIA RTX 4090 | Util: 10% | Mem: 1024/24576 MB",
		"services": []global.DockerServiceStatus{
			{ID: "4ac9ee55d204", Name: "vllm", Image: "vllm/vllm-openai:latest", State: "running", Status: "Up 34 minutes (healthy)", Uptime: "34m"},
			{ID: "22e815dfb667", Name: "anythingllm", Image: "mintplexlabs/anythingllm:1.8.5", State: "running", Status: "Up 36 minutes (healthy)", Uptime: "36m"},
		},
	}
	hbBody, _ := json.Marshal(heartbeat)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/v1/agent/heartbeat", bytes.NewBuffer(hbBody))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 2. Check Dashboard health list
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/dashboard/stats", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var stats map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &stats)
	services := stats["services"].([]interface{})

	foundVllm := false
	foundAnythingLLM := false
	for _, s := range services {
		svc := s.(map[string]interface{})
		if svc["name"] == "vllm" && svc["node_ip"] == "172.20.0.10" {
			foundVllm = true
			assert.Equal(t, "Running", svc["status"])
			assert.Equal(t, "Healthy", svc["health"])
		}
		if svc["name"] == "anythingllm" && svc["node_ip"] == "172.20.0.10" {
			foundAnythingLLM = true
			assert.Equal(t, "Running", svc["status"])
		}
	}
	assert.True(t, foundVllm, "vllm service should be found in stats")
	assert.True(t, foundAnythingLLM, "anythingllm service should be found in stats")

	// 3. Test Control Operation on remote vllm
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

	assert.Equal(t, http.StatusOK, w3.Code, "Remote restart on vllm should succeed")

	// 4. Test New Service Restart Endpoint
	router.POST("/api/v1/services/restart", func(c *gin.Context) {
		c.Set("username", "admin")
		api.RestartService(c)
	})

	restartReq := map[string]interface{}{
		"name":    "vllm",
		"type":    "Container",
		"node_ip": "172.20.0.10",
	}
	rBody, _ := json.Marshal(restartReq)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("POST", "/api/v1/services/restart", bytes.NewBuffer(rBody))
	req4.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code, "Service restart should succeed")
}

func TestAnythingLLMControl(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/container/control", func(c *gin.Context) {
		c.Set("username", "admin")
		api.ControlContainer(c)
	})

	// Test STOP with mixed case name
	t.Run("StopAnythingLLM", func(t *testing.T) {
		control := map[string]interface{}{
			"name":    "AnythingLLM", // Mixed case to test standardization
			"action":  "stop",
			"node_ip": "172.20.0.10",
		}
		cBody, _ := json.Marshal(control)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/container/control", bytes.NewBuffer(cBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Stop operation on AnythingLLM should succeed")
	})

	// Test START with mixed case name
	t.Run("StartAnythingLLM", func(t *testing.T) {
		control := map[string]interface{}{
			"name":    "AnythingLLM",
			"action":  "start",
			"node_ip": "172.20.0.10",
		}
		cBody, _ := json.Marshal(control)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/container/control", bytes.NewBuffer(cBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Start operation on AnythingLLM should succeed")
	})

	// Test DELETE (Remove) service
	t.Run("RemoveAnythingLLM", func(t *testing.T) {
		control := map[string]interface{}{
			"name":    "AnythingLLM",
			"action":  "rm", // Corresponds to docker rm
			"node_ip": "172.20.0.10",
		}
		cBody, _ := json.Marshal(control)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/container/control", bytes.NewBuffer(cBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Remove (rm) operation on AnythingLLM should succeed")
	})
}
