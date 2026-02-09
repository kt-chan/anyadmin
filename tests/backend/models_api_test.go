package backend

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"anyadmin-backend/pkg/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestModelsAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a temporary models directory
	tempDir, err := os.MkdirTemp("", "models-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Override ModelsDir for testing
	api.ModelsDir = tempDir

	r := gin.Default()
	r.GET("/models", api.GetModels)
	r.POST("/models/upload", api.UploadModel)
	r.DELETE("/models/:name", api.DeleteModel)

	// 1. Test List Models (Empty)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/models", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Empty(t, resp["models"])

	// 2. Test Upload Model
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "test-model")
	
	part, _ := writer.CreateFormFile("files", "config.json")
	part.Write([]byte(`{"test": true}`))
	writer.Close()

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/models/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Upload failed with status %d: %s", w.Code, w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Test List Models (1 model)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/models", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &resp)
	models := resp["models"].([]interface{})
	assert.Len(t, models, 1)
	assert.Equal(t, "test-model", models[0].(map[string]interface{})["name"])

	// 4. Test Delete Model
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/models/test-model", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Test List Models (Empty again)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/models", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Empty(t, resp["models"])
}
