package api_test

import (
	"anyadmin-backend/pkg/api"
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/mockdata"
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

	// Test Save Config with Mode
	t.Run("SaveConfigWithMode", func(t *testing.T) {
		config := global.InferenceConfig{
			Name: "default",
			Mode: "max_concurrency",
		}
		body, _ := json.Marshal(config)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/configs/inference", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp global.InferenceConfig
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "max_concurrency", resp.Mode)

		// Verify it persists in mockdata
		mockdata.Mu.Lock()
		found := false
		for _, cfg := range mockdata.InferenceCfgs {
			if cfg.Name == "default" {
				assert.Equal(t, "max_concurrency", cfg.Mode)
				found = true
				break
			}
		}
		mockdata.Mu.Unlock()
		assert.True(t, found)
	})
}
