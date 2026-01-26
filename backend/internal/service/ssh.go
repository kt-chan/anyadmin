package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	keyDir  = "./keys"
	keyName = "id_rsa"
	keyMu   sync.Mutex
)

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

// CheckSSHConnection verifies passwordless access to a node.
func CheckSSHConnection(host string, port int) error {
	if err := EnsureKeys(); err != nil {
		return err
	}

	privPath := filepath.Join(keyDir, keyName)
	key, err := os.ReadFile(privPath)
	if err != nil {
		return fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: "root", // Default to root, maybe parameterize later
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For demo purposes, ignore host key verification
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		// Try to give a better error message
		if strings.Contains(err.Error(), "unable to authenticate") {
			return fmt.Errorf("authentication failed: key not authorized")
		}
		return fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	// Try running a simple command to ensure we really have access
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	if err := session.Run("echo hello"); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}
