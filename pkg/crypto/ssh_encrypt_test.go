package crypto

import (
	"fmt"
	"io"
	"os"
	"os/exec"
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

func TestSSHEnvironmentVariables(t *testing.T) {
	// Test when SSH_AUTH_SOCK is not set
	origAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if err := os.Unsetenv("SSH_AUTH_SOCK"); err != nil {
		t.Fatalf("Failed to unset SSH_AUTH_SOCK: %v", err)
	}

	encryptorNoAgent, err := NewSSHEncryptor(true)
	if err != nil {
		t.Fatalf("Failed to create encryptor without SSH_AUTH_SOCK: %v", err)
	}

	if encryptorNoAgent.agentClient != nil {
		t.Fatal("Expected nil agentClient when SSH_AUTH_SOCK is not set")
	}

	// Restore original value
	if origAuthSock != "" {
		if err := os.Setenv("SSH_AUTH_SOCK", origAuthSock); err != nil {
			t.Fatalf("Failed to restore SSH_AUTH_SOCK: %v", err)
		}
	}
}

func TestEncryptionDecryption(t *testing.T) {
	// Create temporary directory for test keys
	tempDir, err := os.MkdirTemp("", "ssh-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp directory: %v", err)
		}
	}()

	// Generate test SSH key pair
	privateKeyPath, publicKeyPath, err := generateTestKeys(t, tempDir)
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

// Helper function to generate test SSH keys - using Ed25519
func generateTestKeys(t *testing.T, dir string) (privateKeyPath, publicKeyPath string, err error) {
	privateKeyPath = filepath.Join(dir, "id_test")
	publicKeyPath = filepath.Join(dir, "id_test.pub")

	// First try to use ssh-keygen to generate a real Ed25519 key
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", privateKeyPath, "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Fall back to creating a mock key for testing environments without ssh-keygen
		t.Logf("ssh-keygen failed (%v), creating mock keys: %s", err, output)

		// Create a mock private key file
		if err := os.WriteFile(privateKeyPath, []byte("-----BEGIN MOCK ED25519 PRIVATE KEY-----\nMOCK PRIVATE KEY CONTENT\n-----END MOCK ED25519 PRIVATE KEY-----"), 0600); err != nil {
			return "", "", err
		}

		// Create a mock public key file
		if err := os.WriteFile(publicKeyPath, []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMOCK+KEY+FOR+TESTING test@localhost"), 0644); err != nil {
			return "", "", err
		}
	} else {
		// Verify the files exist
		if _, err := os.Stat(privateKeyPath); err != nil {
			return "", "", err
		}
		if _, err := os.Stat(publicKeyPath); err != nil {
			return "", "", err
		}
	}

	return privateKeyPath, publicKeyPath, nil
}

// Mock implementation of ssh.Signer for testing
type mockSigner struct{}

func (s *mockSigner) PublicKey() ssh.PublicKey {
	return &mockPublicKey{}
}

func (s *mockSigner) Sign(_ io.Reader, _ []byte) (*ssh.Signature, error) {
	return &ssh.Signature{
		Format: "mock",
		Blob:   []byte("mock-signature"),
	}, nil
}

// Mock implementation of ssh.PublicKey for testing
type mockPublicKey struct{}

func (p *mockPublicKey) Type() string {
	return "ssh-ed25519"
}

func (p *mockPublicKey) Marshal() []byte {
	return []byte("mock-ed25519-key")
}

func (p *mockPublicKey) Verify(_ []byte, sig *ssh.Signature) error {
	if sig.Format != "mock" {
		return fmt.Errorf("signature format mismatch")
	}
	return nil
}
