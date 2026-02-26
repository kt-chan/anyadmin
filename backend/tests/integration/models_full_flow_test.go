package integration

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"anyadmin-backend/pkg/global"
)

func TestModelsFullFlow(t *testing.T) {
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

	// 2. Create Dummy Tar File
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test-model.tar")
	tarFile, _ := os.Create(tarPath)
	tw := tar.NewWriter(tarFile)
	content := []byte(`{"model_type": "llm"}`)
	tw.WriteHeader(&tar.Header{Name: "config.json", Mode: 0600, Size: int64(len(content))})
	tw.Write(content)
	tw.Close()
	tarFile.Close()

	// Calculate Checksum
	f, _ := os.Open(tarPath)
	h := sha256.New()
	io.Copy(h, f)
	tarSum := hex.EncodeToString(h.Sum(nil))
	f.Close()

	// 3. Init Upload for Tar
	initPayload := map[string]interface{}{"filename": "test-model.tar", "total_size": int64(len(content))} // Actually size of tar file
	tarFileInfo, _ := os.Stat(tarPath)
	initPayload["total_size"] = tarFileInfo.Size()
	
	initBody, _ := json.Marshal(initPayload)
	req, _ := http.NewRequest("POST", baseURL+"/models/upload/init", bytes.NewBuffer(initBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	var initRes struct {
		UploadID string `json:"upload_id"`
	}
	json.NewDecoder(resp.Body).Decode(&initRes)
	tarUploadID := initRes.UploadID
	resp.Body.Close()

	// 4. Upload Chunk for Tar
	tarData, _ := os.ReadFile(tarPath)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("upload_id", tarUploadID)
	part, _ := writer.CreateFormFile("chunk", "test-model.tar")
	part.Write(tarData)
	writer.Close()

	req, _ = http.NewRequest("POST", baseURL+"/models/upload/chunk", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to upload tar chunk: %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 5. Init and Upload Checksum (Simplified: just write the string to a file and upload)
	sumPath := filepath.Join(tmpDir, "test-model.sha256")
	os.WriteFile(sumPath, []byte(tarSum), 0644)
	sumData, _ := os.ReadFile(sumPath)
	
	initPayload["filename"] = "test-model.sha256"
	initPayload["total_size"] = int64(len(sumData))
	initBody, _ = json.Marshal(initPayload)
	req, _ = http.NewRequest("POST", baseURL+"/models/upload/init", bytes.NewBuffer(initBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	json.NewDecoder(resp.Body).Decode(&initRes)
	sumUploadID := initRes.UploadID
	resp.Body.Close()

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	writer.WriteField("upload_id", sumUploadID)
	part, _ = writer.CreateFormFile("chunk", "test-model.sha256")
	part.Write(sumData)
	writer.Close()

	req, _ = http.NewRequest("POST", baseURL+"/models/upload/chunk", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, _ = client.Do(req)
	resp.Body.Close()

	// 6. Finalize Upload
	finalizePayload := map[string]string{
		"model_name":         "TestModel",
		"model_type":         "vlm",
		"tar_upload_id":      tarUploadID,
		"checksum_upload_id": sumUploadID,
	}
	finalizeBody, _ := json.Marshal(finalizePayload)
	req, _ = http.NewRequest("POST", baseURL+"/models/finalize", bytes.NewBuffer(finalizeBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to finalize upload: %d, %s", resp.StatusCode, string(bb))
	}
	resp.Body.Close()

	// 7. Verify Model Exists and has type 'vlm'
	req, _ = http.NewRequest("GET", baseURL+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = client.Do(req)
	var modelsRes struct{ Models []global.Model }
	json.NewDecoder(resp.Body).Decode(&modelsRes)
	resp.Body.Close()

	found := false
	for _, m := range modelsRes.Models {
		if m.Name == "TestModel" {
			found = true
			if m.ModelType != "vlm" {
				t.Errorf("Expected model type 'vlm', got %s", m.ModelType)
			}
			break
		}
	}
	if !found {
		t.Errorf("Uploaded model 'TestModel' not found in list")
	}

	// 8. Edit Model (Change type to 'omni')
	editPayload := map[string]string{"name": "TestModel", "model_type": "omni"}
	editBody, _ := json.Marshal(editPayload)
	req, _ = http.NewRequest("PUT", baseURL+"/models/TestModel", bytes.NewBuffer(editBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to update model: %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 9. Verify Edition
	req, _ = http.NewRequest("GET", baseURL+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = client.Do(req)
	json.NewDecoder(resp.Body).Decode(&modelsRes)
	resp.Body.Close()

	found = false
	for _, m := range modelsRes.Models {
		if m.Name == "TestModel" {
			found = true
			if m.ModelType != "omni" {
				t.Errorf("Expected model type 'omni' after edit, got %s", m.ModelType)
			}
			break
		}
	}
	if !found {
		t.Errorf("Model 'TestModel' disappeared after edit")
	}

	// 10. Clean up (Delete Model)
	req, _ = http.NewRequest("DELETE", baseURL+"/models/TestModel", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = client.Do(req)
	resp.Body.Close()
}
