package api

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTestServiceConnection_RAG(t *testing.T) {
	// Mock AnythingLLM App
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	addr := ts.Listener.Addr().String()
	host, port, _ := net.SplitHostPort(addr)
	if host == "::1" {
		host = "127.0.0.1"
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/deploy/test-connection", TestServiceConnection)

	// Case: RAG App (HTTP /)
	payload := map[string]string{
		"type": "rag_app",
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
	assert.Equal(t, "Successfully connected to AnythingLLM service", resp["message"])
}
