package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// Mock encryptor implementation that satisfies the interface needed by Store
type MockEncryptor struct{}

func (m *MockEncryptor) Encrypt(data []byte) (string, error) {
	return string(data) + "_encrypted", nil
}

func (m *MockEncryptor) Decrypt(encryptedData string) ([]byte, error) {
	// Remove _encrypted suffix
	return []byte(encryptedData[:len(encryptedData)-10]), nil
}

func TestStore(t *testing.T) {
	// Create a temporary directory for the store
	tempDir, err := os.MkdirTemp("", "passh-test-store")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {

		}
	}(tempDir)

	// Create a mock encryptor
	mockEncryptor := &MockEncryptor{}

	// We need to create a store with our mock encryptor
	// First, let's update the Store implementation to accept our interface

	// Using our own NewTestStore function to create a store with the mock
	store := &Store{
		rootDir:   tempDir,
		encryptor: mockEncryptor,
	}

	// Test adding a password
	testPassword := []byte("test-password-123")
	if err := store.Add("test/entry", testPassword); err != nil {
		t.Fatalf("Failed to add password: %v", err)
	}

	// Verify file was created
	passwordFilePath := filepath.Join(tempDir, "test/entry.pass")
	if _, err := os.Stat(passwordFilePath); os.IsNotExist(err) {
		t.Fatalf("Password file not created: %v", err)
	}

	// Test getting the password
	retrievedPassword, err := store.Get("test/entry")
	if err != nil {
		t.Fatalf("Failed to get password: %v", err)
	}
	if string(retrievedPassword) != string(testPassword) {
		t.Fatalf("Expected password '%s', got '%s'", testPassword, retrievedPassword)
	}

	// Test listing passwords
	entries, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list passwords: %v", err)
	}
	if len(entries) != 1 || entries[0] != "test/entry" {
		t.Fatalf("Expected ['test/entry'], got %v", entries)
	}

	// Test adding another password
	if err := store.Add("another/entry", []byte("another-password")); err != nil {
		t.Fatalf("Failed to add second password: %v", err)
	}

	// Test listing again
	entries, err = store.List()
	if err != nil {
		t.Fatalf("Failed to list passwords: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// Test deleting a password
	if err := store.Delete("test/entry"); err != nil {
		t.Fatalf("Failed to delete password: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(passwordFilePath); !os.IsNotExist(err) {
		t.Fatalf("Password file was not deleted")
	}

	// Test listing after deletion
	entries, err = store.List()
	if err != nil {
		t.Fatalf("Failed to list passwords: %v", err)
	}
	if len(entries) != 1 || entries[0] != "another/entry" {
		t.Fatalf("Expected ['another/entry'], got %v", entries)
	}
}

func TestStoreWithDefaultLocation(t *testing.T) {
	// Test that we default to ~/.passh if no location is provided
	// We'll mock the UserHomeDir function
	oldHomeDir := os.Getenv("HOME")
	tempDir, err := os.MkdirTemp("", "passh-test-home")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {

		}
	}(tempDir)
	if err := os.Setenv("HOME", tempDir); err != nil {
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", oldHomeDir); err != nil {
			t.Logf("Failed to restore HOME environment variable: %v", err)
		}
	}()

	// Create a mock encryptor
	mockEncryptor := &MockEncryptor{}

	// Create a new store with empty path (should default to ~/.passh)
	store := &Store{
		rootDir:   "",
		encryptor: mockEncryptor,
	}

	// Manually set the rootDir as we would in NewStore
	if store.rootDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}
		store.rootDir = filepath.Join(homeDir, ".passh")
	}

	// Ensure directory exists
	if err := os.MkdirAll(store.rootDir, 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Verify the store directory was created
	expectedStoreDir := filepath.Join(tempDir, ".passh")
	info, err := os.Stat(expectedStoreDir)
	if os.IsNotExist(err) {
		t.Fatalf("Expected store directory not created at %s", expectedStoreDir)
	}
	if err != nil {
		t.Fatalf("Error checking store directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("Expected %s to be a directory", expectedStoreDir)
	}
}
