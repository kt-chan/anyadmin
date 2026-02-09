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
	"anyadmin-backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupConfigRouter() *gin.Engine {
	r := gin.Default()
	// Mock Auth Middleware
	r.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})
	r.POST("/api/v1/services/vllm/config", api.UpdateVLLMConfig)
	return r
}

func TestUpdateVLLMConfigPersistence(t *testing.T) {
	// Setup temporary data file
	tmpFile, err := os.CreateTemp("", "data_test_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Initialize utils with known state
	utils.DataFile = tmpFile.Name()
	utils.DeploymentNodes = []global.DeploymentNode{
		{
			NodeIP: "172.20.0.10",
			InferenceCfgs: []global.InferenceConfig{
				{
					Name:                 "vllm",
					IP:                   "172.20.0.10",
					MaxModelLen:          2048,
					GpuMemoryUtilization: 0.8,
					MaxNumSeqs:           128,
					MaxNumBatchedTokens:  1024,
				},
			},
		},
	}
	// Save initial state
	err = utils.SaveToFile()
	assert.NoError(t, err)

	router := setupConfigRouter()

	// Prepare request
	payload := map[string]interface{}{
		"node_ip": "172.20.0.10",
		"config": map[string]string{
			"VLLM_MAX_MODEL_LEN":          "4096",
			"VLLM_GPU_MEMORY_UTILIZATION": "0.95",
			"VLLM_MAX_NUM_SEQS":           "256",
			"VLLM_MAX_NUM_BATCHED_TOKENS": "2048",
		},
	}
	jsonValue, _ := json.Marshal(payload)

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/services/vllm/config", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Verify utils Update
	assert.Equal(t, 4096, utils.DeploymentNodes[0].InferenceCfgs[0].MaxModelLen)
	assert.Equal(t, 0.95, utils.DeploymentNodes[0].InferenceCfgs[0].GpuMemoryUtilization)
	assert.Equal(t, 256, utils.DeploymentNodes[0].InferenceCfgs[0].MaxNumSeqs)
	assert.Equal(t, 2048, utils.DeploymentNodes[0].InferenceCfgs[0].MaxNumBatchedTokens)

	// Verify File Persistence
	content, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	
	var data utils.DataStore
	err = json.Unmarshal(content, &data)
	assert.NoError(t, err)
	
	found := false
	for _, node := range data.DeploymentNodes {
		for _, cfg := range node.InferenceCfgs {
			if cfg.Name == "vllm" {
				assert.Equal(t, 4096, cfg.MaxModelLen)
				assert.Equal(t, 0.95, cfg.GpuMemoryUtilization)
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Config should be found in file")
}
