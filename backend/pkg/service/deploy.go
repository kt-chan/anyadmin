package service

import (
	"anyadmin-backend/pkg/global"
	"anyadmin-backend/pkg/mockdata"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func getBackendDir() string {
	cwd, _ := os.Getwd()
	checkPaths := []string{
		filepath.Join(cwd, "backend"),
		filepath.Join(cwd, "..", "backend"),
		filepath.Join(cwd, "..", "..", "backend"),
		filepath.Join(cwd, "..", "..", "..", "backend"),
		cwd,
	}

	for _, p := range checkPaths {
		if _, err := os.Stat(filepath.Join(p, "go.mod")); err == nil {
			data, _ := os.ReadFile(filepath.Join(p, "go.mod"))
			if strings.Contains(string(data), "module anyadmin-backend") {
				return p
			}
		}
	}
	return ""
}

func DeployModels(client *ssh.Client) error {
	backendDir := getBackendDir()
	if backendDir == "" {
		cwd, _ := os.Getwd()
		return fmt.Errorf("could not find backend directory (anyadmin-backend) from %s", cwd)
	}

	log.Println("[Deploy] Copying and extracting model archives...")
	localModelDir := filepath.Join(backendDir, "deployments/models")
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
		remoteModelHomePath := "/home/anyadmin/data/model"
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

		// Extract
		if _, err := ExecuteCommand(client, fmt.Sprintf("tar -xf %s -C %s", remoteTarPath, remoteModelHomePath)); err != nil {
			// Clean up on extraction failure
			ExecuteCommand(client, fmt.Sprintf("rm -rf %s", remoteExtractDir))
			ExecuteCommand(client, fmt.Sprintf("rm -f %s", remoteTarPath))
			return fmt.Errorf("failed to extract %s: %w", tarName, err)
		}

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
	backendDir := getBackendDir()
	if backendDir == "" {
		cwd, _ := os.Getwd()
		return fmt.Errorf("could not find backend directory (anyadmin-backend) from %s", cwd)
	}

	cmd := exec.Command("go", "build", "-o", "./dist/anyadmin-agent", "./cmd/agent/main.go")
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
	nodePort := "22"
	if strings.Contains(nodeIP, ":") {
		parts := strings.Split(nodeIP, ":")
		nodeHost = parts[0]
		nodePort = parts[1]
	}

	RecordLog(user, "Agent Deployment", fmt.Sprintf("Starting deployment to %s (Mode: %s)", nodeIP, mode), "Info")

	// 1. Connect (Assume root or key-based user has rights)
	client, err := GetSSHClient(nodeHost, nodePort)
	if err != nil {
		RecordLog(user, "Agent Deployment", fmt.Sprintf("SSH connection failed to %s: %v", nodeIP, err), "Error")
		return
	}
	defer client.Close()

	// 2. Create User 'anyadmin'
	RecordLog(user, "Agent Deployment", "Ensuring 'anyadmin' user exists...", "Info")
	if err := ensureUser(client); err != nil {
		RecordLog(user, "Agent Deployment", fmt.Sprintf("Failed to create user: %v", err), "Error")
		return
	}

	// 3. Install Go (Only for new_deployment)
	if mode == "new_deployment" {
		RecordLog(user, "Agent Deployment", "Installing Go...", "Info")
		if err := installGo(client); err != nil {
			RecordLog(user, "Agent Deployment", fmt.Sprintf("Failed to install Go: %v", err), "Error")
			return
		}
	} else {
		RecordLog(user, "Agent Deployment", "Skipping Go installation (Integrate Existing Mode)", "Info")
	}

	// 4. Deploy and Run Agent
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

func installGo(client *ssh.Client) error {
	const goVersion = "1.25.6"
	backendDir := getBackendDir()
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

	log.Printf("Deploying agent to %s with Management Server: %s:%s", nodeIP, mgmtHost, mgmtPort)

	log.Println("[Deploy] Starting RebuildAgent...")

	// Rebuild agent before deploying to ensure latest changes

	if err := RebuildAgent(); err != nil {

		log.Printf("Warning: Failed to rebuild agent, using existing binary: %v", err)

	}

	log.Println("[Deploy] RebuildAgent done.")

	backendDir := getBackendDir()

	localPath := filepath.Join(backendDir, "dist/anyadmin-agent")

	// Self-contained in user home

	remoteBin := "/home/anyadmin/bin/anyadmin-agent"

	remoteData := "/home/anyadmin/data"

	remoteDataAnything := "/home/anyadmin/data/anythingllm"

	remoteConfig := "/home/anyadmin/bin/config.json"

	logDir := "/home/anyadmin/logs"

	// 1. Prepare Directories

	log.Println("[Deploy] Preparing directories...")

	prepCmd := fmt.Sprintf("mkdir -p /home/anyadmin/bin %s && chown -R anyadmin:anyadmin /home/anyadmin && chmod 755 %s", logDir, logDir)

	if _, err := ExecuteCommand(client, prepCmd); err != nil {

		return fmt.Errorf("failed to prepare agent directories: %w", err)

	}

	// Stop existing agent before copying to avoid "Text file busy"

	log.Println("[Deploy] Stopping existing agent...")

	ExecuteCommand(client, "pkill -f anyadmin-agent || true")

	time.Sleep(1 * time.Second) // Give it a moment to release file handle

	// 2. Create Config File

	log.Println("[Deploy] Creating config file...")

	configContent := fmt.Sprintf(`{"mgmt_host": "%s", "mgmt_port": "%s", "node_ip": "%s", "deployment_time": "%s", "log_file": "/home/anyadmin/logs/agent.log"}`,

		mgmtHost, mgmtPort, nodeIP, time.Now().Format(time.RFC3339))

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
	localComposePath := filepath.Join(backendDir, "deployments/dockers/yaml/docker-compose.yml")
	remoteComposePath := "/home/anyadmin/docker/docker-compose.yaml"
	// Ensure directory exists
	if _, err := ExecuteCommand(client, "mkdir -p /home/anyadmin/docker && chown -R anyadmin:anyadmin /home/anyadmin/docker"); err != nil {
		return fmt.Errorf("failed to create docker directory: %w", err)
	}

	if err := CopyFile(client, localComposePath, remoteComposePath); err != nil {
		log.Printf("Warning: failed to copy docker-compose.yml: %v", err)
	} else {
		ExecuteCommand(client, fmt.Sprintf("chown anyadmin:anyadmin %s", remoteComposePath))
	}

	// Copy Environment files
	log.Println("[Deploy] Copying environment files...")
	localEnvDir := filepath.Join(backendDir, "deployments/dockers/yaml")
	envFiles, err := os.ReadDir(localEnvDir)
	if err == nil {
		for _, file := range envFiles {
			if !file.IsDir() && strings.HasPrefix(file.Name(), ".env") {
				localEnvPath := filepath.Join(localEnvDir, file.Name())
				// Use path.Join for remote Unix paths
				remoteEnvPath := "/home/anyadmin/docker/" + file.Name()
				if err := CopyFile(client, localEnvPath, remoteEnvPath); err != nil {
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

	// 4. Run Agent

	// The agent now looks for config.json in the same directory by default (or we can specify it)

	// We'll run it from the bin directory using absolute paths for everything

	remoteBinAbs := "/home/anyadmin/bin/anyadmin-agent"

	// Wrap in runuser and nohup. Use -c "cd ... && nohup ... > ... < /dev/null &"

	// Redirecting stdin from /dev/null is crucial for nohup via ssh to not hang

	log.Println("[Deploy] Starting agent...")

	fullCmd := fmt.Sprintf("runuser -l anyadmin -c 'cd /home/anyadmin/bin && (nohup %s -config config.json -log /home/anyadmin/logs/agent.log > /home/anyadmin/logs/agent.log 2>&1 < /dev/null &) >/dev/null 2>&1'", remoteBinAbs)

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

	return nil
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

// ControlAgent manages the agent process on a remote node
func ControlAgent(nodeIP, action string) error {
	user := "admin"
	nodeHost := nodeIP
	nodePort := "22"
	if strings.Contains(nodeIP, ":") {
		parts := strings.Split(nodeIP, ":")
		nodeHost = parts[0]
		nodePort = parts[1]
	}

	// Run long operations in background
	go func() {
		RecordLog(user, "Agent Control", fmt.Sprintf("Initiating %s on agent %s", action, nodeIP), "Info")

		client, err := GetSSHClient(nodeHost, nodePort)
		if err != nil {
			msg := fmt.Sprintf("[Agent Control] SSH connection failed to %s: %v", nodeIP, err)
			log.Println(msg)
			RecordLog(user, "Agent Control", msg, "Error")
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

			mockdata.Mu.Lock()
			mgmtHost := mockdata.MgmtHost
			mgmtPort := mockdata.MgmtPort
			mockdata.Mu.Unlock()

			if mgmtHost == "" {
				mockdata.LoadFromFile()
				mockdata.Mu.Lock()
				mgmtHost = mockdata.MgmtHost
				mgmtPort = mockdata.MgmtPort
				mockdata.Mu.Unlock()
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

			mockdata.Mu.Lock()
			mgmtHost := mockdata.MgmtHost
			mgmtPort := mockdata.MgmtPort
			mockdata.Mu.Unlock()

			if mgmtHost == "" {
				mockdata.LoadFromFile()
				mockdata.Mu.Lock()
				mgmtHost = mockdata.MgmtHost
				mgmtPort = mockdata.MgmtPort
				mockdata.Mu.Unlock()
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
