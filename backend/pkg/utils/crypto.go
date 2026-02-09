package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

var (
	PrivateKeyPath string
	PublicKeyPath  string
)

func init() {
	// Find keys in backend directory similar to store.go
	cwd, _ := os.Getwd()
	checkPaths := []string{
		filepath.Join(cwd, "backend", "keys"),
		filepath.Join(cwd, "..", "backend", "keys"),
		filepath.Join(cwd, "..", "..", "backend", "keys"),
		filepath.Join(cwd, "keys"),
	}

	for _, p := range checkPaths {
		if _, err := os.Stat(filepath.Join(p, "id_rsa")); err == nil {
			PrivateKeyPath = filepath.Join(p, "id_rsa")
			PublicKeyPath = filepath.Join(p, "id_rsa.pub")
			break
		}
	}
}

// EnsureKeysExist checks if the RSA keys exist, and generates them if not.
func EnsureKeysExist() error {
	cwd, _ := os.Getwd()
	keysDir := ""
	checkPaths := []string{
		filepath.Join(cwd, "backend", "keys"),
		filepath.Join(cwd, "..", "backend", "keys"),
		filepath.Join(cwd, "keys"),
	}

	for _, p := range checkPaths {
		if _, err := os.Stat(p); err == nil {
			keysDir = p
			break
		}
	}

	if keysDir == "" {
		// Create keys directory in backend if not found
		if _, err := os.Stat(filepath.Join(cwd, "backend")); err == nil {
			keysDir = filepath.Join(cwd, "backend", "keys")
		} else {
			keysDir = filepath.Join(cwd, "keys")
		}
		os.MkdirAll(keysDir, 0755)
	}

	PrivateKeyPath = filepath.Join(keysDir, "id_rsa")
	PublicKeyPath = filepath.Join(keysDir, "id_rsa.pub")

	if _, err := os.Stat(PrivateKeyPath); err == nil {
		return nil // Keys already exist
	}

	fmt.Println("Generating RSA keys...")
	// Generate key pair
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Save Private Key
	privFile, err := os.OpenFile(PrivateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer privFile.Close()

	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}
	if err := pem.Encode(privFile, privBlock); err != nil {
		return err
	}

	// Save Public Key (OpenSSH format for compatibility with remote agent scripts if needed, but we also use PKIX for frontend)
	// Actually the prompt specifically asked for id_rsa and id_rsa.pub.
	// We'll save the public key in PKIX format as well for our GetPublicKeyContent.
	pubFile, err := os.OpenFile(PublicKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer pubFile.Close()

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return err
	}
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	if err := pem.Encode(pubFile, pubBlock); err != nil {
		return err
	}

	return nil
}

// GetPublicKeyContent reads and returns the content of the public key file
// Since id_rsa.pub might be in OpenSSH format, we derive the Public Key from the Private Key
// and export it as PEM (PKIX) which is required for JSEncrypt and Go's x509.
func GetPublicKeyContent() (string, error) {
	if PrivateKeyPath == "" {
		return "", fmt.Errorf("private key not found")
	}

	privContent, err := os.ReadFile(PrivateKeyPath)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(privContent)
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the private key")
	}

	var priv *rsa.PrivateKey
	priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try parsing as PKCS8
		key, errPKCS8 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if errPKCS8 != nil {
			return "", fmt.Errorf("failed to parse private key: %v", err)
		}
		var ok bool
		priv, ok = key.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("key is not an RSA private key")
		}
	}

	// Marshal Public Key to PKIX (Standard for Web/Java/etc)
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", err
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return string(pubPEM), nil
}

// EncryptPassword encrypts a password using the RSA public key and returns a base64 string
func EncryptPassword(password string) (string, error) {
	pubKeyContent, err := GetPublicKeyContent()
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode([]byte(pubKeyContent))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// Try parsing as PKCS1 if PKIX fails
		rsaPub, errPKCS1 := x509.ParsePKCS1PublicKey(block.Bytes)
		if errPKCS1 != nil {
			return "", fmt.Errorf("failed to parse public key: %v", err)
		}
		pub = rsaPub
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("key is not an RSA public key")
	}

	encryptedBytes, err := rsa.EncryptPKCS1v15(
		rand.Reader,
		rsaPub,
		[]byte(password),
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

// DecryptPassword decrypts a base64 encoded encrypted password using the RSA private key
func DecryptPassword(encryptedPassword string) (string, error) {
	if PrivateKeyPath == "" {
		return "", fmt.Errorf("private key not found")
	}

	privKeyContent, err := os.ReadFile(PrivateKeyPath)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(privKeyContent)
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the private key")
	}

	var priv *rsa.PrivateKey
	priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try parsing as PKCS8
		key, errPKCS8 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if errPKCS8 != nil {
			return "", fmt.Errorf("failed to parse private key: %v", err)
		}
		var ok bool
		priv, ok = key.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("key is not an RSA private key")
		}
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := rsa.DecryptPKCS1v15(
		rand.Reader,
		priv,
		encryptedBytes,
	)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
}
