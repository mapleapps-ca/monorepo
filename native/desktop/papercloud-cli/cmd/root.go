// Package cmd provides the CLI commands
package cmd

import (
	"github.com/spf13/cobra"

	config_cmd "github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/remote"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/verifyemail"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/version"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
)

// NewRootCmd creates a new root command with all dependencies injected
func NewRootCmd(configService config.ConfigService) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "papercloud-cli",
		Short: "PaperCloud CLI",
		Long:  `PaperCloud Command Line Interface`,
		Run: func(cmd *cobra.Command, args []string) {
			// Root command does nothing by default
			cmd.Help()
		},
	}

	// Attach sub-commands to our main root
	rootCmd.AddCommand(version.VersionCmd())
	rootCmd.AddCommand(config_cmd.ConfigCmd(configService))
	rootCmd.AddCommand(remote.RemoteCmd(configService))
	rootCmd.AddCommand(register.RegisterCmd(configService))
	rootCmd.AddCommand(verifyemail.VerifyEmailCmd(configService))

	return rootCmd
}
