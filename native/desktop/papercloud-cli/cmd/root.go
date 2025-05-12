package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/version"
)

var rootCmd = &cobra.Command{
	Use:   "papercloud-cli",
	Short: "PaperCloud CLI",
	Long:  `PaperCloud Command Line Interface`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do nothing.
	},
}

func Execute() {
	// Attach sub-commands to our main root.
	rootCmd.AddCommand(version.VersionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
