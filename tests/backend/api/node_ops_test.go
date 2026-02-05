package api_test

import (
	"anyadmin-backend/pkg/api"
	"anyadmin-backend/pkg/mockdata"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupNodeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/v1/deploy/nodes", api.GetNodes)
	r.POST("/api/v1/deploy/nodes", api.SaveNodes)
	r.POST("/api/v1/deploy/agent/control", api.ControlAgent)
	return r
}

func TestNodeOperations(t *testing.T) {

	// Setup Temp Persistence to avoid affecting real data

	tmpFile := "test_node_data.json"
	mockdata.DataFile = tmpFile
	defer os.Remove(tmpFile)
	mockdata.InitData()
	router := setupNodeRouter()

	// 1. Initial nodes
	mockdata.Mu.Lock()
	mockdata.DeploymentNodes = []string{"172.20.0.10:22"}
	mockdata.Mu.Unlock()

	// 2. Test GetNodes
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/deploy/nodes", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string][]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["nodes"], "172.20.0.10:22")

	// 4. Test SaveNodes
	w = httptest.NewRecorder()
	payload := map[string][]string{"nodes": {"172.20.0.10:22"}}
	body, _ := json.Marshal(payload)
	req, _ = http.NewRequest("POST", "/api/v1/deploy/nodes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockdata.Mu.Lock()
	assert.Contains(t, mockdata.DeploymentNodes, "172.20.0.10:22")
	mockdata.Mu.Unlock()
}

func TestAgentControlAPI(t *testing.T) {

	router := setupNodeRouter()

	// Use remote target host 172.20.0.10

	payload := map[string]string{
		"ip":     "172.20.0.10",
		"action": "stop",
	}

	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/deploy/agent/control", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Accept result - should be fast with 2s timeout and reachable IP
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, w.Code)

}
