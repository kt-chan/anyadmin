package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPDFConvertTool(t *testing.T) {
	// Paths relative to tests/backend/utils directory
	pdfPath := filepath.Join("..", "..", "..", "docs", "知识库管理界面需求.pdf")
	outputDir := filepath.Join("..", "..", "..", "docs")
	serverURL := "http://172.20.0.10:8010"
	toolPath := filepath.Join("..", "..", "..", "backend", "tools", "pdf_covert.go")

	// Ensure PDF exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Fatalf("Test PDF not found at %s", pdfPath)
	}

	// Run the tool using 'go run'
	cmd := exec.Command("go", "run", toolPath, pdfPath, outputDir, serverURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Note: This test will attempt to connect to the real serverURL.
	// In a CI environment, we might want to mock this, but the objective 
	// asks to run it against the specified serverURL.
	err := cmd.Run()
	if err != nil {
		t.Logf("Note: Conversion might fail if server %s is unreachable", serverURL)
		t.Errorf("Tool execution failed: %v", err)
	}
}
