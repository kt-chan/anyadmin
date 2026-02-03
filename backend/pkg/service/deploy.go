package service

import (
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
	"anyadmin-backend/pkg/mockdata"
	"anyadmin-backend/pkg/global"

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
	backendDir := getBackendDir()
	// Source path
	localPath := filepath.Join(backendDir, "deployments/tars/os/ubuntu/amd64/jammy/go1.25.6.linux-amd64.tar.gz")
	remotePath := "/tmp/go.tar.gz"

	// 1. Calculate local hash
	localHash, err := calculateHash(localPath)
	if err != nil {
		return fmt.Errorf("failed to calc local hash: %w", err)
	}

	// 2. Transfer
	if err := CopyFile(client, localPath, remotePath); err != nil {
		return fmt.Errorf("failed to transfer Go tarball: %w", err)
	}

	// 3. Verify Remote Hash
	output, err := ExecuteCommand(client, fmt.Sprintf("sha256sum %s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to verify remote hash: %w", err)
	}
	// Output format: "hash  filename"
	if len(strings.Fields(output)) == 0 {
		return fmt.Errorf("sha256sum returned empty output")
	}
	remoteHash := strings.Fields(output)[0]
	if remoteHash != localHash {
		return fmt.Errorf("checksum mismatch: local %s != remote %s", localHash, remoteHash)
	}

	// 4. Install
	// Remove old installation and extract new
	cmd := fmt.Sprintf("rm -rf /usr/local/go && tar -C /usr/local -xzf %s", remotePath)
	if _, err := ExecuteCommand(client, cmd); err != nil {
		return fmt.Errorf("failed to extract Go: %w", err)
	}

	// Add to PATH (system-wide or for users)
	// We'll add to /etc/profile.d which loads for all shells
	cmd = "echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh && chmod +x /etc/profile.d/go.sh"
	ExecuteCommand(client, cmd)

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

	remoteConfig := "/home/anyadmin/bin/config.json"

	logDir := "/home/anyadmin/logs"



	// 1. Prepare Directories

	log.Println("[Deploy] Preparing directories...")

	prepCmd := fmt.Sprintf("mkdir -p /home/anyadmin/bin %s && chown -R anyadmin:anyadmin /home/anyadmin", logDir)

	if _, err := ExecuteCommand(client, prepCmd); err != nil {

		return fmt.Errorf("failed to prepare agent directories: %w", err)

	}



	// Stop existing agent before copying to avoid "Text file busy"

	log.Println("[Deploy] Stopping existing agent...")

	ExecuteCommand(client, "pkill -f anyadmin-agent || true")

	time.Sleep(1 * time.Second) // Give it a moment to release file handle



	// 2. Create Config File

	log.Println("[Deploy] Creating config file...")

	configContent := fmt.Sprintf(`{"mgmt_host": "%s", "mgmt_port": "%s", "node_ip": "%s", "deployment_time": "%s"}`, 

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
	if _, err := ExecuteCommand(client, "mkdir -p /home/anyadmin/docker/env && chown -R anyadmin:anyadmin /home/anyadmin/docker"); err != nil {
		return fmt.Errorf("failed to create docker directory: %w", err)
	}
	
	if err := CopyFile(client, localComposePath, remoteComposePath); err != nil {
		log.Printf("Warning: failed to copy docker-compose.yml: %v", err)
	} else {
		ExecuteCommand(client, fmt.Sprintf("chown anyadmin:anyadmin %s", remoteComposePath))
	}

	// Copy Environment files
	log.Println("[Deploy] Copying environment files...")
	localEnvDir := filepath.Join(backendDir, "deployments/dockers/yaml/env")
	envFiles, err := os.ReadDir(localEnvDir)
	var mergedEnv strings.Builder
	if err == nil {
		for _, file := range envFiles {
			if !file.IsDir() {
				localEnvPath := filepath.Join(localEnvDir, file.Name())
				// Use path.Join for remote Unix paths
				remoteEnvPath := "/home/anyadmin/docker/env/" + file.Name()
				if err := CopyFile(client, localEnvPath, remoteEnvPath); err != nil {
					log.Printf("Warning: failed to copy env file %s: %v", file.Name(), err)
				} else {
					ExecuteCommand(client, fmt.Sprintf("chown anyadmin:anyadmin %s", remoteEnvPath))
				}

				// Read for merging into main .env for interpolation
				content, _ := os.ReadFile(localEnvPath)
				mergedEnv.Write(content)
				mergedEnv.WriteString("\n")
			}
		}
		
		// Create master .env for interpolation
		localTempEnv := filepath.Join(os.TempDir(), "master.env")
		os.WriteFile(localTempEnv, []byte(mergedEnv.String()), 0644)
		CopyFile(client, localTempEnv, "/home/anyadmin/docker/.env")
		ExecuteCommand(client, "chown anyadmin:anyadmin /home/anyadmin/docker/.env")
		os.Remove(localTempEnv)
	} else {
		log.Printf("Warning: failed to read local env directory: %v", err)
	}

	// Ensure executable and owned by anyadmin

	log.Println("[Deploy] Setting permissions...")

	ExecuteCommand(client, fmt.Sprintf("chmod +x %s && chown anyadmin:anyadmin %s %s", remoteBin, remoteBin, remoteConfig))



	// 4. Run Agent

	// The agent now looks for config.json in the same directory by default (or we can specify it)

	// We'll run it from the bin directory using absolute paths for everything

	remoteBinAbs := "/home/anyadmin/bin/anyadmin-agent"

	

		// Wrap in runuser and nohup. Use -c "cd ... && nohup ... > ... < /dev/null &"

	

		// Redirecting stdin from /dev/null is crucial for nohup via ssh to not hang

	

		log.Println("[Deploy] Starting agent...")

	

		fullCmd := fmt.Sprintf("runuser -l anyadmin -c 'cd /home/anyadmin/bin && (nohup %s > %s/agent.log 2>&1 < /dev/null &) >/dev/null 2>&1'", remoteBinAbs, logDir)

	

		

	

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
