package crypto

// Encryptor defines the interface for encryption/decryption operations
type Encryptor interface {
	Encrypt(data []byte) (string, error)
	Decrypt(encryptedData string) ([]byte, error)
}
