package service

import (
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/utils"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

func GetBackendDir() string {
	// 1. Check environment variable
	if envPath := os.Getenv("BACKEND_DIR"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	cwd, _ := os.Getwd()
	checkPaths := []string{
		cwd,
		filepath.Join(cwd, "backend"),
		filepath.Join(cwd, "..", "backend"),
		filepath.Join(cwd, "..", "..", "backend"),
		filepath.Join(cwd, "..", "..", "..", "backend"),
		"/home/anyadmin/app/backend", // Docker default
	}

	for _, p := range checkPaths {
		// First try to find go.mod (development/source mode)
		if _, err := os.Stat(filepath.Join(p, "go.mod")); err == nil {
			data, _ := os.ReadFile(filepath.Join(p, "go.mod"))
			if strings.Contains(string(data), "module anyadmin-backend") {
				return p
			}
		}
		// Then try to find data.json (production/binary mode)
		if _, err := os.Stat(filepath.Join(p, "data.json")); err == nil {
			return p
		}
	}
	return ""
}

// downloadFile downloads a file from a URL to a local path
func downloadFile(url string, destPath string) error {
	log.Printf("[Deploy] Downloading %s to %s...", url, destPath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for download: %w", err)
	}

	resp, err := utils.Get(url, 30*time.Minute) // Long timeout for large files
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}

	return nil
}

func DeployModels(client *ssh.Client) error {
	backendDir := GetBackendDir()
	if backendDir == "" {
		cwd, _ := os.Getwd()
		return fmt.Errorf("could not find backend directory (anyadmin-backend) from %s", cwd)
	}

	log.Println("[Deploy] Copying and extracting model archives...")
	localModelDir := filepath.Join(backendDir, "deployments/tars/models")
	modelFiles, err := os.ReadDir(localModelDir)
	if err != nil {
		return fmt.Errorf("failed to read local model directory: %w", err)
	}

	// Collect tar files from root and one level deep subdirectories
	var tarFiles []string
	for _, file := range modelFiles {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tar") {
			tarFiles = append(tarFiles, filepath.Join(localModelDir, file.Name()))
		} else if file.IsDir() {
			subDir := filepath.Join(localModelDir, file.Name())
			subEntries, err := os.ReadDir(subDir)
			if err == nil {
				for _, sub := range subEntries {
					if !sub.IsDir() && strings.HasSuffix(sub.Name(), ".tar") {
						tarFiles = append(tarFiles, filepath.Join(subDir, sub.Name()))
					}
				}
			}
		}
	}

	for _, localTarPath := range tarFiles {
		tarName := filepath.Base(localTarPath)
		baseName := strings.TrimSuffix(tarName, ".tar")
		remoteModelHomePath := viper.GetString("VLLM_MODEL_PATH")
		if remoteModelHomePath == "" {
			remoteModelHomePath = "/home/anyadmin/data/model"
		}
		remoteExtractDir := remoteModelHomePath + "/" + baseName + "/"

		// Check if model already exists on remote
		checkCmd := fmt.Sprintf("[ -d %s ] && echo \"exists\"", remoteExtractDir)
		output, err := ExecuteCommand(client, checkCmd)
		if err == nil && strings.TrimSpace(output) == "exists" {
			log.Printf("Model %s already exists at %s, skipping copy.", baseName, remoteExtractDir)
			continue
		}

		// Local paths
		// Checksum path (expected in the same directory as the tar file)
		localChecksumPath := filepath.Join(filepath.Dir(localTarPath), baseName+".tar.sha256")

		// Get expected checksum
		expectedChecksum := ""
		if checksumData, err := os.ReadFile(localChecksumPath); err == nil {
			expectedChecksum = strings.Fields(string(checksumData))[0]
			log.Printf("Expected checksum for %s: %s...", tarName, expectedChecksum[:16])
		} else {
			log.Printf("Warning: no checksum file found for %s, skipping verification", tarName)
		}

		// Copy tar file
		log.Printf("Copying %s...", tarName)

		remoteTarPath := remoteModelHomePath + "/" + tarName
		if err := CopyFile(client, localTarPath, remoteTarPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", tarName, err)
		}

		// Set ownership
		if _, err := ExecuteCommand(client, fmt.Sprintf("chown anyadmin:anyadmin %s", remoteTarPath)); err != nil {
			return fmt.Errorf("failed to set ownership for %s: %w", remoteTarPath, err)
		}

		// Verify checksum on remote
		if expectedChecksum != "" {
			log.Printf("Verifying checksum on remote for %s...", tarName)
			output, err := ExecuteCommand(client, fmt.Sprintf("sha256sum %s | cut -d' ' -f1", remoteTarPath))
			if err != nil {
				// Clean up on verification failure
				ExecuteCommand(client, fmt.Sprintf("rm -f %s", remoteTarPath))
				return fmt.Errorf("failed to verify checksum for %s: %w", tarName, err)
			}

			actualChecksum := strings.TrimSpace(output)
			if actualChecksum != expectedChecksum {
				// Clean up corrupted file
				ExecuteCommand(client, fmt.Sprintf("rm -f %s", remoteTarPath))
				return fmt.Errorf("checksum mismatch for %s", tarName)
			}
			log.Printf("Checksum verified for %s", tarName)
		}

		// Extract tar file
		log.Printf("Extracting %s to %s...", tarName, remoteExtractDir)

		// Ensure remoteExtractDir exists
		if _, err := ExecuteCommand(client, fmt.Sprintf("mkdir -p %s", remoteExtractDir)); err != nil {
			return fmt.Errorf("failed to create remote extract directory %s: %w", remoteExtractDir, err)
		}

		// Extract into remoteExtractDir.
		// We use --strip-components=1 if we suspect the tar already has the folder,
		// but if we don't know, it's safer to extract and then check.
		// However, most model tars contain files directly or in a subfolder.
		// Let's try to extract into remoteExtractDir.
		if _, err := ExecuteCommand(client, fmt.Sprintf("tar -xf %s -C %s", remoteTarPath, remoteExtractDir)); err != nil {
			// Clean up on extraction failure
			ExecuteCommand(client, fmt.Sprintf("rm -rf %s", remoteExtractDir))
			ExecuteCommand(client, fmt.Sprintf("rm -f %s", remoteTarPath))
			return fmt.Errorf("failed to extract %s: %w", tarName, err)
		}

		// Handle potential nested directory: if we extracted and found only one directory inside with the same name
		// (e.g. /home/anyadmin/data/model/Qwen/Qwen/...)
		// we might want to move it up, but for now let's just ensure permissions.

		// Set ownership recursively
		if _, err := ExecuteCommand(client, fmt.Sprintf("chown -R anyadmin:anyadmin %s && chmod -R 755 %s", remoteExtractDir, remoteExtractDir)); err != nil {
			return fmt.Errorf("failed to set ownership for extracted files: %w", err)
		}

		// Delete tar file
		if _, err := ExecuteCommand(client, fmt.Sprintf("rm -f %s", remoteTarPath)); err != nil {
			return fmt.Errorf("failed to delete tar file %s: %w", remoteTarPath, err)
		}

		log.Printf("Successfully extracted %s to %s", tarName, remoteExtractDir)
	}

	return nil
}

