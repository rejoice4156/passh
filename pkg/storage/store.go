package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rejoice4156/passh/pkg/crypto"
)

// Store handles the storage and retrieval of password entries
type Store struct {
	rootDir   string
	encryptor crypto.Encryptor
}

// NewStore creates a new password store
func NewStore(rootDir string, encryptor crypto.Encryptor) (*Store, error) {
	if rootDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		rootDir = filepath.Join(homeDir, ".passh")
	}

	if err := os.MkdirAll(rootDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	return &Store{
		rootDir:   rootDir,
		encryptor: encryptor,
	}, nil
}

// Add adds a new password entry
func (s *Store) Add(name string, password []byte) error {
	// Encrypt the password
	encryptedData, err := s.encryptor.Encrypt(password)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Ensure the directory structure exists
	dir := filepath.Dir(filepath.Join(s.rootDir, name))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Write the encrypted data to the file
	filePath := filepath.Join(s.rootDir, name+".pass")
	if err := os.WriteFile(filePath, []byte(encryptedData), 0600); err != nil {
		return fmt.Errorf("failed to write password file: %w", err)
	}

	return nil
}

// Get retrieves a password entry
func (s *Store) Get(name string) ([]byte, error) {
	filePath := filepath.Join(s.rootDir, name+".pass")

	encryptedData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read password file: %w", err)
	}

	// Decrypt the password
	password, err := s.encryptor.Decrypt(string(encryptedData))
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return password, nil
}

// List returns all password entries
func (s *Store) List() ([]string, error) {
	var entries []string

	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".pass") {
			// Get relative path and remove the .pass extension
			relPath, err := filepath.Rel(s.rootDir, path)
			if err != nil {
				return err
			}
			entry := strings.TrimSuffix(relPath, ".pass")
			entries = append(entries, entry)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list password entries: %w", err)
	}

	return entries, nil
}

// Delete removes a password entry
func (s *Store) Delete(name string) error {
	filePath := filepath.Join(s.rootDir, name+".pass")

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete password file: %w", err)
	}

	return nil
}
