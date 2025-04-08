package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestNewSSHEncryptor(t *testing.T) {
	// Test with useAgent=true
	encryptorWithAgent, err := NewSSHEncryptor(true)
	if err != nil {
		t.Fatalf("Failed to create encryptor with agent: %v", err)
	}
	if encryptorWithAgent == nil {
		t.Fatal("Expected non-nil encryptor with agent")
	}
	if !encryptorWithAgent.useAgent {
		t.Fatal("Expected useAgent to be true")
	}

	// Test with useAgent=false
	encryptorWithoutAgent, err := NewSSHEncryptor(false)
	if err != nil {
		t.Fatalf("Failed to create encryptor without agent: %v", err)
	}
	if encryptorWithoutAgent == nil {
		t.Fatal("Expected non-nil encryptor without agent")
	}
	if encryptorWithoutAgent.useAgent {
		t.Fatal("Expected useAgent to be false")
	}
}

func TestEncryptionDecryption(t *testing.T) {
	// Create temporary directory for test keys
	tempDir, err := os.MkdirTemp("", "ssh-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {

		}
	}(tempDir)

	// Generate test SSH key pair
	privateKeyPath, publicKeyPath, err := generateTestKeys(tempDir)
	if err != nil {
		t.Fatalf("Failed to generate test keys: %v", err)
	}

	// Create encryptor
	encryptor, err := NewSSHEncryptor(false)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	// Add keys
	if err := encryptor.AddPublicKeyFromFile(publicKeyPath); err != nil {
		t.Fatalf("Failed to add public key: %v", err)
	}
	if err := encryptor.AddPrivateKeyFromFile(privateKeyPath, nil); err != nil {
		t.Fatalf("Failed to add private key: %v", err)
	}

	// Test encryption/decryption
	testData := []byte("This is a test password")
	encrypted, err := encryptor.Encrypt(testData)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	if encrypted == "" {
		t.Fatal("Expected non-empty encrypted result")
	}

	// Decrypt and verify
	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}
	if string(decrypted) != string(testData) {
		t.Fatalf("Expected '%s', got '%s'", testData, decrypted)
	}
}

// Helper function to generate test SSH keys
func generateTestKeys(dir string) (privateKeyPath, publicKeyPath string, err error) {
	privateKeyPath = filepath.Join(dir, "id_test")
	publicKeyPath = filepath.Join(dir, "id_test.pub")

	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Convert to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write private key to file
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return "", "", err
	}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		err := privateKeyFile.Close()
		if err != nil {
			return "", "", err
		}
		return "", "", err
	}
	if err := privateKeyFile.Close(); err != nil {
		return "", "", err
	}

	// Change file permissions
	if err := os.Chmod(privateKeyPath, 0600); err != nil {
		return "", "", err
	}

	// Generate public key
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	// Write public key to file
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	if err := os.WriteFile(publicKeyPath, publicKeyBytes, 0644); err != nil {
		return "", "", err
	}

	return privateKeyPath, publicKeyPath, nil
}