// RebuildAgent recompiles the agent for Linux AMD64
func RebuildAgent() error {
	backendDir := GetBackendDir()
	if backendDir == "" {
		cwd, _ := os.Getwd()
		return fmt.Errorf("could not find backend directory (anyadmin-backend) from %s", cwd)
	}

	// Check if source code exists (go.mod and cmd/agent/main.go)
	// If missing, we are in a binary-only distribution (like production Docker)
	if _, err := os.Stat(filepath.Join(backendDir, "go.mod")); err != nil {
		log.Printf("[Deploy] go.mod not found in %s, skipping agent rebuild and using existing binary.", backendDir)
		return nil
	}
	if _, err := os.Stat(filepath.Join(backendDir, "cmd", "agent", "main.go")); err != nil {
		log.Printf("[Deploy] Agent source (cmd/agent/main.go) not found in %s, skipping agent rebuild and using existing binary.", backendDir)
		return nil
	}

	// Check if 'go' command exists
	_, err := exec.LookPath("go")
	if err != nil {
		log.Println("[Deploy] 'go' command not found, skipping agent rebuild and using existing binary.")
		// Check if the binary already exists
		localPath := filepath.Join(backendDir, "dist", "anyadmin-agent")
		if _, err := os.Stat(localPath); err != nil {
			return fmt.Errorf("agent binary not found at %s and 'go' is not installed to rebuild it", localPath)
		}
		return nil
	}

	cmd := exec.Command("go", "build", "-o", filepath.Join("dist", "anyadmin-agent"), "./cmd/agent/main.go")
	cmd.Dir = backendDir
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to rebuild agent in %s: %w, output: %s", backendDir, err, output)
	}
	return nil
}

// GenerateAndStart is kept for compatibility
func GenerateAndStart(config global.InferenceConfig) (string, string, error) {
	return "", "", nil
}

// DeployAgent performs the actual deployment steps
func DeployAgent(nodeIP, mgmtHost, mgmtPort, mode string) {
	user := "admin" // Local log user
	// Handle Port in nodeIP
	nodeHost := nodeIP
	nodeSshPort := "22"
	if strings.Contains(nodeIP, ":") {
		parts := strings.Split(nodeIP, ":")
		nodeHost = parts[0]
		nodeSshPort = parts[1]
	}

	RecordLog(user, "Agent Deployment", fmt.Sprintf("Starting deployment to %s (Mode: %s)", nodeIP, mode), "Info")

	// 1. Connect (Assume root or key-based user has rights)
	client, err := GetSSHClient(nodeHost, nodeSshPort)
	if err != nil {
		RecordLog(user, "Agent Deployment", fmt.Sprintf("SSH connection failed to %s: %v", nodeIP, err), "Error")
		return
	}
	defer client.Close()

	// 2. Install Docker (Only for new_deployment)
	if mode == "new_deployment" {
		RecordLog(user, "Agent Deployment", "Installing Docker...", "Info")
		if err := installDocker(client); err != nil {
			RecordLog(user, "Agent Deployment", fmt.Sprintf("Failed to install Docker: %v", err), "Error")
			return
		}

		RecordLog(user, "Agent Deployment", "Installing Node.js...", "Info")
		if err := installNodeNpm(client); err != nil {
			RecordLog(user, "Agent Deployment", fmt.Sprintf("Failed to install Node.js: %v", err), "Error")
			return
		}
	}

	// 3. Create User 'anyadmin'
	RecordLog(user, "Agent Deployment", "Ensuring 'anyadmin' user exists...", "Info")
	if err := ensureUser(client); err != nil {
		RecordLog(user, "Agent Deployment", fmt.Sprintf("Failed to create user: %v", err), "Error")
		return
	}

	// 4. Install Go (Only for new_deployment)
	if mode == "new_deployment" {
		RecordLog(user, "Agent Deployment", "Installing Go...", "Info")
		if err := installGo(client); err != nil {
			RecordLog(user, "Agent Deployment", fmt.Sprintf("Failed to install Go: %v", err), "Error")
			return
		}
	} else {
		RecordLog(user, "Agent Deployment", "Skipping Go installation (Integrate Existing Mode)", "Info")
	}

	// 5. Deploy and Run Agent
	RecordLog(user, "Agent Deployment", "Deploying Agent...", "Info")
	if err := deployAndRunAgent(client, nodeHost, mgmtHost, mgmtPort); err != nil {
		RecordLog(user, "Agent Deployment", fmt.Sprintf("Failed to deploy agent: %v", err), "Error")
		return
	}

	RecordLog(user, "Agent Deployment", fmt.Sprintf("Agent deployment completed for %s", nodeIP), "Success")
}

func ensureUser(client *ssh.Client) error {
	username := "anyadmin"

	// Check if user exists
	_, err := ExecuteCommand(client, fmt.Sprintf("id -u %s", username))
	if err == nil {
		return nil // User exists
	}

	// Create user with sudo and docker access
	// -m: create home directory
	// -s: shell
	// -G: groups (sudo, docker)
	cmd := fmt.Sprintf("useradd -m -s /bin/bash -G sudo,docker %s || (usermod -aG sudo %s && usermod -aG docker %s)", username, username, username)
	if _, err := ExecuteCommand(client, cmd); err != nil {
		return fmt.Errorf("failed to ensure user groups: %w", err)
	}

	// Ensure passwordless sudo for convenience (optional but recommended for agents)
	sudoers := fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL", username)
	cmd = fmt.Sprintf("echo '%s' | tee /etc/sudoers.d/%s", sudoers, username)
	if _, err := ExecuteCommand(client, cmd); err != nil {
		return fmt.Errorf("failed to configure sudoers: %w", err)
	}

	return nil
}

