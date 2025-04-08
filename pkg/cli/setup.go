package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Set up passh environment",
		Long:  "Check and set up the environment needed for passh including SSH keys and agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup()
		},
	}
}

func runSetup() error {
	fmt.Println("🔑 Passh Setup Wizard")
	fmt.Println("=====================")

	// 1. Check for SSH installation
	fmt.Print("Checking for SSH installation... ")
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		fmt.Println("❌ Not Found")
		fmt.Println("\nSSH is required but not installed or not in your PATH.")
		fmt.Println("Please install SSH with one of these commands:")
		fmt.Println("  - On Debian/Ubuntu: sudo apt-get install openssh-client")
		fmt.Println("  - On Fedora/RHEL: sudo dnf install openssh-clients")
		fmt.Println("  - On macOS: brew install openssh")
		fmt.Println("  - On Windows: Install Git for Windows or OpenSSH via Windows Optional Features")
		return fmt.Errorf("SSH not installed")
	}
	fmt.Printf("✅ Found (%s)\n", sshPath)

	// 2. Check for existing keys
	fmt.Print("Checking for existing SSH keys... ")
	sshDir := filepath.Join(os.Getenv("HOME"), ".ssh")

	keyTypes := []struct {
		name    string
		private string
		public  string
	}{
		{"Ed25519", "id_ed25519", "id_ed25519.pub"},
		{"ECDSA", "id_ecdsa", "id_ecdsa.pub"},
		{"RSA", "id_rsa", "id_rsa.pub"},
	}

	foundKeys := false
	var foundKeyType string
	var foundKeyPath string

	for _, keyType := range keyTypes {
		privatePath := filepath.Join(sshDir, keyType.private)
		if _, err := os.Stat(privatePath); err == nil {
			foundKeys = true
			foundKeyType = keyType.name
			foundKeyPath = privatePath
			break
		}
	}

	if foundKeys {
		fmt.Printf("✅ Found %s key (%s)\n", foundKeyType, foundKeyPath)
	} else {
		fmt.Println("❌ Not Found")
		fmt.Print("\nWould you like to generate a new Ed25519 SSH key? [y/N]: ")
		var userResponse string
		if _, err := fmt.Scanln(&userResponse); err != nil {
			// Handle specific error for unexpected newline
			if err.Error() != "unexpected newline" {
				fmt.Printf("Error reading input: %v\n", err)
			}
			// Default to "n" for empty or error
			userResponse = "n"
		}

		if strings.ToLower(userResponse) == "y" || strings.ToLower(userResponse) == "yes" {
			// Ensure SSH directory exists
			if err := os.MkdirAll(sshDir, 0700); err != nil {
				return fmt.Errorf("failed to create SSH directory: %w", err)
			}

			// Generate a new Ed25519 key
			fmt.Println("Generating new Ed25519 key...")
			cmd := exec.Command("ssh-keygen", "-t", "ed25519")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to generate SSH key: %w", err)
			}
			foundKeys = true
			fmt.Println("✅ Key generated successfully")
		} else {
			fmt.Println("\nPlease generate SSH keys manually with 'ssh-keygen -t ed25519' before using passh.")
			return fmt.Errorf("no SSH keys available")
		}
	}

	// 3. Check for SSH agent
	fmt.Print("Checking for SSH agent... ")
	agentSock := os.Getenv("SSH_AUTH_SOCK")
	if agentSock == "" {
		fmt.Println("❌ Not Running")
		fmt.Print("\nWould you like to start the SSH agent? [y/N]: ")
		var userResponse string
		if _, err := fmt.Scanln(&userResponse); err != nil {
			if err.Error() != "unexpected newline" {
				fmt.Printf("Error reading input: %v\n", err)
			}
			// Default to "n" for empty or error
			userResponse = "n"
		}

		if strings.ToLower(userResponse) == "y" || strings.ToLower(userResponse) == "yes" {
			fmt.Println("\nStarting SSH agent...")
			fmt.Println("Please run these commands in your shell:")
			fmt.Println("  eval `ssh-agent`")
			fmt.Println("  ssh-add")
			fmt.Println("\nAfter starting the agent, run passh again.")
		} else {
			fmt.Println("\nWithout the SSH agent, you'll need to enter your key passphrase each time.")
		}
	} else {
		fmt.Printf("✅ Running (Socket: %s)\n", agentSock)

		// Check if keys are added to agent
		fmt.Print("Checking if keys are added to agent... ")
		cmd := exec.Command("ssh-add", "-l")
		output, err := cmd.CombinedOutput()
		if err != nil || string(output) == "The agent has no identities.\n" {
			fmt.Println("❌ No keys added")
			fmt.Print("\nWould you like to add your key to the SSH agent? [y/N]: ")
			var userResponse string
			if _, err := fmt.Scanln(&userResponse); err != nil {
				if err.Error() != "unexpected newline" {
					fmt.Printf("Error reading input: %v\n", err)
				}
				// Default to "n" for empty or error
				userResponse = "n"
			}

			if strings.ToLower(userResponse) == "y" || strings.ToLower(userResponse) == "yes" {
				cmd := exec.Command("ssh-add")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to add key to SSH agent: %w", err)
				}
				fmt.Println("✅ Key added to agent")
			}
		} else {
			fmt.Println("✅ Keys are present in agent")
		}
	}

	fmt.Println("✅ Passh setup complete!")
	fmt.Println("You can now use passh to securely store and retrieve passwords.")
	fmt.Println("Try: passh add example/password")

	return nil
}
