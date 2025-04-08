package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Use the provided date and user
	fmt.Println("Running passh test suite...")
	fmt.Println("Date: 2025-04-08 11:14:19")

	// Get username, with fallback if environment variable is missing
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") // For Windows
		if username == "" {
			username = "rejoice4156" // Fallback to provided username
		}
	}
	fmt.Printf("User: %s\n", username)

	// Get the project root directory
	root, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Navigate to project root if script is run from scripts directory
	if filepath.Base(root) == "scripts" {
		root = filepath.Dir(root)
	}

	// Run tests for each package
	packages := []string{
		"./pkg/crypto",
		"./pkg/storage",
		"./pkg/cli",
	}

	allPassed := true
	for _, pkg := range packages {
		fmt.Printf("\n=== Testing %s ===\n", pkg)
		cmd := exec.Command("go", "test", "-v", pkg)
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Tests failed for %s: %v\n", pkg, err)
			allPassed = false
		}
	}

	fmt.Println("\n=== Test Summary ===")
	if allPassed {
		fmt.Println("All tests passed!")
		os.Exit(0)
	} else {
		fmt.Println("Some tests failed!")
		os.Exit(1)
	}
}
