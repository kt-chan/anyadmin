package integration_test

import (
	"anyadmin-backend/pkg/api"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentViewContainerControl(t *testing.T) {
	// Setup Router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/container/control", func(c *gin.Context) {
		c.Set("username", "admin")
		api.ControlContainer(c)
	})

	// Target the real remote host
	targetIP := "172.20.0.10"
	targetService := "vllm" // Use a known service

	// 1. Test Restart Action (mirrors the new button in Deployment view)
	t.Run("RestartVLLMFromDeploymentView", func(t *testing.T) {
		controlPayload := map[string]interface{}{
			"name":           targetService,
			"action":         "restart",
			"node_ip":        targetIP,
		}
		body, _ := json.Marshal(controlPayload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/container/control", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		router.ServeHTTP(w, req)

		// We expect 200 OK because the agent is reachable
		// If the agent is not reachable, this might fail (500), which is a valid test failure
		assert.Equal(t, http.StatusOK, w.Code, "Restart request to remote agent should return 200 OK")
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["message"], "Response should contain a message")
		assert.Equal(t, "Action triggered successfully", response["message"], "Message should indicate success")
	})

	// 2. Test Stop Action (mirrors the new button in Deployment view)
	// We'll skip actually stopping it to avoid disruption, or use a dummy name which will return 200 (task triggered) 
	// but might fail in background logging. The agent returns 200 if the command starts.
	t.Run("StopServiceFromDeploymentView", func(t *testing.T) {
		controlPayload := map[string]interface{}{
			"name":           "test_dummy_service",
			"action":         "stop",
			"node_ip":        targetIP,
		}
		body, _ := json.Marshal(controlPayload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/container/control", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Stop request should be accepted")
	})

	// Give a little time for the async operations to clear (optional)
	time.Sleep(1 * time.Second)
}
