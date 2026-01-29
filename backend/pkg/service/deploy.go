package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"anyadmin-backend/pkg/global"

	"golang.org/x/crypto/ssh"
)

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

	// Create user with sudo access
	// -m: create home directory
	// -s: shell
	// -G: groups (sudo)
	cmd := fmt.Sprintf("useradd -m -s /bin/bash -G sudo %s", username)
	if _, err := ExecuteCommand(client, cmd); err != nil {
		return fmt.Errorf("failed to add user: %w", err)
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
	// Source path
	localPath := "./deployments/tars/os/ubuntu/amd64/jammy/go1.25.6.linux-amd64.tar.gz"
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
	localPath := "./dist/agent_linux"
	// Self-contained in user home
	remoteBin := "/home/anyadmin/bin/anyadmin-agent"
	logDir := "/home/anyadmin/logs"

	// 1. Prepare Directories
	// Create bin and logs directories and ensure ownership of the entire home dir
	prepCmd := fmt.Sprintf("mkdir -p /home/anyadmin/bin %s && chown -R anyadmin:anyadmin /home/anyadmin", logDir)
	if _, err := ExecuteCommand(client, prepCmd); err != nil {
		return fmt.Errorf("failed to prepare agent directories: %w", err)
	}

	// 2. Copy Binary
	if err := CopyFile(client, localPath, remoteBin); err != nil {
		return fmt.Errorf("failed to copy agent binary: %w", err)
	}
	// Ensure executable and owned by anyadmin (CopyFile might create it as root)
	ExecuteCommand(client, fmt.Sprintf("chmod +x %s && chown anyadmin:anyadmin %s", remoteBin, remoteBin))

	// 3. Run Agent
	// Command to run:
	// nohup /home/anyadmin/bin/anyadmin-agent -server http://<mgmt>:port -ip <nodeIP> > /home/anyadmin/logs/agent.log 2>&1 &
	
	agentCmd := fmt.Sprintf("%s -server http://%s:%s -ip %s", remoteBin, mgmtHost, mgmtPort, nodeIP)
	
	// Wrap in runuser and nohup
	fullCmd := fmt.Sprintf("runuser -l anyadmin -c 'nohup %s > %s/agent.log 2>&1 &'", agentCmd, logDir)
	
	if _, err := ExecuteCommand(client, fullCmd); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

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
