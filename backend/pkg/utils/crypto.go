package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
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

	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		rsaPub,
		[]byte(password),
		nil,
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

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
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

	decryptedBytes, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		priv,
		encryptedBytes,
		nil,
	)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
}
