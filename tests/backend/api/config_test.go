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

func TestInferenceConfig(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/configs/inference", func(c *gin.Context) {
		c.Set("username", "admin")
		api.SaveInferenceConfig(c)
	})
	router.GET("/api/v1/configs/inference", api.GetInferenceConfigs)

	// Initialize utils Data
	utils.DeploymentNodes = []global.DeploymentNode{
		{
			NodeIP: "192.168.1.1",
			InferenceCfgs: []global.InferenceConfig{
				{Name: "vllm-service", IP: "192.168.1.1", Port: "8000", MaxModelLen: 1024, Mode: "balanced"},
			},
		},
		{
			NodeIP: "192.168.1.2",
			InferenceCfgs: []global.InferenceConfig{
				{Name: "vllm-service", IP: "192.168.1.2", Port: "8000", MaxModelLen: 1024, Mode: "balanced"},
				{Name: "other-service", IP: "192.168.1.2", Port: "9000", MaxModelLen: 1024, Mode: "balanced"},
			},
		},
	}

	// Test 1: Global Update (Empty IP)
	t.Run("GlobalUpdate", func(t *testing.T) {
		// Update "vllm-service" globally to MaxModelLen 2048
		config := global.InferenceConfig{
			Name:        "vllm-service",
			IP:          "", // Empty IP implies global
			MaxModelLen: 2048,
		}
		body, _ := json.Marshal(config)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configs/inference", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify changes in utils Data
		utils.ExecuteRead(func() {
			// Check Node 1
			assert.Equal(t, 2048, utils.DeploymentNodes[0].InferenceCfgs[0].MaxModelLen)
			
			// Check Node 2
			assert.Equal(t, 2048, utils.DeploymentNodes[1].InferenceCfgs[0].MaxModelLen)

			// Check Unaffected Service
			assert.Equal(t, 1024, utils.DeploymentNodes[1].InferenceCfgs[1].MaxModelLen)
		})
	})

	// Test 2: Specific Node Update
	t.Run("SpecificNodeUpdate", func(t *testing.T) {
		// Update "vllm-service" on Node 1 only to MaxModelLen 4096
		config := global.InferenceConfig{
			Name:        "vllm-service",
			IP:          "192.168.1.1",
			MaxModelLen: 4096,
		}
		body, _ := json.Marshal(config)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configs/inference", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify changes
		utils.ExecuteRead(func() {
			// Node 1 should be updated
			assert.Equal(t, 4096, utils.DeploymentNodes[0].InferenceCfgs[0].MaxModelLen)
			
			// Node 2 should NOT be updated (should remain 2048 from previous test)
			assert.Equal(t, 2048, utils.DeploymentNodes[1].InferenceCfgs[0].MaxModelLen)
		})
	})
}

func TestRagConfig(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/configs/rag", func(c *gin.Context) {
		c.Set("username", "admin")
		api.SaveRagAppConfig(c)
	})

	// Initialize utils Data
	utils.DeploymentNodes = []global.DeploymentNode{
		{
			NodeIP: "192.168.1.1",
			RagAppCfgs: []global.RagAppConfig{
				{Name: "anythingllm", Host: "192.168.1.1", Port: "3001", StorageDir: "/old/dir"},
			},
		},
		{
			NodeIP: "192.168.1.2",
			RagAppCfgs: []global.RagAppConfig{
				{Name: "anythingllm", Host: "192.168.1.2", Port: "3001", StorageDir: "/old/dir"},
			},
		},
	}

	// Test: Global Update for RAG
	t.Run("GlobalUpdateRag", func(t *testing.T) {
		config := global.RagAppConfig{
			Name:       "anythingllm",
			Host:       "", // Global
			StorageDir: "/new/shared/dir",
		}
		body, _ := json.Marshal(config)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configs/rag", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		utils.ExecuteRead(func() {
			// Verify Node 1
			cfg1 := utils.DeploymentNodes[0].RagAppCfgs[0]
			assert.Equal(t, "/new/shared/dir", cfg1.StorageDir)
			assert.Equal(t, "192.168.1.1", cfg1.Host) // Should be preserved

			// Verify Node 2
			cfg2 := utils.DeploymentNodes[1].RagAppCfgs[0]
			assert.Equal(t, "/new/shared/dir", cfg2.StorageDir)
			assert.Equal(t, "192.168.1.2", cfg2.Host) // Should be preserved
		})
	})
}