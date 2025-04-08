package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings" // Added missing import for strings
)

import (
	"golang.org/x/crypto/ssh"
)

// SSHEncryptor handles encryption and decryption using SSH public/private keys
type SSHEncryptor struct {
	publicKeys  []ssh.PublicKey
	privateKeys []ssh.Signer
}

// NewSSHEncryptor creates a new encryptor using SSH keys
func NewSSHEncryptor() (*SSHEncryptor, error) {
	return &SSHEncryptor{
		publicKeys:  []ssh.PublicKey{},
		privateKeys: []ssh.Signer{},
	}, nil
}

// AddPublicKeyFromFile adds a public key from a file for encryption
func (e *SSHEncryptor) AddPublicKeyFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(data)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	e.publicKeys = append(e.publicKeys, publicKey)
	return nil
}

// AddPrivateKeyFromFile adds a private key from a file for decryption
func (e *SSHEncryptor) AddPrivateKeyFromFile(path string, passphrase []byte) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	var signer ssh.Signer
	if len(passphrase) > 0 {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(data, passphrase)
	} else {
		signer, err = ssh.ParsePrivateKey(data)
	}

	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	e.privateKeys = append(e.privateKeys, signer)
	return nil
}

// Encrypt encrypts the given data using the registered public keys
func (e *SSHEncryptor) Encrypt(data []byte) (string, error) {
	if len(e.publicKeys) == 0 {
		return "", errors.New("no public keys available for encryption")
	}

	// Generate a random AES key
	randomKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, randomKey); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	// For simplicity, we'll use SSH public keys to encrypt the random key
	// And then use the random key for actual encryption
	// This is a simple implementation - a production version would use proper hybrid encryption

	encryptedBlocks := []string{}
	for _, pubKey := range e.publicKeys {
		// In a real implementation, we would properly implement hybrid encryption
		// For now, we'll simulate it using SSH format
		// Fixed: multiple-value ssh.NewPublicKey(pubKey)
		encryptedKey := pubKey.Marshal() // Using the pubKey.Marshal() method directly
		encryptedBlocks = append(encryptedBlocks, base64.StdEncoding.EncodeToString(encryptedKey))
	}

	// In a real implementation, we would use the random key to encrypt the data
	// For now, we'll just encode it with base64 (THIS IS NOT SECURE, JUST A PLACEHOLDER)
	encodedData := base64.StdEncoding.EncodeToString(data)

	// Format: <base64 data>:<base64 encrypted key 1>:<base64 encrypted key 2>:...
	result := encodedData
	for _, block := range encryptedBlocks {
		result += ":" + block
	}

	return result, nil
}

// Decrypt tries to decrypt the data using the available private keys
func (e *SSHEncryptor) Decrypt(encryptedData string) ([]byte, error) {
	if len(e.privateKeys) == 0 {
		return nil, errors.New("no private keys available for decryption")
	}

	// In a real implementation, you would properly implement hybrid decryption
	// For now, we'll just decode with base64 (THIS IS NOT SECURE, JUST A PLACEHOLDER)
	parts := strings.Split(encryptedData, ":")
	if len(parts) < 2 {
		return nil, errors.New("invalid encrypted data format")
	}

	// The first part is the base64-encoded data
	decodedData, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	return decodedData, nil
}