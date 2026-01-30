package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"anyadmin-backend/pkg/api"
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/mockdata"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupDeploymentRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	r.GET("/deploy/ssh-key", api.GetSystemSSHKey)
	r.POST("/deploy/nodes", api.SaveNodes)
	r.GET("/deploy/nodes", api.GetNodes)
	r.POST("/deploy/test-connection", api.TestServiceConnection)
	r.POST("/deploy/generate", api.DeployService)
	
	return r
}

func TestDeploymentFlow(t *testing.T) {
	// Setup Temp Persistence
	tmpFile := "test_flow_data.json"
	mockdata.DataFile = tmpFile
	defer os.Remove(tmpFile)
	mockdata.InitData()

	r := setupDeploymentRouter()

	// 1. Step 1: Download SSH Key
	t.Run("GetSSHKey", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/deploy/ssh-key", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ssh-rsa")
	})

	// 2. Step 1: Save Target Nodes
	t.Run("SaveNodes", func(t *testing.T) {
		nodes := []string{"192.168.1.100:22", "192.168.1.101:22"}
		body, _ := json.Marshal(map[string]interface{}{"nodes": nodes})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/deploy/nodes", bytes.NewBuffer(body))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify persistence
		assert.Equal(t, nodes, mockdata.DeploymentNodes)
	})

	// 3. Step 1: Test SSH Connection
	t.Run("TestConnection_SSH", func(t *testing.T) {
		payload := map[string]string{
			"type": "ssh",
			"host": "127.0.0.1", 
			"port": "8080", // Use backend port to ensure it's open and fast
		}
		body, _ := json.Marshal(payload)
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/deploy/test-connection", bytes.NewBuffer(body))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp, "status") 
	})

	// 4. Step 3 & 4: Test Other Connections
	t.Run("TestConnection_Inference", func(t *testing.T) {
		payload := map[string]string{
			"type": "inference",
			"host": "127.0.0.1",
			"port": "12345",
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/deploy/test-connection", bytes.NewBuffer(body))
		r.ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)
	})

	// 5. Step 5: Generate Deployment
	t.Run("GenerateDeployment", func(t *testing.T) {
		config := global.DeploymentConfig{
			Mode:          "integrate_existing",
			Platform:      "nvidia",
			InferenceHost: "127.0.0.1",
			InferencePort: "8000",
			ModelName:     "Test-Llama-3",
			TargetNodes:   "127.0.0.1:22",
			MgmtHost:      "127.0.0.1",
			MgmtPort:      "8080",
		}
		body, _ := json.Marshal(config)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/deploy/generate", bytes.NewBuffer(body))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		
		artifacts, ok := resp["artifacts"].(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, artifacts, "deploy_script.sh")
		
		found := false
		for _, cfg := range mockdata.InferenceCfgs {
			if cfg.Name == "vllm" {
				found = true
				break
			}
		}
		assert.True(t, found, "Deployment config should be saved")
	})
}
