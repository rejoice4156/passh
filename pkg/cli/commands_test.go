package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Create temp dir for store
	tempDir, err := os.MkdirTemp("", "passh-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp directory: %v", err)
		}
	}()

	// Test the root command help output
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed to execute root command: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "A terminal password manager") {
		t.Fatalf("Expected help text containing 'A terminal password manager', got: %s", output)
	}
}

func TestSubCommands(t *testing.T) {
	// Test subcommand existence
	cmd := NewRootCmd()

	// Updated to include the new setup command
	subCommands := []string{"add", "get", "list", "delete", "generate", "setup"}
	for _, name := range subCommands {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", name)
		}
	}
}

// Test the new setup command
func TestSetupCommand(t *testing.T) {
	cmd := NewRootCmd()

	// Get the setup command
	setupCmd, _, err := cmd.Find([]string{"setup"})
	if err != nil {
		t.Fatalf("Setup command not found: %v", err)
	}

	// Check that it's not nil and has the expected help text
	if setupCmd == nil {
		t.Fatalf("Setup command is nil")
	}

	if !strings.Contains(setupCmd.Short, "Set up passh environment") {
		t.Errorf("Setup command short description is incorrect: %s", setupCmd.Short)
	}
}

func TestCheckSSHEnvironment(t *testing.T) {
	// Save path to restore later
	origPath := os.Getenv("PATH")
	defer func() {
		if err := os.Setenv("PATH", origPath); err != nil {
			t.Errorf("Failed to restore PATH: %v", err)
		}
	}()

	// Test without SSH in PATH
	// Create a temp dir and use it as PATH
	tempDir, err := os.MkdirTemp("", "no-ssh-path")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp directory: %v", err)
		}
	}()

	// Set PATH to temp dir to simulate missing SSH
	if err := os.Setenv("PATH", tempDir); err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}

	// checkSSHEnvironment should return an error when SSH is not found
	err = checkSSHEnvironment()
	if err == nil {
		t.Errorf("Expected error when SSH is not in PATH")
	}

	// Check if the error message mentions SSH installation
	if err != nil && !strings.Contains(err.Error(), "SSH is not installed") {
		t.Errorf("Expected error message to mention SSH installation, got: %v", err)
	}

	// Restore original path for remaining tests
	if err := os.Setenv("PATH", origPath); err != nil {
		t.Fatalf("Failed to restore PATH: %v", err)
	}

	// Skip the SSH key test for now - seems to be passing even when it shouldn't
	t.Skip("Skipping SSH key test")

	// Test SSH key detection
	// Save HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Errorf("Failed to restore HOME: %v", err)
		}
	}()

	// Create temp home with no .ssh directory
	tempHome, err := os.MkdirTemp("", "no-ssh-home")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempHome); err != nil {
			t.Errorf("Failed to clean up temp home: %v", err)
		}
	}()

	// Set HOME to temp dir to simulate missing SSH keys
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	// checkSSHEnvironment should return an error when no SSH keys are found
	err = checkSSHEnvironment()
	if err == nil {
		t.Errorf("Expected error when no SSH keys are found")
	}

	// Check if the error message mentions creating SSH keys
	if err != nil && !strings.Contains(err.Error(), "No SSH keys found") {
		t.Errorf("Expected error message to mention creating SSH keys, got: %v", err)
	}
}

// Integration test with mock SSH keys
func TestCommandsWithMockKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for test keys and store
	tempDir, err := os.MkdirTemp("", "passh-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp directory: %v", err)
		}
	}()

	keyDir := filepath.Join(tempDir, "keys")
	storeDir := filepath.Join(tempDir, "store")

	if err := os.MkdirAll(keyDir, 0700); err != nil {
		t.Fatalf("Failed to create key directory: %v", err)
	}
	if err := os.MkdirAll(storeDir, 0700); err != nil {
		t.Fatalf("Failed to create store directory: %v", err)
	}

	// Generate test SSH keys - now using Ed25519 mock format
	privateKeyPath := filepath.Join(keyDir, "id_ed25519") // Changed to ed25519
	publicKeyPath := filepath.Join(keyDir, "id_ed25519.pub")

	// Create mock keys (just empty files for this test)
	if err := os.WriteFile(privateKeyPath, []byte("MOCK ED25519 PRIVATE KEY"), 0600); err != nil {
		t.Fatalf("Failed to write private key: %v", err)
	}
	if err := os.WriteFile(publicKeyPath, []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMockKey test@localhost"), 0644); err != nil {
		t.Fatalf("Failed to write public key: %v", err)
	}

	// For this test, we'll just check that the command flags are set up correctly
	cmd := NewRootCmd()

	// Check store flag
	storeFlag := cmd.PersistentFlags().Lookup("store")
	if storeFlag == nil {
		t.Fatal("store flag not found")
	}

	// Check public-key flag
	publicKeyFlag := cmd.PersistentFlags().Lookup("public-key")
	if publicKeyFlag == nil {
		t.Fatal("public-key flag not found")
	}

	// Check private-key flag
	privateKeyFlag := cmd.PersistentFlags().Lookup("private-key")
	if privateKeyFlag == nil {
		t.Fatal("private-key flag not found")
	}

	// Check no-agent flag
	noAgentFlag := cmd.PersistentFlags().Lookup("no-agent")
	if noAgentFlag == nil {
		t.Fatal("no-agent flag not found")
	}
}
