package api_test

import (
	"anyadmin-backend/pkg/api"
	"anyadmin-backend/pkg/service"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type HeartbeatRequest struct {
	NodeIP      string  `json:"node_ip"`
	Hostname    string  `json:"hostname"`
	Status      string  `json:"status"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/api/v1/agent/heartbeat", api.ReceiveHeartbeat)
	r.GET("/api/v1/deploy/status", api.CheckAgentStatus)
	return r
}

func TestAgentHeartbeat(t *testing.T) {
	router := setupRouter()

	// 1. Send Heartbeat
	payload := HeartbeatRequest{
		NodeIP:      "10.0.0.1",
		Hostname:    "test-node",
		Status:      "online",
		CPUUsage:    20.0,
		MemoryUsage: 40.0,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/agent/heartbeat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Verify Status in Service
	status, exists := service.GetAgentStatus("10.0.0.1")
	assert.True(t, exists)
	assert.Equal(t, "test-node", status.Hostname)
	assert.Equal(t, 20.0, status.CPUUsage)

	// 3. Check via API
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/deploy/status?ip=10.0.0.1", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	
	assert.Equal(t, "test-node", data["hostname"])
}

func TestCheckAgentStatusNotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/deploy/status?ip=10.0.0.99", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
