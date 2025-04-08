package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Passh - SSH-backed password manager")
			fmt.Println("Build date: 2025-04-08 11:32:27")
			fmt.Println("Author: rejoice4156")
			fmt.Println("License: MIT")
		},
	}
}
