package integration

import (
	"anyadmin-backend/pkg/global"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestModelClassificationFlow(t *testing.T) {
	baseURL := "http://localhost:8080/api/v1"

	// 1. Login to get token
	loginPayload := map[string]string{
		"username": "admin",
		"password": "password",
	}
	loginBody, _ := json.Marshal(loginPayload)
	resp, err := http.Post(baseURL+"/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	var loginRes struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginRes)
	token := loginRes.Token

	client := &http.Client{Timeout: 5 * time.Second}

	// 2. Get Model Types
	req, _ := http.NewRequest("GET", baseURL+"/model-types", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to get model types: %v", err)
	}
	var types []string
	json.NewDecoder(resp.Body).Decode(&types)
	foundLLM := false
	for _, tt := range types {
		if tt == "llm" {
			foundLLM = true
			break
		}
	}
	if !foundLLM {
		t.Errorf("Expected 'llm' in model types, got %v", types)
	}

	// 3. Add Custom Model Type
	newType := "test-type-" + fmt.Sprintf("%d", time.Now().Unix())
	addPayload := map[string]string{"type": newType}
	addBody, _ := json.Marshal(addPayload)
	req, _ = http.NewRequest("POST", baseURL+"/model-types", bytes.NewBuffer(addBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to add model type: %v", err)
	}

	// 4. Verify Custom Type exists
	req, _ = http.NewRequest("GET", baseURL+"/model-types", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	json.NewDecoder(resp.Body).Decode(&types)
	foundCustom := false
	for _, tt := range types {
		if tt == newType {
			foundCustom = true
			break
		}
	}
	if !foundCustom {
		t.Errorf("Expected custom type %s in model types, got %v", newType, types)
	}

	// 5. Test Deployment with Model Type
	deployConfig := global.DeploymentConfig{
		MgmtHost:      "172.20.0.1",
		MgmtPort:      "8080",
		TargetNodes:   "172.25.208.100:22",
		Mode:          "new_deployment",
		Platform:      "nvidia",
		ModelType:     "llm",
		InferenceHost: "172.25.208.100",
		InferencePort: "8000",
		ModelName:     "llama-3.2-1B-Instruct",
	}
	deployBody, _ := json.Marshal(deployConfig)
	req, _ = http.NewRequest("POST", baseURL+"/deploy/generate", bytes.NewBuffer(deployBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to trigger deployment: %v", err)
	}

	// 6. Verify Service Config has ModelType
	req, _ = http.NewRequest("GET", baseURL+"/configs/services", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	var svcConfig struct {
		Nodes []global.DeploymentNode `json:"nodes"`
	}
	json.NewDecoder(resp.Body).Decode(&svcConfig)

	foundSvc := false
	for _, node := range svcConfig.Nodes {
		if node.NodeIP == "172.25.208.100" {
			for _, cfg := range node.InferenceCfgs {
				if cfg.ModelType == "llm" {
					foundSvc = true
					// Check Name standardization: engine-type
					expectedName := "vllm-llm"
					if cfg.Name != expectedName {
						t.Errorf("Expected service name %s, got %s", expectedName, cfg.Name)
					}
					break
				}
			}
		}
	}
	if !foundSvc {
		t.Errorf("Expected service with model_type 'llm' not found in config")
	}
}
