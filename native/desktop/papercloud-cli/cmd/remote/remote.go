package remote

import (
	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
)

func RemoteCmd(configService config.ConfigService) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "remote",
		Short: "Execute commands related to making remote API calls",
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add Remote-related commands
	cmd.AddCommand(HealthCheckCmd(configService))
	cmd.AddCommand(RegisterUserCmd(configService))
	cmd.AddCommand(VerifyEmailCmd(configService))

	return cmd
}
