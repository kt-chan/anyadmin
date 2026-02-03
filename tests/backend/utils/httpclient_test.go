package utils_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"anyadmin-backend/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewNoProxyClient(t *testing.T) {
	client := utils.NewNoProxyClient(10 * time.Second)
	assert.NotNil(t, client)
	assert.Equal(t, 10*time.Second, client.Timeout)
	
	transport, ok := client.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.Nil(t, transport.Proxy, "Proxy should be nil for NoProxyClient")
}

func TestPostJSON(t *testing.T) {
	// Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "test", body["key"])

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	// Test Payload
	payload := map[string]string{"key": "test"}

	// Execute
	resp, err := utils.PostJSON(ts.URL, payload, 5*time.Second)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGet(t *testing.T) {
	// Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	// Execute
	resp, err := utils.Get(ts.URL, 5*time.Second)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
