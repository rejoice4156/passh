package cli

import (
	"context" // Add missing context import
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/rejoice4156/passh/pkg/crypto"
	"github.com/rejoice4156/passh/pkg/storage"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Context keys
const (
	encryptorKey contextKey = "encryptor"
)

// Default SSH key paths
var (
	defaultSSHDir         = filepath.Join(os.Getenv("HOME"), ".ssh")
	defaultSSHPrivateKeys = []string{"id_rsa", "id_ed25519"}
	defaultSSHPublicKeys  = []string{"id_rsa.pub", "id_ed25519.pub"}
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	var storeDir string
	var publicKeyPath string
	var privateKeyPath string

	rootCmd := &cobra.Command{
		Use:   "passh",
		Short: "A terminal password manager backed by SSH keys",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip setup for completion and help commands
			if cmd.Name() == "completion" || cmd.Name() == "help" {
				return nil
			}

			return setupEncryptor(cmd, publicKeyPath, privateKeyPath)
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&storeDir, "store", "", "Password store directory (default: ~/.passh)")
	rootCmd.PersistentFlags().StringVar(&publicKeyPath, "public-key", "", "SSH public key path (default: ~/.ssh/id_rsa.pub or ~/.ssh/id_ed25519.pub)")
	rootCmd.PersistentFlags().StringVar(&privateKeyPath, "private-key", "", "SSH private key path (default: ~/.ssh/id_rsa or ~/.ssh/id_ed25519)")

	// Add subcommands
	rootCmd.AddCommand(
		newAddCmd(),
		newGetCmd(),
		newListCmd(),
		newDeleteCmd(),
		newGenerateCmd(),
	)

	return rootCmd
}

// setupEncryptor initializes the SSH encryptor and attaches it to the command context
func setupEncryptor(cmd *cobra.Command, publicKeyPath, privateKeyPath string) error {
	encryptor, err := crypto.NewSSHEncryptor()
	if err != nil {
		return fmt.Errorf("failed to create encryptor: %w", err)
	}

	// Try to find SSH keys if not specified
	if publicKeyPath == "" {
		for _, name := range defaultSSHPublicKeys {
			path := filepath.Join(defaultSSHDir, name)
			if _, err := os.Stat(path); err == nil {
				publicKeyPath = path
				break
			}
		}
	}

	if privateKeyPath == "" {
		for _, name := range defaultSSHPrivateKeys {
			path := filepath.Join(defaultSSHDir, name)
			if _, err := os.Stat(path); err == nil {
				privateKeyPath = path
				break
			}
		}
	}

	if publicKeyPath == "" {
		return fmt.Errorf("no SSH public key found, specify with --public-key")
	}

	if privateKeyPath == "" {
		return fmt.Errorf("no SSH private key found, specify with --private-key")
	}

	// Load the keys
	if err := encryptor.AddPublicKeyFromFile(publicKeyPath); err != nil {
		return fmt.Errorf("failed to load public key: %w", err)
	}
	// Store the encryptor in the command context
	ctx := context.WithValue(cmd.Context(), encryptorKey, encryptor)
	cmd.SetContext(ctx)
	// This is simplified for now
	if err := encryptor.AddPrivateKeyFromFile(privateKeyPath, nil); err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}
	
	return nil
}

// getStore gets the storage from command context
func getStore(cmd *cobra.Command) (*storage.Store, error) {
	storeDir, _ := cmd.Flags().GetString("store")
	encryptor := cmd.Context().Value(encryptorKey).(*crypto.SSHEncryptor)

	return storage.NewStore(storeDir, encryptor)
}