// installNodeNpm installs node and npm if it does not exist in target host.
func installNodeNpm(client *ssh.Client) error {
	log.Println("[Deploy] Installing Node and NPM on Ubuntu 22.04...")
	const nodeVersion = "v22.19.0"

	// 1. Check current version
	output, err := ExecuteCommand(client, "node -v")
	if err == nil {
		version := strings.TrimSpace(output)
		if strings.HasPrefix(version, "v22.") {
			log.Printf("[Deploy] Node.js %s already installed, skipping.", version)
			return nil
		}
		log.Printf("[Deploy] Found old Node.js version %s, upgrading to %s...", version, nodeVersion)
	}

	commands := []string{
		"apt-get update",
		"apt-get remove -y nodejs npm",
		"curl -fsSL https://deb.nodesource.com/setup_22.x | bash -",
		"apt-get install -y nodejs",
		"node -v",
		"npm -v",
	}

	for _, cmd := range commands {
		log.Printf("[Deploy] Running: %s", cmd)
		if _, err := ExecuteCommand(client, cmd); err != nil {
			return fmt.Errorf("failed to execute command '%s': %w", cmd, err)
		}
	}

	log.Printf("[Deploy] Node.js %s and NPM installed successfully.", nodeVersion)
	return nil
}

func installDocker(client *ssh.Client) error {
	log.Println("[Deploy] Installing Docker on Ubuntu 22.04...")

	// 1. Check if docker is already installed
	_, err := ExecuteCommand(client, "docker --version")
	if err == nil {
		log.Println("[Deploy] Docker is already installed, skipping.")
		return nil
	}

	commands := []string{
		"apt-get update",
		"apt-get install -y ca-certificates curl gnupg",
		"install -m 0755 -d /etc/apt/keyrings",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor --yes -o /etc/apt/keyrings/docker.gpg",
		"chmod a+r /etc/apt/keyrings/docker.gpg",
		"echo \"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo \\\"$VERSION_CODENAME\\\") stable\" | tee /etc/apt/sources.list.d/docker.list > /dev/null",
		"apt-get update",
		"apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin",
		"systemctl start docker",
		"systemctl enable docker",
	}

	for _, cmd := range commands {
		log.Printf("[Deploy] Running: %s", cmd)
		if _, err := ExecuteCommand(client, cmd); err != nil {
			return fmt.Errorf("failed to execute command '%s': %w", cmd, err)
		}
	}

	log.Println("[Deploy] Docker installed successfully.")
	return nil
}

func installGo(client *ssh.Client) error {
	const goVersion = "1.25.6"
	backendDir := GetBackendDir()
	if backendDir == "" {
		return fmt.Errorf("could not find backend directory")
	}

	localTarPath := filepath.Join(backendDir, "deployments/tars/os/ubuntu/amd64/jammy/go"+goVersion+".linux-amd64.tar.gz")
	remoteTarPath := "/tmp/go.tar.gz"
	remoteGoHome := "/home/anyadmin/bin/go"
	remoteInstallDir := fmt.Sprintf("%s/go-%s", remoteGoHome, goVersion)
	remoteGoBinPath := fmt.Sprintf("%s/go/bin/go", remoteInstallDir)

	// Check if Go is already installed
	checkCmd := fmt.Sprintf("[ -f %s ] && %s version", remoteGoBinPath, remoteGoBinPath)
	if output, err := ExecuteCommand(client, checkCmd); err == nil && strings.Contains(output, "go version") {
		log.Printf("Go %s already installed at %s, skipping installation.", goVersion, remoteInstallDir)
		return nil
	}

	// Check if local tar exists, if not download it
	if _, err := os.Stat(localTarPath); os.IsNotExist(err) {
		log.Printf("[Deploy] Go tarball missing at %s, attempting download...", localTarPath)
		downloadURL := fmt.Sprintf("https://go.dev/dl/go%s.linux-amd64.tar.gz", goVersion)
		if err := downloadFile(downloadURL, localTarPath); err != nil {
			return fmt.Errorf("failed to download Go tarball: %w", err)
		}
	}

	log.Printf("Installing Go %s to %s...", goVersion, remoteInstallDir)

	// 1. Calculate local hash
	localHash, err := calculateHash(localTarPath)
	if err != nil {
		return fmt.Errorf("failed to calculate local hash: %w", err)
	}

	// 2. Transfer
	if err := CopyFile(client, localTarPath, remoteTarPath); err != nil {
		return fmt.Errorf("failed to transfer Go tarball: %w", err)
	}
	// Ensure cleanup of the tarball on the remote host
	defer ExecuteCommand(client, "rm -f "+remoteTarPath)

	// 3. Verify Remote Hash
	output, err := ExecuteCommand(client, fmt.Sprintf("sha256sum %s", remoteTarPath))
	if err != nil {
		return fmt.Errorf("failed to verify remote hash: %w", err)
	}
	fields := strings.Fields(output)
	if len(fields) == 0 {
		return fmt.Errorf("sha256sum returned empty output")
	}
	if fields[0] != localHash {
		return fmt.Errorf("checksum mismatch: local %s != remote %s", localHash, fields[0])
	}

	// 4. Install
	log.Printf("Extracting Go to %s...", remoteInstallDir)
	extractCmd := fmt.Sprintf("rm -rf %s && mkdir -p %s && tar -C %s -xzf %s", remoteInstallDir, remoteInstallDir, remoteInstallDir, remoteTarPath)
	if _, err := ExecuteCommand(client, extractCmd); err != nil {
		return fmt.Errorf("failed to extract Go: %w", err)
	}

	// 5. Update PATH (system-wide)
	remoteGoBinDir := fmt.Sprintf("%s/go/bin", remoteInstallDir)
	setPathCmd := fmt.Sprintf("echo 'export PATH=$PATH:%s' > /etc/profile.d/go.sh && chmod +x /etc/profile.d/go.sh", remoteGoBinDir)
	if _, err := ExecuteCommand(client, setPathCmd); err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}

	log.Printf("Go %s installed successfully.", goVersion)
	return nil
}

