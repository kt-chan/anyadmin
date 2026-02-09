package agent_test

import (
	"anyadmin-backend/pkg/agent"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleUpdateConfig(t *testing.T) {
	// Setup temporary directory
	tempDir, err := ioutil.TempDir("", "agent_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override DockerDir
	originalDockerDir := agent.DockerDir
	agent.DockerDir = tempDir + string(os.PathSeparator)
	defer func() { agent.DockerDir = originalDockerDir }()

	// Create dummy .env-vllm
	envPath := filepath.Join(tempDir, ".env-vllm")
	initialContent := "MODEL_NAME=OldModel\nEXISTING_VAR=123\n"
	if err := ioutil.WriteFile(envPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial env file: %v", err)
	}

	// Prepare request
	payload := map[string]interface{}{
		"container_name": "vllm",
		"config": map[string]string{
			"model_name":    "NewModel",
			"max_model_len": "8192",
		},
		"restart": false,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/config/update", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Call handler
	agent.HandleUpdateConfig(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify file content
	content, err := ioutil.ReadFile(envPath)
	if err != nil {
		t.Fatalf("Failed to read env file: %v", err)
	}
	strContent := string(content)

	if !strings.Contains(strContent, "VLLM_MODEL_NAME=NewModel") {
		t.Errorf("Expected VLLM_MODEL_NAME=NewModel, got:\n%s", strContent)
	}
	if !strings.Contains(strContent, "VLLM_MAX_MODEL_LEN=8192") {
		t.Errorf("Expected VLLM_MAX_MODEL_LEN=8192, got:\n%s", strContent)
	}
	if !strings.Contains(strContent, "EXISTING_VAR=123") {
		t.Errorf("Expected EXISTING_VAR=123 to remain, got:\n%s", strContent)
	}
}