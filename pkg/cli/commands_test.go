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
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {

		}
	}(tempDir)

	// Test the root command help output
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil { // Fixed: Previously unhandled error
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

	subCommands := []string{"add", "get", "list", "delete", "generate"}
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
		if err := os.RemoveAll(tempDir); err != nil { // Fixed: Previously unhandled error
			t.Errorf("Failed to remove temp directory: %v", err)
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

	// Generate test SSH keys (simplified for test)
	privateKeyPath := filepath.Join(keyDir, "id_test")
	publicKeyPath := filepath.Join(keyDir, "id_test.pub")

	// Create mock keys (just empty files for this test)
	if err := os.WriteFile(privateKeyPath, []byte("MOCK PRIVATE KEY"), 0600); err != nil {
		t.Fatalf("Failed to write private key: %v", err)
	}
	if err := os.WriteFile(publicKeyPath, []byte("MOCK PUBLIC KEY"), 0644); err != nil {
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
