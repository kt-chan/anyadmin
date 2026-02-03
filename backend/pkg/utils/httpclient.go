package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NewNoProxyClient returns an HTTP client configured to bypass proxies.
func NewNoProxyClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: nil,
		},
		Timeout: timeout,
	}
}

// PostJSON sends a POST request with JSON payload using a no-proxy client.
// It handles marshaling the payload and creating the request.
func PostJSON(url string, payload interface{}, timeout time.Duration) (*http.Response, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	client := NewNoProxyClient(timeout)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Get sends a GET request using a no-proxy client.
func Get(url string, timeout time.Duration) (*http.Response, error) {
	client := NewNoProxyClient(timeout)
	return client.Get(url)
}
