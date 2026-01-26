package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSSHKeyGeneration(t *testing.T) {
	// Temporarily redirect keyDir
	tempDir, err := os.MkdirTemp("", "ssh_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	originalKeyDir := keyDir
	keyDir = tempDir
	defer func() { keyDir = originalKeyDir }()

	// Test Generation
	err = EnsureKeys()
	if err != nil {
		t.Fatalf("EnsureKeys failed: %v", err)
	}

	// Verify files
	if _, err := os.Stat(filepath.Join(tempDir, "id_rsa")); os.IsNotExist(err) {
		t.Error("Private key was not created")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "id_rsa.pub")); os.IsNotExist(err) {
		t.Error("Public key was not created")
	}

	// Test Retrieval
	pubKey, err := GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey failed: %v", err)
	}
	if len(pubKey) == 0 {
		t.Error("Public key is empty")
	}
}
