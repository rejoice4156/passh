package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/rejoice4156/passh/pkg/crypto"
	"github.com/rejoice4156/passh/pkg/storage"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Default SSH key paths - prioritize modern Ed25519 keys over RSA
var (
	defaultSSHDir         = filepath.Join(os.Getenv("HOME"), ".ssh")
	defaultSSHPrivateKeys = []string{"id_ed25519", "id_ecdsa", "id_rsa"} // Ed25519 first, RSA last
	defaultSSHPublicKeys  = []string{"id_ed25519.pub", "id_ecdsa.pub", "id_rsa.pub"}
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	var storeDir string
	var publicKeyPath string
	var privateKeyPath string
	var noAgent bool

	rootCmd := &cobra.Command{
		Use:   "passh",
		Short: "A terminal password manager backed by SSH keys",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip setup for completion and help commands
			if cmd.Name() == "completion" || cmd.Name() == "help" {
				return nil
			}

			// Check for SSH environment first
			if err := checkSSHEnvironment(); err != nil {
				return err
			}

			return setupEncryptor(cmd, publicKeyPath, privateKeyPath, noAgent)
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&storeDir, "store", "", "Password store directory (default: ~/.passh)")
	rootCmd.PersistentFlags().StringVar(&publicKeyPath, "public-key", "", "SSH public key path (default: ~/.ssh/id_ed25519.pub)")
	rootCmd.PersistentFlags().StringVar(&privateKeyPath, "private-key", "", "SSH private key path (default: ~/.ssh/id_ed25519)")
	rootCmd.PersistentFlags().BoolVar(&noAgent, "no-agent", false, "Don't use SSH agent even if available")

	// Add subcommands
	rootCmd.AddCommand(
		newSetupCmd(),
		newVersionCmd(),
		newAddCmd(),
		newGetCmd(),
		newListCmd(),
		newDeleteCmd(),
		newGenerateCmd(),
	)

	return rootCmd
}

// checkSSHEnvironment verifies that SSH is installed and keys are available
func checkSSHEnvironment() error {
	// Check if ssh is installed
	if _, err := exec.LookPath("ssh"); err != nil {
		return fmt.Errorf("SSH is not installed or not in PATH. Please install SSH before using passh:\n" +
			"  - On Debian/Ubuntu: sudo apt-get install openssh-client\n" +
			"  - On Fedora/RHEL: sudo dnf install openssh-clients\n" +
			"  - On macOS: brew install openssh\n" +
			"  - On Windows: Install Git for Windows or OpenSSH via Windows Optional Features")
	}

	// Check for existing SSH keys
	keysExist := false
	for _, keyName := range defaultSSHPrivateKeys {
		keyPath := filepath.Join(defaultSSHDir, keyName)
		if _, err := os.Stat(keyPath); err == nil {
			keysExist = true
			break
		}
	}

	if !keysExist {
		// Suggest creating SSH keys
		return fmt.Errorf("No SSH keys found. Please create SSH keys before using passh:\n\n" +
			"To create a new Ed25519 key (recommended):\n" +
			"  ssh-keygen -t ed25519\n\n" +
			"To add your key to the SSH agent:\n" +
			"  ssh-add ~/.ssh/id_ed25519\n\n" +
			"After creating keys, run passh again.")
	}

	// Check if SSH agent is running
	agentSock := os.Getenv("SSH_AUTH_SOCK")
	if agentSock == "" {
		fmt.Println("Note: SSH agent is not running. You may need to enter your key passphrase repeatedly.")
		fmt.Println("To start the SSH agent:")
		fmt.Println("  eval `ssh-agent`")
		fmt.Println("  ssh-add")
	}

	return nil
}

// setupEncryptor initializes the SSH encryptor and attaches it to the command context
func setupEncryptor(cmd *cobra.Command, publicKeyPath, privateKeyPath string, noAgent bool) error {
	// Pass the inverse of noAgent to indicate whether to use the agent
	encryptor, err := crypto.NewSSHEncryptor(!noAgent)
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

	// First try without passphrase
	err = encryptor.AddPrivateKeyFromFile(privateKeyPath, nil)
	if err != nil && isPassphraseError(err) {
		// If it fails due to passphrase, prompt for it
		fmt.Printf("Enter passphrase for key '%s': ", privateKeyPath)
		passphrase, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read passphrase: %w", err)
		}
		fmt.Println() // Add newline after passphrase input

		// Try again with the passphrase
		if err := encryptor.AddPrivateKeyFromFile(privateKeyPath, passphrase); err != nil {
			return fmt.Errorf("failed to load private key with passphrase: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	// Store the encryptor in the command context
	ctx := context.WithValue(cmd.Context(), "encryptor", encryptor)
	cmd.SetContext(ctx)

	return nil
}

// isPassphraseError checks if an error is due to a missing passphrase
func isPassphraseError(err error) bool {
	return err != nil && (err.Error() == "ssh: this private key is passphrase protected" ||
		err.Error() == "failed to parse private key: ssh: this private key is passphrase protected")
}

// getStore gets the storage from command context
func getStore(cmd *cobra.Command) (*storage.Store, error) {
	storeDir, _ := cmd.Flags().GetString("store")
	encryptor := cmd.Context().Value("encryptor").(crypto.Encryptor)

	return storage.NewStore(storeDir, encryptor)
}
