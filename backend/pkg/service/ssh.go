package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	keyDir  = ""
	keyName = "id_rsa"
	keyMu   sync.Mutex
)

func init() {
	// Find backend directory to locate keys
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
				keyDir = filepath.Join(p, "keys")
				break
			}
		}
	}
	if keyDir == "" {
		keyDir = "./keys" // Fallback
	}
}

// EnsureKeys checks for keys and generates them if missing.
func EnsureKeys() error {
	keyMu.Lock()
	defer keyMu.Unlock()

	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		if err := os.MkdirAll(keyDir, 0700); err != nil {
			return fmt.Errorf("failed to create key directory: %w", err)
		}
	}

	privPath := filepath.Join(keyDir, keyName)
	pubPath := filepath.Join(keyDir, keyName+".pub")

	if _, err := os.Stat(privPath); err == nil {
		if _, err := os.Stat(pubPath); err == nil {
			return nil // Keys exist
		}
	}

	// Generate keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Save private key
	privFile, err := os.OpenFile(privPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}
	defer privFile.Close()

	privPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(privFile, privPEM); err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	// Generate and save public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to generate public key: %w", err)
	}

	pubBytes := ssh.MarshalAuthorizedKey(pub)
	if err := os.WriteFile(pubPath, pubBytes, 0644); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	return nil
}

// GetPublicKey returns the content of the public key.
func GetPublicKey() (string, error) {
	if err := EnsureKeys(); err != nil {
		return "", err
	}
	content, err := os.ReadFile(filepath.Join(keyDir, keyName+".pub"))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetSSHClient establishes an SSH connection with retries
func GetSSHClient(host string, port string) (*ssh.Client, error) {
	if err := EnsureKeys(); err != nil {
		return nil, err
	}

	privPath := filepath.Join(keyDir, keyName)
	key, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: "root", // Default to root
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
			ssh.Password("password"), // Fallback to default password
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Handle Host:Port or just Host
	addr := host
	if !strings.Contains(host, ":") {
		addr = net.JoinHostPort(host, port)
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		client, err := ssh.Dial("tcp", addr, config)
		if err == nil {
			return client, nil
		}
		lastErr = err
		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("failed to dial %s after 3 attempts: %w", addr, lastErr)
}

// ExecuteCommand runs a command on the remote host
func ExecuteCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w, output: %s", err, output)
	}
	return string(output), nil
}

// CopyFile transfers a local file to the remote host using base64 encoding for reliability
func CopyFile(client *ssh.Client, localPath, remotePath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Stat(); err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	// Use base64 to ensure binary integrity across shell pipes
	cmd := fmt.Sprintf("mkdir -p %s && base64 -d > %s", filepath.Dir(remotePath), remotePath)
	
	go func() {
		defer stdin.Close()
		encoder := base64.NewEncoder(base64.StdEncoding, stdin)
		defer encoder.Close()
		io.Copy(encoder, f)
	}()

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	
	return nil
}

// CheckSSHConnection verifies passwordless access to a node.
func CheckSSHConnection(host string, port int) error {
	client, err := GetSSHClient(host, fmt.Sprintf("%d", port))
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = ExecuteCommand(client, "echo hello")
	return err
}