func deployAndRunAgent(client *ssh.Client, nodeIP, mgmtHost, mgmtPort string) error {

	nodePort := "8082"

	log.Printf("Deploying agent to %s:%s with Management Server: %s:%s", nodeIP, nodePort, mgmtHost, mgmtPort)

	log.Println("[Deploy] Starting RebuildAgent...")

	// Rebuild agent before deploying to ensure latest changes

	if err := RebuildAgent(); err != nil {

		log.Printf("Warning: Failed to rebuild agent, using existing binary: %v", err)

	}

	log.Println("[Deploy] RebuildAgent done.")

	backendDir := GetBackendDir()

	localPath := filepath.Join(backendDir, "dist/anyadmin-agent")

	remoteBinDir := viper.GetString("REMOTE_BIN_DIR")
	if remoteBinDir == "" {
		remoteBinDir = "/home/anyadmin/bin"
	}

	remoteBin := remoteBinDir + "/anyadmin-agent"
	remoteData := "/home/anyadmin/data"
	remoteDataAnything := "/home/anyadmin/data/anythingllm"
	remoteConfig := remoteBinDir + "/config.json"
	logDir := "/home/anyadmin/logs"

	// 1. Prepare Directories
	log.Println("[Deploy] Preparing directories...")
	prepCmd := fmt.Sprintf("mkdir -p %s %s && chown -R anyadmin:anyadmin /home/anyadmin && chmod 755 %s", remoteBinDir, logDir, logDir)

	if _, err := ExecuteCommand(client, prepCmd); err != nil {

		return fmt.Errorf("failed to prepare agent directories: %w", err)

	}

	// Stop existing agent before copying to avoid "Text file busy"

	log.Println("[Deploy] Stopping existing agent...")

	ExecuteCommand(client, "pkill -f anyadmin-agent || true")

	time.Sleep(1 * time.Second) // Give it a moment to release file handle

	// 2. Create Config File

	log.Println("[Deploy] Creating config file...")

	config := map[string]interface{}{
		"mgmt_host":       mgmtHost,
		"mgmt_port":       mgmtPort,
		"node_ip":         nodeIP,
		"node_port":       nodePort,
		"deployment_time": time.Now().Format(time.RFC3339),
		"log_file":        "/home/anyadmin/logs/agent.log",
	}

	configBytes, _ := json.Marshal(config)
	configContent := string(configBytes)

	localConfigPath := filepath.Join(os.TempDir(), fmt.Sprintf("config_%s.json", strings.ReplaceAll(nodeIP, ".", "_")))

	if err := os.WriteFile(localConfigPath, []byte(configContent), 0644); err != nil {

		return fmt.Errorf("failed to create local config file: %w", err)

	}

	defer os.Remove(localConfigPath)

	// 3. Copy Binary, Config, and Docker Compose

	log.Println("[Deploy] Copying binaries...")

	if err := CopyFile(client, localPath, remoteBin); err != nil {

		return fmt.Errorf("failed to copy agent binary: %w", err)

	}

	if err := CopyFile(client, localConfigPath, remoteConfig); err != nil {

		return fmt.Errorf("failed to copy agent config: %w", err)

	}

	// Copy Docker Compose file
	log.Println("[Deploy] Copying docker-compose...")
	localComposePath := filepath.Join(backendDir, "deployments/dockers/yaml/template/docker-compose.yaml")
	remoteComposePath := "/home/anyadmin/docker/docker-compose.yaml"
	// Ensure directory exists
	if _, err := ExecuteCommand(client, "mkdir -p /home/anyadmin/docker && chown -R anyadmin:anyadmin /home/anyadmin/docker"); err != nil {
		return fmt.Errorf("failed to create docker directory: %w", err)
	}

	// Materialize docker-compose.yaml based on actual InferenceCfgs
	composeContent, err := os.ReadFile(localComposePath)
	if err == nil {
		contentStr := string(composeContent)

		// 1. Identify the template block for vLLM LLM
		templateStartMarker := "  vllm-llm-template:"

		startIdx := strings.Index(contentStr, templateStartMarker)
		if startIdx != -1 {
			// Find end of block (next service or end of file)
			// A better way is to look for the next top-level key (2 space indent)
			endIdx := strings.Index(contentStr[startIdx+len(templateStartMarker):], "\n  ")
			if endIdx != -1 {
				endIdx += startIdx + len(templateStartMarker)
			} else {
				endIdx = len(contentStr)
			}

			templateBlock := contentStr[startIdx:endIdx]

			// Collect all materialized blocks
			var materializedBlocks []string

			utils.ExecuteRead(func() {
				for _, node := range utils.DeploymentNodes {
					if node.NodeIP == nodeIP {
						for _, cfg := range node.InferenceCfgs {
							if cfg.IsManaged && cfg.ModelType == "llm" && strings.HasPrefix(cfg.Name, "vllm-llm-") {
								// Create a materialized block for this specific model
								materializedBlock := strings.ReplaceAll(templateBlock, "vllm-llm-template", cfg.Name)
								// Also ensure the placeholder inside the block is replaced if it exists
								materializedBlock = strings.ReplaceAll(materializedBlock, "{VLLM_MODEL_NAME}", cfg.ModelName)
								materializedBlock = strings.ReplaceAll(materializedBlock, "${VLLM_MODEL_NAME:-Qwen3-1.7B}", cfg.ModelName)
								materializedBlock = strings.ReplaceAll(materializedBlock, "${VLLM_SERVED_MODEL_NAME:-model}", "model")
								materializedBlock = strings.ReplaceAll(materializedBlock, "${VLLM_MAX_MODEL_LEN:-4096}", fmt.Sprintf("%d", cfg.MaxModelLen))
								materializedBlock = strings.ReplaceAll(materializedBlock, "${VLLM_MAX_NUM_SEQS:-8}", fmt.Sprintf("%d", cfg.MaxNumSeqs))
								materializedBlock = strings.ReplaceAll(materializedBlock, "${VLLM_MAX_NUM_BATCHED_TOKENS:-8192}", fmt.Sprintf("%d", cfg.MaxNumBatchedTokens))
								materializedBlock = strings.ReplaceAll(materializedBlock, "${VLLM_GPU_MEMORY_UTILIZATION:-0.85}", fmt.Sprintf("%.2f", cfg.GpuMemoryUtilization))
								materializedBlock = strings.ReplaceAll(materializedBlock, "${VLLM_LLM_PORT:-8000}", cfg.Port)
								materializedBlock = strings.ReplaceAll(materializedBlock, "${VLLM_NETWORK_ALIAS:-vllm-instance}", cfg.Name)
								materializedBlocks = append(materializedBlocks, materializedBlock)
							}
						}
						break
					}
				}
			})

			if len(materializedBlocks) > 0 {
				// Replace the template with all materialized blocks
				newContent := contentStr[:startIdx] + strings.Join(materializedBlocks, "\n\n") + contentStr[endIdx:]

				// 2. Remove env_file section from litellm service to make it self-contained
				// This matches the env_file block and removes it
				envFileRe := regexp.MustCompile(`(?m)^\s+env_file:\n(\s+-\s+\.env.*\n)+`)
				newContent = envFileRe.ReplaceAllString(newContent, "")

				// Use temporary file for transfer
				tempComposePath := filepath.Join(os.TempDir(), "docker-compose.yaml_deploy")
				if err := os.WriteFile(tempComposePath, []byte(newContent), 0644); err == nil {
					localComposePath = tempComposePath
					defer os.Remove(tempComposePath)

					// Also overwrite the project file for local reference
					projectComposePath := filepath.Join(backendDir, "deployments/dockers/yaml/docker-compose.yaml")
					os.WriteFile(projectComposePath, []byte(newContent), 0644)
				}
			}
		}
	}

	if err := CopyFile(client, localComposePath, remoteComposePath); err != nil {
		log.Printf("Warning: failed to copy docker-compose.yaml: %v", err)
	} else {
		ExecuteCommand(client, fmt.Sprintf("chown anyadmin:anyadmin %s", remoteComposePath))
	}

	// Copy LiteLLM Config
	log.Println("[Deploy] Copying litellm_config.yaml...")
	localLiteLLMPath := filepath.Join(backendDir, "deployments/dockers/yaml/template/litellm_config.yaml")
	remoteLiteLLMPath := "/home/anyadmin/docker/litellm_config.yaml"
	if _, err := os.Stat(localLiteLLMPath); err == nil {
		// Materialize litellm_config.yaml
		liteContent, err := os.ReadFile(localLiteLLMPath)
		if err == nil {
			contentStr := string(liteContent)

			// 1. Identify the template block for vLLM LLM entry
			templateMarker := "  - model_name: {model_name}"
			startIdx := strings.Index(contentStr, templateMarker)

			if startIdx != -1 {
				// Find end of the block (next entry or litellm_settings)
				endIdx := strings.Index(contentStr[startIdx+1:], "\n  -")
				if endIdx != -1 {
					endIdx += startIdx + 1
				} else {
					endIdx = strings.Index(contentStr[startIdx:], "\nlitellm_settings:")
					if endIdx != -1 {
						endIdx += startIdx
					} else {
						endIdx = len(contentStr)
					}
				}

				templateBlock := contentStr[startIdx:endIdx]
				var materializedEntries []string

				utils.ExecuteRead(func() {
					for _, node := range utils.DeploymentNodes {
						if node.NodeIP == nodeIP {
							for _, cfg := range node.InferenceCfgs {
								if cfg.IsManaged && cfg.ModelType == "llm" && strings.HasPrefix(cfg.Name, "vllm-llm-") {
									// Create a materialized block for this specific model
									entry := strings.ReplaceAll(templateBlock, "{model_name}", cfg.ModelName)
									entry = strings.ReplaceAll(entry, "{service_name}", cfg.Name)
									materializedEntries = append(materializedEntries, entry)
								}
							}
							break
						}
					}
				})

				if len(materializedEntries) > 0 {
					// Replace the template with all materialized entries
					newContent := contentStr[:startIdx] + strings.Join(materializedEntries, "\n") + contentStr[endIdx:]

					tempLitePath := filepath.Join(os.TempDir(), "litellm_config.yaml_deploy")
					if err := os.WriteFile(tempLitePath, []byte(newContent), 0644); err == nil {
						localLiteLLMPath = tempLitePath
						defer os.Remove(tempLitePath)
					}
				}
			}
		}

		if err := CopyFile(client, localLiteLLMPath, remoteLiteLLMPath); err != nil {
			log.Printf("Warning: failed to copy litellm_config.yaml: %v", err)
		} else {
			ExecuteCommand(client, fmt.Sprintf("chown anyadmin:anyadmin %s", remoteLiteLLMPath))
		}
	}

	// Copy Environment files
	log.Println("[Deploy] Copying environment files...")
	localEnvDir := filepath.Join(backendDir, "deployments/dockers/yaml/template")
	envFiles, err := os.ReadDir(localEnvDir)
	if err == nil {
		for _, file := range envFiles {
			if !file.IsDir() && strings.HasPrefix(file.Name(), ".env") {
				localEnvPath := filepath.Join(localEnvDir, file.Name())

				// Read template content
				content, err := os.ReadFile(localEnvPath)
				if err != nil {
					log.Printf("Warning: failed to read local env file %s: %v", file.Name(), err)
					continue
				}
				strContent := string(content)

				// Identify which service this env file belongs to
				// .env-vllm-llm, .env-anythingllm, etc.
				serviceSuffix := strings.TrimPrefix(file.Name(), ".env-")

				// Apply customizations from data.json (utils.DeploymentNodes)
				utils.ExecuteRead(func() {
					for _, node := range utils.DeploymentNodes {
						if node.NodeIP == nodeIP {
							// 1. Global node-level variables for the main .env or .env.example
							if file.Name() == ".env" || file.Name() == ".env.example" {
								strContent = updateEnvVar(strContent, "VLLM_MODEL_PATH", viper.GetString("VLLM_MODEL_PATH"))

								// Update ports for all recognized services on this node
								for _, cfg := range node.InferenceCfgs {
									switch cfg.ModelType {
									case "llm":
										strContent = updateEnvVar(strContent, "VLLM_LLM_PORT", cfg.Port)
									case "mineru":
										strContent = updateEnvVar(strContent, "VLLM_MINERU_PORT", cfg.Port)
									case "embedding":
										strContent = updateEnvVar(strContent, "VLLM_EMBEDDING_PORT", cfg.Port)
									}
								}
								for _, cfg := range node.RagAppCfgs {
									if cfg.Name == "anythingllm" {
										strContent = updateEnvVar(strContent, "ANYTHINGLLM_PORT", cfg.Port)
									} else if cfg.Name == "mineru-api" {
										strContent = updateEnvVar(strContent, "MINERU_PORT", cfg.Port)
									}
								}
							}

							// 2. Service-specific variables and placeholders
							// Check InferenceCfgs
							for _, cfg := range node.InferenceCfgs {
								// Match base name (e.g., vllm-llm matches vllm-llm-qwen)
								if strings.HasPrefix(cfg.Name, serviceSuffix) || (strings.Contains(serviceSuffix, "vllm-llm") && cfg.ModelType == "llm") {
									// Update common VLLM variables
									if cfg.Engine == "vLLM" {
										strContent = updateEnvVar(strContent, "VLLM_MODEL_NAME", cfg.ModelName)
										strContent = updateEnvVar(strContent, "VLLM_MAX_MODEL_LEN", fmt.Sprintf("%d", cfg.MaxModelLen))
										strContent = updateEnvVar(strContent, "VLLM_MAX_NUM_SEQS", fmt.Sprintf("%d", cfg.MaxNumSeqs))
										strContent = updateEnvVar(strContent, "VLLM_MAX_NUM_BATCHED_TOKENS", fmt.Sprintf("%d", cfg.MaxNumBatchedTokens))
										strContent = updateEnvVar(strContent, "VLLM_GPU_MEMORY_UTILIZATION", fmt.Sprintf("%.2f", cfg.GpuMemoryUtilization))

										// Handle placeholders
										strContent = strings.ReplaceAll(strContent, "{model_name}", cfg.ModelName)
									}
								}
							}
							// Check RagAppCfgs
							for _, cfg := range node.RagAppCfgs {
								if strings.HasPrefix(cfg.Name, serviceSuffix) {
									strContent = updateEnvVar(strContent, "ANYTHINGLLM_PORT", cfg.Port)
									strContent = updateEnvVar(strContent, "LLM_PROVIDER", cfg.LLMProvider)
									strContent = updateEnvVar(strContent, "GENERIC_OPEN_AI_BASE_PATH", cfg.GenericOpenAIBasePath)
									strContent = updateEnvVar(strContent, "GENERIC_OPEN_AI_MODEL_PREF", cfg.GenericOpenAIModelPref)
									strContent = updateEnvVar(strContent, "GENERIC_OPEN_AI_MODEL_TOKEN_LIMIT", fmt.Sprintf("%d", cfg.GenericOpenAIModelTokenLimit))
									strContent = updateEnvVar(strContent, "GENERIC_OPEN_AI_MAX_TOKENS", fmt.Sprintf("%d", cfg.GenericOpenAIMaxTokens))
									strContent = updateEnvVar(strContent, "VECTOR_DB", cfg.VectorDB)

									decKey := cfg.GenericOpenAIKey
									if dec, err := utils.DecryptPassword(cfg.GenericOpenAIKey); err == nil {
										decKey = dec
									}
									strContent = updateEnvVar(strContent, "GENERIC_OPEN_AI_API_KEY", decKey)

									// Handle placeholders
									strContent = strings.ReplaceAll(strContent, "{BASE_PATH}", cfg.GenericOpenAIBasePath)
									strContent = strings.ReplaceAll(strContent, "{model_name}", cfg.GenericOpenAIModelPref)
									strContent = strings.ReplaceAll(strContent, "{TOKEN_LIMIT}", fmt.Sprintf("%d", cfg.GenericOpenAIModelTokenLimit))
									strContent = strings.ReplaceAll(strContent, "{MAX_TOKENS}", fmt.Sprintf("%d", cfg.GenericOpenAIMaxTokens))
									strContent = strings.ReplaceAll(strContent, "{API_KEY}", decKey)

									if cfg.StorageDir != "" {
										strContent = updateEnvVar(strContent, "STORAGE_DIR", cfg.StorageDir)
									}
								}
							}

							// 3. Global template placeholders (UID, GID, etc)
							strContent = strings.ReplaceAll(strContent, "{UID}", viper.GetString("UID"))
							strContent = strings.ReplaceAll(strContent, "{GID}", viper.GetString("GID"))
							if viper.GetString("UID") != "" {
								strContent = updateEnvVar(strContent, "UID", viper.GetString("UID"))
							}
							if viper.GetString("GID") != "" {
								strContent = updateEnvVar(strContent, "GID", viper.GetString("GID"))
							}
						}
					}
				})

				// Write to a temporary file for transfer
				tempEnvPath := filepath.Join(os.TempDir(), file.Name()+"_deploy")
				if err := os.WriteFile(tempEnvPath, []byte(strContent), 0644); err != nil {
					log.Printf("Warning: failed to create temporary env file: %v", err)
					continue
				}
				defer os.Remove(tempEnvPath)

				// Also overwrite the project file for local reference
				projectEnvPath := filepath.Join(backendDir, "deployments/dockers/yaml", file.Name())
				os.WriteFile(projectEnvPath, []byte(strContent), 0644)

				// Use path.Join for remote Unix paths
				remoteEnvPath := "/home/anyadmin/docker/" + file.Name()
				if err := CopyFile(client, tempEnvPath, remoteEnvPath); err != nil {
					log.Printf("Warning: failed to copy env file %s: %v", file.Name(), err)
				} else {
					ExecuteCommand(client, fmt.Sprintf("chown anyadmin:anyadmin %s", remoteEnvPath))
				}
			}
		}

	} else {
		log.Printf("Warning: failed to read local env directory: %v", err)
	}

	// copy model binary
	log.Println("[Deploy] copying model binary...")
	if err := DeployModels(client); err != nil {
		return fmt.Errorf("failed to deploy models: %w", err)
	}

	// Ensure executable and owned by anyadmin
	log.Println("[Deploy] Setting permissions...")
	ExecuteCommand(client, fmt.Sprintf("chmod +x %s && chown anyadmin:anyadmin %s %s %s", remoteBin, remoteBin, remoteData, remoteConfig))

	log.Println("[Deploy] Override permissions for anythingllm using uid 1000...")

	ExecuteCommand(client, fmt.Sprintf("mkdir -p %s", remoteDataAnything))
	ExecuteCommand(client, fmt.Sprintf("chown 1000:1000 -R %s", remoteDataAnything))
	ExecuteCommand(client, fmt.Sprintf("chmod 777 -R %s", remoteDataAnything))

	// 4. Run Agent

	// The agent now looks for config.json in the same directory by default (or we can specify it)

	// We'll run it from the app directory using absolute paths for everything
	remoteBinAbs := remoteBinDir + "/anyadmin-agent"

	// Wrap in runuser and nohup. Use -c "cd ... && nohup ... > ... < /dev/null &"
	// Redirecting stdin from /dev/null is crucial for nohup via ssh to not hang
	log.Println("[Deploy] Starting agent...")

	fullCmd := fmt.Sprintf("runuser -l anyadmin -c 'cd %s && (nohup %s -config config.json -log /home/anyadmin/logs/agent.log > /home/anyadmin/logs/agent.log 2>&1 < /dev/null &) >/dev/null 2>&1'", remoteBinDir, remoteBinAbs)

	if _, err := ExecuteCommand(client, fullCmd); err != nil {

		return fmt.Errorf("failed to execute start command: %w", err)

	}

	log.Println("[Deploy] Agent start command sent.")

	// 5. Verify process started

	log.Printf("Verifying agent start on %s...", nodeIP)

	time.Sleep(2 * time.Second)
	// Use a more robust check that returns a number
	countStr, err := ExecuteCommand(client, "ps ax | grep anyadmin-agent | grep -v grep | wc -l")
	if err != nil {
		return fmt.Errorf("failed to check agent process: %w", err)
	}

	count, _ := strconv.Atoi(strings.TrimSpace(countStr))
	if count == 0 {
		logTail, _ := ExecuteCommand(client, fmt.Sprintf("tail -n 20 %s/agent.log", logDir))
		return fmt.Errorf("agent failed to start or died immediately on %s. Log tail:\n%s", nodeIP, logTail)
	}
	log.Printf("Agent successfully started on %s (count: %d)", nodeIP, count)

	// Sync agent_config back to utils
	utils.ExecuteWrite(func() {
		for i, node := range utils.DeploymentNodes {
			if node.NodeIP == nodeIP {
				utils.DeploymentNodes[i].AgentConfig = global.AgentConfig{
					MgmtHost:       mgmtHost,
					MgmtPort:       mgmtPort,
					NodeIP:         nodeIP,
					NodePort:       nodePort,
					DeploymentTime: time.Now().Format(time.RFC3339),
					LogFile:        "/home/anyadmin/logs/agent.log",
				}
				break
			}
		}
	}, true)

	// Trigger starting all registered containers
	go StartNodeServices(nodeIP)

	return nil
}

