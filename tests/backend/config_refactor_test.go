package tests

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

func TestSaveInferenceConfig_Refactor(t *testing.T) {
	// Setup utils Data
	utils.DeploymentNodes = []global.DeploymentNode{
		{
			NodeIP: "172.20.0.10",
			InferenceCfgs: []global.InferenceConfig{
				{
					Name: "old_model",
					IP:   "172.20.0.10",
					Port: "8000",
				},
			},
		},
	}
	
	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("username", "admin")
		c.Next()
	})
	r.POST("/api/config/inference", api.SaveInferenceConfig)

	// Test Case 1: Update existing config
	newConfig := global.InferenceConfig{
		Name:      "old_model", // Match by name
		IP:        "172.20.0.10",
		ModelName: "NewModel-v1",
		Engine:    "vLLM",
	}
	body, _ := json.Marshal(newConfig)
	req, _ := http.NewRequest("POST", "/api/config/inference", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify update in nested structure
	assert.Equal(t, "NewModel-v1", utils.DeploymentNodes[0].InferenceCfgs[0].ModelName)
	assert.Equal(t, "vLLM", utils.DeploymentNodes[0].InferenceCfgs[0].Engine)

	// Test Case 2: Add new config to existing node
	addConfig := global.InferenceConfig{
		Name: "new_service",
		IP:   "172.20.0.10", // Should match existing node
		Port: "8001",
	}
	body, _ = json.Marshal(addConfig)
	req, _ = http.NewRequest("POST", "/api/config/inference", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, utils.DeploymentNodes[0].InferenceCfgs, 2)
	assert.Equal(t, "new_service", utils.DeploymentNodes[0].InferenceCfgs[1].Name)

	// Test Case 3: Add new config to NEW node
	newNodeConfig := global.InferenceConfig{
		Name: "remote_service",
		IP:   "192.168.1.100", // New IP
		Port: "9000",
	}
	body, _ = json.Marshal(newNodeConfig)
	req, _ = http.NewRequest("POST", "/api/config/inference", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should have added a new node
	assert.Len(t, utils.DeploymentNodes, 2)
	assert.Equal(t, "192.168.1.100", utils.DeploymentNodes[1].NodeIP)
	assert.Len(t, utils.DeploymentNodes[1].InferenceCfgs, 1)
	assert.Equal(t, "remote_service", utils.DeploymentNodes[1].InferenceCfgs[0].Name)
}
