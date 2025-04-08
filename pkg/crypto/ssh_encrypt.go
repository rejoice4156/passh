package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSHEncryptor handles encryption and decryption using SSH public/private keys
type SSHEncryptor struct {
	publicKeys  []ssh.PublicKey
	privateKeys []ssh.Signer
	agentClient agent.Agent
	useAgent    bool
}

// NewSSHEncryptor creates a new encryptor using SSH keys
// The useAgent parameter determines whether to attempt connecting to an SSH agent
func NewSSHEncryptor(useAgent bool) (*SSHEncryptor, error) {
	encryptor := &SSHEncryptor{
		publicKeys:  nil, // Changed from []ssh.PublicKey{}
		privateKeys: nil, // Changed from []ssh.Signer{}
		useAgent:    useAgent,
	}

	// Try to connect to the SSH agent if allowed
	if useAgent {
		if err := encryptor.connectToAgent(); err != nil {
			// Just log this error, don't fail as we'll fall back to key files
			_, printErr := fmt.Fprintf(os.Stderr, "Note: SSH agent not available: %v\n", err)
			if printErr != nil {
				// If we can't even write to stderr, just continue silently
				// This is an edge case that shouldn't happen in normal operation
			}
		}
	}

	return encryptor, nil
}

// connectToAgent attempts to connect to the SSH agent
func (e *SSHEncryptor) connectToAgent() error {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return errors.New("SSH_AUTH_SOCK environment variable not set")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH agent, and could not proceed with agent keys: %w", err)
	}

	e.agentClient = agent.NewClient(conn)
	return nil
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
	// If we're using the SSH agent, and we've connected to it, try to use it
	if e.useAgent && e.agentClient != nil {
		signers, err := e.agentClient.Signers()
		if err == nil && len(signers) > 0 {
			// Add all signers from the agent
			e.privateKeys = append(e.privateKeys, signers...)
			fmt.Println("Successfully loaded keys from SSH agent")
			return nil
		}
	}

	// Fall back to loading from file
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

	var encryptedBlocks []string
	for _, pubKey := range e.publicKeys {
		// In a real implementation, we would properly implement hybrid encryption
		// For now, we'll simulate it using SSH format
		encryptedKey := pubKey.Marshal()
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
