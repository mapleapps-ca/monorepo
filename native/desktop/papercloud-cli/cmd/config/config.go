// monorepo/native/desktop/papercloud-cli/cmd/config/config.go
package config

import (
	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
)

func ConfigCmd(configService config.ConfigService) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "Execute commands related to configuring the local papercloud-cli",
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add configuration-related commands
	cmd.AddCommand(getConfigCmd(configService))
	cmd.AddCommand(setConfigCmd(configService))

	return cmd
}
