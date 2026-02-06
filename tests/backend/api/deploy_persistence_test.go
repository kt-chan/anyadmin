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

func TestDeployServicePersistence(t *testing.T) {
	// Setup temporary data file
	testFile := "test_data.json"
	mockdata.DataFile = testFile
	defer os.Remove(testFile)

	// Initialize data (creates file)
	mockdata.InitData()
	
	// Create Gin router
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/deploy", api.DeployService)

	// Prepare Request
	req := global.DeploymentConfig{
		MgmtHost:      "127.0.0.1",
		MgmtPort:      "8080",
		Mode:          "integrate_existing",
		Platform:      "nvidia",
		InferenceHost: "1.2.3.4",
		InferencePort: "8000",
		ModelName:     "TestModel-v1",
	}
	body, _ := json.Marshal(req)

	// Perform Request
	w := httptest.NewRecorder()
	c, _ := http.NewRequest("POST", "/deploy", bytes.NewBuffer(body))
	r.ServeHTTP(w, c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify InMemory Update
	found := false
	for _, node := range mockdata.DeploymentNodes {
		for _, cfg := range node.InferenceCfgs {
			if cfg.Name == "vllm" {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Config should be in memory")

	// Verify File Persistence
	// Reload from file to a new struct
	fileContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)

	var data mockdata.DataStore
	err = json.Unmarshal(fileContent, &data)
	assert.NoError(t, err)

	foundInFile := false
	for _, node := range data.DeploymentNodes {
		for _, cfg := range node.InferenceCfgs {
			if cfg.Name == "vllm" {
				foundInFile = true
				break
			}
		}
	}
	assert.True(t, foundInFile, "Config should be persisted to file")
}
