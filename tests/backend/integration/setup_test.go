package integration

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// 1. Backup data.json
	// Get absolute path to data.json in the backend directory
	cwd, _ := os.Getwd()
	
	// We need to find the backend root. In tests\backend\integration, it's ../../../backend
	backendRoot := cwd
	if filepath.Base(cwd) == "integration" {
		backendRoot = filepath.Join(cwd, "../../../backend")
	} else if filepath.Base(cwd) == "backend" {
		backendRoot = filepath.Join(cwd, "../../backend")
	}
	
	dataJsonPath := filepath.Join(backendRoot, "data.json")
	dataJsonBackupPath := filepath.Join(backendRoot, "data.json.bak")

	fmt.Printf("--- [PRE-TEST] Backing up %s ---\n", dataJsonPath)
	err := copyFile(dataJsonPath, dataJsonBackupPath)
	if err != nil {
		fmt.Printf("Warning: Failed to backup data.json: %v\n", err)
	}

	// 2. Run Tests
	code := m.Run()

	// 3. Restore data.json
	fmt.Printf("--- [POST-TEST] Restoring %s ---\n", dataJsonPath)
	
	// Try to stop the backend first
	stopBackend(backendRoot)

	err = copyFile(dataJsonBackupPath, dataJsonPath)
	if err != nil {
		fmt.Printf("Error: Failed to restore data.json: %v\n", err)
	} else {
		os.Remove(dataJsonBackupPath)
	}

	// 4. Restart Backend
	startBackend(backendRoot)

	os.Exit(code)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func stopBackend(backendRoot string) {
	fmt.Println("Stopping backend process...")
	// Project root is one level up from backend usually, or same if backend is root.
	// In this repo, it's the parent of backend.
	projectRoot := filepath.Join(backendRoot, "..")
	stopScript := filepath.Join(projectRoot, "scripts/stop-pids.ps1")
	
	cmd := exec.Command("powershell.exe", "-NoProfile", "-File", stopScript)
	cmd.Run()
	time.Sleep(2 * time.Second)
}

func startBackend(backendRoot string) {
	fmt.Println("Restarting backend process...")
	projectRoot := filepath.Join(backendRoot, "..")
	
	// Start backend in separate process
	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", 
		fmt.Sprintf("Start-Process powershell.exe -ArgumentList '-NoProfile', '-Command', 'cd %s; go run cmd/server/main.go' -RedirectStandardOutput '%s' -RedirectStandardError '%s'", 
			backendRoot, 
			filepath.Join(projectRoot, "logs/backend.log"),
			filepath.Join(projectRoot, "logs/backend-error.log")))
	cmd.Run()

	// Wait for it
	waitScript := filepath.Join(projectRoot, "scripts/wait-for-backend.ps1")
	waitCmd := exec.Command("powershell.exe", "-NoProfile", "-File", waitScript)
	waitCmd.Run()
}