// StartNodeServices triggers all managed containers on a node to start
func StartNodeServices(nodeIP string) {
	var node global.DeploymentNode
	utils.ExecuteRead(func() {
		for _, n := range utils.DeploymentNodes {
			if n.NodeIP == nodeIP {
				node = n
				break
			}
		}
	})

	if node.NodeIP == "" {
		log.Printf("[Deploy] No registered services found for node %s in utils.DeploymentNodes", nodeIP)
		return
	}

	log.Printf("[Deploy] Starting all registered services for node %s", nodeIP)

	// Wait a bit for agent to be fully up and listening
	time.Sleep(5 * time.Second)

	// 1. Sync LiteLLM Config first (ensure proxy has the latest routes)
	if err := SyncLiteLLMConfig(nodeIP); err != nil {
		log.Printf("[Deploy] Warning: failed to sync LiteLLM config on %s: %v", nodeIP, err)
	}

	// 2. Inference Services
	for _, cfg := range node.InferenceCfgs {
		if cfg.IsManaged {
			// Proxy LiteLLM is already handled by SyncLiteLLMConfig
			if cfg.Name == "litellm" {
				continue
			}

			log.Printf("[Deploy] Starting inference service: %s on %s", cfg.Name, nodeIP)
			if cfg.Engine == "vLLM" {
				configMap := map[string]string{
					"model_name": cfg.ModelName,
					"port":       cfg.Port,
				}
				UpdateVLLMConfig(nodeIP, cfg.Name, configMap, true)
			} else {
				ControlContainer(cfg.Name, "start", nodeIP)
			}
		}
	}

	// 3. RAG App Services
	for _, cfg := range node.RagAppCfgs {
		if cfg.IsManaged {
			log.Printf("[Deploy] Starting RAG service: %s on %s", cfg.Name, nodeIP)
			if cfg.Name == "anythingllm" {
				configMap := map[string]string{
					"port": cfg.Port,
				}
				if cfg.GenericOpenAIKey != "" {
					if dec, err := utils.DecryptPassword(cfg.GenericOpenAIKey); err == nil {
						configMap["generic_open_ai_api_key"] = dec
					} else {
						configMap["generic_open_ai_api_key"] = cfg.GenericOpenAIKey
					}
				}
				UpdateAnythingLLMConfig(nodeIP, cfg.Name, configMap, true)
			} else {
				ControlContainer(cfg.Name, "start", nodeIP)
			}
		}
	}
}

func calculateHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// SyncLiteLLMConfig regenerates litellm_config.yaml from data.json and pushes to all nodes
func SyncLiteLLMConfig(nodeIP string) error {
	backendDir := GetBackendDir()
	if backendDir == "" {
		return fmt.Errorf("backend dir not found")
	}

	// 1. Generate new litellm_config.yaml content with ENV placeholders
	var sb strings.Builder
	sb.WriteString("model_list:\n")

	// Collect keys to be pushed to .env
	envKeys := make(map[string]string)

	utils.ExecuteRead(func() {
		for _, node := range utils.DeploymentNodes {
			for _, cfg := range node.InferenceCfgs {
				if cfg.ModelType == "proxy" || cfg.ModelType == "vectordb" {
					continue
				}

				if cfg.IsManaged && cfg.Engine == "vLLM" {
					sb.WriteString(fmt.Sprintf("  - model_name: %s\n", cfg.ModelName))
					sb.WriteString("    litellm_params:\n")
					sb.WriteString(fmt.Sprintf("      model: openai/%s\n", "model")) // Use standard 'model' served-name
					// Use service name (cfg.Name) instead of IP for managed vLLM services to avoid fixed IPs
					sb.WriteString(fmt.Sprintf("      api_base: http://%s:8000/v1\n", cfg.Name))
					sb.WriteString("      api_key: \"not-needed\"\n")
				} else if !cfg.IsManaged && cfg.Engine == "External" {
					// External Cloud (OpenAI / DeepSeek / Zhipu etc)
					sb.WriteString(fmt.Sprintf("  - model_name: %s\n", cfg.ModelName))
					sb.WriteString("    litellm_params:\n")
					sb.WriteString(fmt.Sprintf("      model: %s\n", cfg.ModelName))
					sb.WriteString(fmt.Sprintf("      api_base: %s\n", cfg.BaseURL))
					sb.WriteString("      custom_llm_provider: openai\n")

					if cfg.APIKey != "" {
						// Generic placeholder based on model name
						cleanName := strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(cfg.ModelName), "-", "_"), ".", "_")
						envVarName := cleanName + "_API_KEY"

						sb.WriteString(fmt.Sprintf("      api_key: \"os.environ/%s\"\n", envVarName))

						decKey := cfg.APIKey
						if dec, err := utils.DecryptPassword(cfg.APIKey); err == nil {
							decKey = dec
						}
						envKeys[envVarName] = decKey
					} else {
						sb.WriteString("      api_key: \"not-needed\"\n")
					}
				}
			}
		}
	})

	// 2. Add litellm_settings section
	sb.WriteString("\nlitellm_settings:\n")
	sb.WriteString("  drop_params: true\n")
	sb.WriteString("  set_verbose: true\n")

	// 2. Create temporary file for LiteLLM config
	tempLitePath := filepath.Join(os.TempDir(), "litellm_config.yaml_sync")
	if err := os.WriteFile(tempLitePath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write temporary litellm_config.yaml: %w", err)
	}
	defer os.Remove(tempLitePath)

	// Also overwrite the project file for local reference
	projectLitePath := filepath.Join(backendDir, "deployments/dockers/yaml/litellm_config.yaml")
	os.WriteFile(projectLitePath, []byte(sb.String()), 0644)

	// 2. Push to nodes and trigger agent updates
	pushToNode := func(ip string) {
		host := ip
		port := "22"
		if strings.Contains(ip, ":") {
			parts := strings.Split(ip, ":")
			host = parts[0]
			port = parts[1]
		}

		client, err := GetSSHClient(host, port)
		if err != nil {
			log.Printf("[SyncLiteLLM] SSH failed for %s: %v", ip, err)
			return
		}
		defer client.Close()

		remotePath := "/home/anyadmin/docker/litellm_config.yaml"
		if err := CopyFile(client, tempLitePath, remotePath); err != nil {
			log.Printf("[SyncLiteLLM] Copy failed for %s: %v", ip, err)
			return
		}
		ExecuteCommand(client, fmt.Sprintf("chown anyadmin:anyadmin %s", remotePath))

		// Push .env-litellm if keys exist
		if len(envKeys) > 0 {
			UpdateLiteLLMEnv(host, envKeys)
		}

		// Reload LiteLLM
		log.Printf("[SyncLiteLLM] Reloading LiteLLM on %s", ip)
		ExecuteCommand(client, "cd /home/anyadmin/docker && docker compose -p litellm up -d --force-recreate litellm")
	}

	if nodeIP != "" {
		pushToNode(nodeIP)
	} else {
		utils.ExecuteRead(func() {
			for _, node := range utils.DeploymentNodes {
				pushToNode(node.NodeIP)
			}
		})
	}

	return nil
}

