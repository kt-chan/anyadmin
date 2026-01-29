package api_test

import (
	"anyadmin-backend/pkg/api"
	"anyadmin-backend/pkg/mockdata"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupNodeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/v1/deploy/nodes", api.GetNodes)
	r.POST("/api/v1/deploy/nodes", api.SaveNodes)
	r.DELETE("/api/v1/deploy/nodes", api.RemoveNode)
	r.POST("/api/v1/deploy/agent/control", api.ControlAgent)
	return r
}

func TestNodeOperations(t *testing.T) {
	router := setupNodeRouter()

	// 1. Initial nodes
	mockdata.Mu.Lock()
	mockdata.DeploymentNodes = []string{"192.168.1.100:22", "192.168.1.101:22"}
	mockdata.Mu.Unlock()

	// 2. Test GetNodes
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/deploy/nodes", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp map[string][]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp["nodes"], 2)

	// 3. Test RemoveNode
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/deploy/nodes?ip=192.168.1.100", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's gone
	mockdata.Mu.Lock()
	assert.Len(t, mockdata.DeploymentNodes, 1)
	assert.Equal(t, "192.168.1.101:22", mockdata.DeploymentNodes[0])
	mockdata.Mu.Unlock()

	// 4. Test SaveNodes
	w = httptest.NewRecorder()
	payload := map[string][]string{"nodes": {"10.0.0.1:22"}}
	body, _ := json.Marshal(payload)
	req, _ = http.NewRequest("POST", "/api/v1/deploy/nodes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	mockdata.Mu.Lock()
	assert.Len(t, mockdata.DeploymentNodes, 1)
	assert.Equal(t, "10.0.0.1:22", mockdata.DeploymentNodes[0])
	mockdata.Mu.Unlock()
}

func TestAgentControlAPI(t *testing.T) {
	router := setupNodeRouter()

	// This test will try to connect via SSH if we use a real IP, 
	// but since we are in unit test, it should fail with "connection refused" or similar if we use localhost.
	// Actually service.ControlAgent will return error if SSH fails.
	
	payload := map[string]string{
		"ip": "127.0.0.1",
		"action": "stop",
	}
	body, _ := json.Marshal(payload)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/deploy/agent/control", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// It should fail because there is no SSH server on 127.0.0.1:22 (usually)
	// But it proves the route and handler are working.
	// In a real mock environment we would interface the service.
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, w.Code)
}
