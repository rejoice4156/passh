package main

import (
	"fmt"
	"os"

	"github.com/rejoice4156/passh/pkg/cli"
)

func main() {
	rootCmd := cli.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		// Simply use fmt.Println instead of fmt.Fprintf to avoid potential stderr issues
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
