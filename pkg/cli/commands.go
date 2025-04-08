package cli

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Subcommands

func newAddCmd() *cobra.Command {
	var generatePassword bool
	var passwordLength int

	cmd := &cobra.Command{
		Use:   "add NAME",
		Short: "Add a new password",
		Long:  "Add a new password entry to the store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getStore(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			var password []byte

			if generatePassword {
				// Generate a random password
				password, err = generateRandomPassword(passwordLength)
				if err != nil {
					return err
				}
				fmt.Printf("Generated password for '%s': %s\n", name, password)
			} else {
				// Read password from stdin with confirmation
				fmt.Printf("Enter password for '%s': ", name)
				password, err = term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				fmt.Println() // Add newline after password input

				// Ask for confirmation
				fmt.Print("Confirm password: ")
				confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("failed to read confirmation password: %w", err)
				}
				fmt.Println() // Add newline after confirmation input

				// Check if passwords match
				if string(password) != string(confirmPassword) {
					return fmt.Errorf("passwords do not match")
				}
			}

			// Add the password to the store
			if err := store.Add(name, password); err != nil {
				return err
			}

			fmt.Printf("Added password '%s'\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&generatePassword, "generate", "g", false, "Generate a random password")
	cmd.Flags().IntVarP(&passwordLength, "length", "l", 16, "Length of generated password")

	return cmd
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Retrieve a password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			store, err := getStore(cmd)
			if err != nil {
				return err
			}

			password, err := store.Get(name)
			if err != nil {
				return err
			}

			fmt.Println(string(password))
			return nil
		},
	}
	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all passwords",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getStore(cmd)
			if err != nil {
				return err
			}

			entries, err := store.List()
			if err != nil {
				return err
			}

			for _, entry := range entries {
				fmt.Println(entry)
			}
			return nil
		},
	}
	return cmd
}

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a password",
		Long:  "Delete a stored password entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getStore(cmd)
			if err != nil {
				return err
			}

			name := args[0]

			// Check if password exists first
			_, err = store.Get(name)
			if err != nil {
				return fmt.Errorf("password '%s' not found: %w", name, err)
			}

			// Ask for confirmation before deleting
			fmt.Printf("Are you sure you want to delete password '%s'? (y/N): ", name)
			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				if err.Error() != "unexpected newline" {
					fmt.Printf("Error reading input: %v\n", err)
				}
				// Default to "n" for empty or error
				response = "n"
			}

			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Deletion cancelled")
				return nil
			}

			// Proceed with deletion
			if err := store.Delete(name); err != nil {
				return err
			}

			fmt.Printf("Deleted password '%s'\n", name)
			return nil
		},
	}
}

func newGenerateCmd() *cobra.Command {
	var length int
	var noSymbols bool

	cmd := &cobra.Command{
		Use:   "generate [name]",
		Short: "Generate a password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Character sets for password generation
			lowerChars := "abcdefghijklmnopqrstuvwxyz"
			upperChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
			numberChars := "0123456789"
			symbolChars := "!@#$%^&*()-_=+[]{}|;:,.<>?"

			charset := lowerChars + upperChars + numberChars
			if !noSymbols {
				charset += symbolChars
			}

			// Generate the password
			passwordBytes := make([]byte, length)
			for i := 0; i < length; i++ {
				n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
				if err != nil {
					return fmt.Errorf("failed to generate random number: %w", err)
				}
				passwordBytes[i] = charset[n.Int64()]
			}
			password := []byte(string(passwordBytes))

			// Save the password
			store, err := getStore(cmd)
			if err != nil {
				return err
			}

			if err := store.Add(name, password); err != nil {
				return err
			}

			fmt.Println(string(password))
			return nil
		},
	}

	cmd.Flags().IntVarP(&length, "length", "l", 16, "Password length")
	cmd.Flags().BoolVarP(&noSymbols, "no-symbols", "n", false, "Don't include symbols in the password")

	return cmd
}
