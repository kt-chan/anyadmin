package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MineruResponse mirrors the JSON structure returned by the API
type MineruResponse struct {
	Results     map[string]FileData `json:"results"`
	ModelOutput interface{}         `json:"model_output"`
	ContentList interface{}         `json:"content_list"`
}

type FileData struct {
	MdContent string            `json:"md_content"`
	Md        string            `json:"md"`
	Images    map[string]string `json:"images"`
}

func main() {
	pdfPath := "docs/知识库管理界面需求.pdf"
	outputDir := "docs"
	serverURL := "http://172.20.0.10:8010"

	if len(os.Args) > 1 {
		pdfPath = os.Args[1]
	}
	if len(os.Args) > 2 {
		outputDir = os.Args[2]
	}
	if len(os.Args) > 3 {
		serverURL = os.Args[3]
	}

	// Ensure serverURL has a schema
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}

	backendVllmServerURL := "http://host.docker.internal:8000"
	err := convertPDFToMarkdown(pdfPath, outputDir, serverURL, backendVllmServerURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Conversion completed successfully.")
}

func convertPDFToMarkdown(pdfPath, outputDir, apiServerURL, backendVllmServerURL string) error {
	// 1. Validate File
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return fmt.Errorf("PDF file not found: %s", pdfPath)
	}

	pdfBase := filepath.Base(pdfPath)
	pdfStem := strings.TrimSuffix(pdfBase, filepath.Ext(pdfBase))
	finalOutputDir := filepath.Join(outputDir, pdfStem)

	if err := os.MkdirAll(finalOutputDir, 0755); err != nil {
		return err
	}

	fmt.Printf("Converting: %s\nOutput to: %s\n", pdfPath, finalOutputDir)

	// 2. Prepare Multipart Form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add PDF file
	file, err := os.Open(pdfPath)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("files", pdfBase)
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, file); err != nil {
		return err
	}

	// Add Form Fields
	params := map[string]string{
		"return_middle_json":  "false",
		"return_model_output": "false",
		"return_md":           "true",
		"return_images":       "true",
		"end_page_id":         "99999",
		"parse_method":        "auto",
		"start_page_id":       "0",
		"lang_list":           "ch,en",
		"output_dir":          outputDir,
		"server_url":          backendVllmServerURL,
		"return_content_list": "false",
		"backend":             "vlm-http-client",
		"table_enable":        "true",
		"formula_enable":      "true",
		"response_format_zip": "false",
	}
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	writer.Close()

	// 3. Execute Request
	req, err := http.NewRequest("POST", apiServerURL+"/file_parse", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("accept", "application/json")

	client := &http.Client{
		Timeout: 300 * time.Second,
	}
	fmt.Printf("Sending request to: %s/file_parse\n", apiServerURL)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// 4. Parse Response
	fmt.Println("Parsing response...")
	var result MineruResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// 5. Handle Markdown & Images
	if result.Results == nil {
		return fmt.Errorf("API returned empty results")
	}

	if fileData, ok := result.Results[pdfStem]; ok {
		mdContent := fileData.MdContent
		if mdContent == "" {
			mdContent = fileData.Md
		}

		if mdContent == "" {
			return fmt.Errorf("no markdown content found in results for %s", pdfStem)
		}

		mdFilePath := filepath.Join(finalOutputDir, pdfStem+".md")
		err = os.WriteFile(mdFilePath, []byte(mdContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to save markdown: %w", err)
		}
		fmt.Printf("Markdown successfully saved to: %s\n", mdFilePath)

		// Save Images
		if len(fileData.Images) > 0 {
			imgDir := filepath.Join(finalOutputDir, "images")
			if err := os.MkdirAll(imgDir, 0755); err != nil {
				return fmt.Errorf("failed to create image directory: %w", err)
			}

			count := 0
			for name, b64Data := range fileData.Images {
				parts := strings.Split(b64Data, ",")
				rawStr := parts[len(parts)-1]

				imgBytes, err := base64.StdEncoding.DecodeString(rawStr)
				if err != nil {
					fmt.Printf("Warning: failed to decode image %s: %v\n", name, err)
					continue
				}

				err = os.WriteFile(filepath.Join(imgDir, name), imgBytes, 0644)
				if err != nil {
					fmt.Printf("Warning: failed to save image %s: %v\n", name, err)
					continue
				}
				count++
			}
			fmt.Printf("Saved %d images to: %s\n", count, imgDir)
		}
	} else {
		var keys []string
		for k := range result.Results {
			keys = append(keys, k)
		}
		return fmt.Errorf("key '%s' not found in results. Available keys: %v", pdfStem, keys)
	}

	return nil
}
