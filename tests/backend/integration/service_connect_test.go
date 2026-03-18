package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"anyadmin-backend/pkg/global"
)

func TestServiceConnectFlow(t *testing.T) {
	baseURL := "http://localhost:8080/api/v1"
	client := &http.Client{Timeout: 5 * time.Second}

	// 1. Login
	loginPayload := map[string]string{"username": "admin", "password": "password"}
	loginBody, _ := json.Marshal(loginPayload)
	resp, _ := http.Post(baseURL+"/login", "application/json", bytes.NewBuffer(loginBody))
	var loginRes struct{ Token string }
	json.NewDecoder(resp.Body).Decode(&loginRes)
	token := loginRes.Token
	resp.Body.Close()

	// 2. Test "Connect Existing" Mode (Integration)
	connectExistingPayload := global.DeploymentConfig{
		MgmtHost:      "172.20.0.1",
		MgmtPort:      "8080",
		TargetNodes:   "172.20.0.11",
		Mode:          "integrate_existing",
		Platform:      "nvidia",
		ModelType:     "llm",
		InferenceHost: "172.20.0.11",
		InferencePort: "8001",
		ModelName:     "ExistingModel",
	}
	existBody, _ := json.Marshal(connectExistingPayload)
	req, _ := http.NewRequest("POST", baseURL+"/deploy/generate", bytes.NewBuffer(existBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to integrate existing service: %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 3. Verify Existing Service is in data.json via API
	req, _ = http.NewRequest("GET", baseURL+"/configs/services", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = client.Do(req)
	var svcConfig struct {
		Nodes []global.DeploymentNode `json:"nodes"`
	}
	json.NewDecoder(resp.Body).Decode(&svcConfig)
	resp.Body.Close()

	foundExisting := false
	for _, node := range svcConfig.Nodes {
		if node.NodeIP == "172.20.0.11" {
			for _, cfg := range node.InferenceCfgs {
				if cfg.Port == "8001" && cfg.ModelName == "ExistingModel" {
					foundExisting = true
					if cfg.IsManaged {
						t.Errorf("Expected integrated existing service to be NOT managed")
					}
					break
				}
			}
		}
	}
	if !foundExisting {
		t.Errorf("Integrated existing service not found in config")
	}

	// 4. Test "New Deployment" Mode (Managed)
	managedPayload := global.DeploymentConfig{
		MgmtHost:      "172.20.0.1",
		MgmtPort:      "8080",
		TargetNodes:   "172.20.0.12",
		Mode:          "new_deployment",
		Platform:      "nvidia",
		ModelType:     "vlm",
		InferenceHost: "172.20.0.12",
		InferencePort: "8002",
		ModelName:     "ManagedModel",
	}
	managedBody, _ := json.Marshal(managedPayload)
	req, _ = http.NewRequest("POST", baseURL+"/deploy/generate", bytes.NewBuffer(managedBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to deploy managed service: %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 5. Verify Managed Service is in data.json via API
	req, _ = http.NewRequest("GET", baseURL+"/configs/services", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = client.Do(req)
	json.NewDecoder(resp.Body).Decode(&svcConfig)
	resp.Body.Close()

	foundManaged := false
	for _, node := range svcConfig.Nodes {
		if node.NodeIP == "172.20.0.12" {
			for _, cfg := range node.InferenceCfgs {
				if cfg.Port == "8002" && cfg.ModelName == "ManagedModel" {
					foundManaged = true
					if !cfg.IsManaged {
						t.Errorf("Expected new deployment service to be managed")
					}
					break
				}
			}
		}
	}
	if !foundManaged {
		t.Errorf("Managed service not found in config")
	}
}