// ControlAgent manages the agent process on a remote node
func ControlAgent(nodeIP, action string) error {
	user := "admin"
	nodeHost := nodeIP
	nodeSshPort := "22"
	if strings.Contains(nodeIP, ":") {
		parts := strings.Split(nodeIP, ":")
		nodeHost = parts[0]
		nodeSshPort = parts[1]
	}

	// Run long operations in background
	go func() {
		RecordLog(user, "Agent Control", fmt.Sprintf("Initiating %s on agent %s", action, nodeIP), "Info")

		client, err := GetSSHClient(nodeHost, nodeSshPort)
		if err != nil {
			msg := fmt.Sprintf("[Agent Control] SSH connection failed to %s: %v", nodeIP, err)
			log.Println(msg)
			RecordLog(user, "Agent Deployment", msg, "Error")
			return
		}
		defer client.Close()

		switch action {
		case "start":
			// Check if already running
			_, err := ExecuteCommand(client, "pgrep -f anyadmin-agent")
			if err == nil {
				RecordLog(user, "Agent Control", "Agent already running on "+nodeIP, "Info")
				return
			}

			var mgmtHost, mgmtPort string
			utils.ExecuteRead(func() {
				mgmtHost = utils.MgmtHost
				mgmtPort = utils.MgmtPort
			})

			if mgmtHost == "" {
				utils.LoadFromFile()
				utils.ExecuteRead(func() {
					mgmtHost = utils.MgmtHost
					mgmtPort = utils.MgmtPort
				})
			}
			if mgmtHost == "" {
				mgmtHost = "172.20.0.1"
				mgmtPort = "8080"
			}

			// Ensure clean start
			ExecuteCommand(client, "pkill -f anyadmin-agent || true")

			if err := deployAndRunAgent(client, nodeHost, mgmtHost, mgmtPort); err != nil {
				msg := fmt.Sprintf("[Agent Control] failed to start agent on %s: %v", nodeIP, err)
				log.Println(msg)
				RecordLog(user, "Agent Control", msg, "Error")
				return
			}
			RecordLog(user, "Agent Control", "Successfully started agent on "+nodeIP, "Success")

		case "stop":
			log.Printf("[Agent Control] Stopping agent on %s", nodeIP)
			ExecuteCommand(client, "pkill -f anyadmin-agent || true")

			// Verification
			time.Sleep(1 * time.Second)
			_, err := ExecuteCommand(client, "pgrep -f anyadmin-agent")
			if err == nil {
				log.Printf("[Agent Control] Agent still running on %s, using SIGKILL", nodeIP)
				ExecuteCommand(client, "pkill -9 -f anyadmin-agent || true")
			}

			RecordLog(user, "Agent Control", "Stopped agent on "+nodeIP, "Success")

		case "restart":
			log.Printf("[Agent Control] Restarting agent on %s", nodeIP)
			ExecuteCommand(client, "pkill -f anyadmin-agent || true")
			time.Sleep(1 * time.Second)

			var mgmtHost, mgmtPort string
			utils.ExecuteRead(func() {
				mgmtHost = utils.MgmtHost
				mgmtPort = utils.MgmtPort
			})

			if mgmtHost == "" {
				utils.LoadFromFile()
				utils.ExecuteRead(func() {
					mgmtHost = utils.MgmtHost
					mgmtPort = utils.MgmtPort
				})
			}
			if mgmtHost == "" {
				mgmtHost = "172.20.0.1"
				mgmtPort = "8080"
			}

			if err := deployAndRunAgent(client, nodeHost, mgmtHost, mgmtPort); err != nil {
				msg := fmt.Sprintf("[Agent Control] failed to restart agent on %s: %v", nodeIP, err)
				log.Println(msg)
				RecordLog(user, "Agent Control", msg, "Error")
				return
			}
			RecordLog(user, "Agent Control", "Successfully restarted agent on "+nodeIP, "Success")

		case "fix-docker":
			RecordLog(user, "Agent Control", "Attempting to fix docker on "+nodeIP, "Info")
			ExecuteCommand(client, "usermod -aG docker anyadmin")
			ExecuteCommand(client, "systemctl restart docker || service docker restart")
			time.Sleep(2 * time.Second)
			ControlAgent(nodeIP, "restart")

		default:
			log.Printf("[Agent Control] unsupported action: %s", action)
		}
	}()

	return nil
}

// updateEnvVar replaces or adds an environment variable in a .env file content string
func updateEnvVar(content, key, value string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?m)^%s=.*$`, key))
	if re.MatchString(content) {
		return re.ReplaceAllString(content, fmt.Sprintf("%s=%s", key, value))
	}
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + fmt.Sprintf("%s=%s\n", key, value)
}
