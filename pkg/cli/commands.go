package cli

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Subcommands

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			
			fmt.Print("Enter password: ")
			password, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Println()
			
			store, err := getStore(cmd)
			if err != nil {
				return err
			}
			
			return store.Add(name, password)
		},
	}
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
	cmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			
			store, err := getStore(cmd)
			if err != nil {
				return err
			}
			
			return store.Delete(name)
		},
	}
	return cmd
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