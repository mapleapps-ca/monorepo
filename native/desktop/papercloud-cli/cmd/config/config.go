package config

import (
	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/config"
)

func ConfigCmd(configUseCase config.ConfigUseCase) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "Execute commands related to configuring the local papercloud-cli",
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add Remote-related commands
	cmd.AddCommand(getConfigCmd(configUseCase))
	cmd.AddCommand(setConfigCmd(configUseCase))

	return cmd
}
