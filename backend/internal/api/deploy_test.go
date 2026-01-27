package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFetchVLLMModels(t *testing.T) {
	// 1. Mock the vLLM server
	vllmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"object": "list",
				"data": [
					{"id": "test-model-1", "object": "model", "created": 123456, "owned_by": "vllm"}
				]
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer vllmServer.Close()

	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/deploy/vllm-models", FetchVLLMModels)

	// Extract host and port
	addr := vllmServer.Listener.Addr().String()
	host, port, _ := net.SplitHostPort(addr)
    if host == "::1" { host = "127.0.0.1" }

	// 2. Make Request
	payload := map[string]string{
		"host": host,
		"port": port,
	}
	jsonValue, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/deploy/vllm-models", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// 3. Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// Check if data is present
	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, data)
	model1 := data[0].(map[string]interface{})
	assert.Equal(t, "test-model-1", model1["id"])
}

func TestTestServiceConnection(t *testing.T) {
	// Mock a service (Inference / vLLM)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock health endpoint
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	addr := ts.Listener.Addr().String()
	host, port, _ := net.SplitHostPort(addr)
    if host == "::1" { host = "127.0.0.1" }

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/deploy/test-connection", TestServiceConnection)

	// Case 1: Inference (HTTP /health)
	payload := map[string]string{
		"type": "inference",
		"host": host,
		"port": port,
	}
	jsonValue, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/deploy/test-connection", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "success", resp["status"])

	// Case 2: Generic TCP
	payload2 := map[string]string{
		"type": "vectordb",
		"host": host,
		"port": port,
	}
	jsonValue2, _ := json.Marshal(payload2)
	req2, _ := http.NewRequest("POST", "/deploy/test-connection", bytes.NewBuffer(jsonValue2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	var resp2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	assert.Equal(t, "success", resp2["status"])

	// Case 3: SSH with multiple nodes
	// We use the same mock server host:port for success
	sshHostStr := fmt.Sprintf("%s:%s\n%s", host, port, host) // 1. explicitly with port, 2. implicitly using default port (we will set default port to match our mock)
	
	payload3 := map[string]string{
		"type": "ssh",
		"host": sshHostStr,
		"port": port, // set default port to our mock port so the implicit one works
	}
	jsonValue3, _ := json.Marshal(payload3)
	req3, _ := http.NewRequest("POST", "/deploy/test-connection", bytes.NewBuffer(jsonValue3))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()

	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
	var resp3 map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &resp3)
	// Expect success because our mock server accepts connections
	assert.Equal(t, "success", resp3["status"])
	assert.Contains(t, resp3["message"], "Successfully connected to all 2 nodes")
}