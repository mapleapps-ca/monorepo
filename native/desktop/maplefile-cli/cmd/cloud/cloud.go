// monorepo/native/desktop/maplefile-cli/cmd/cloud/cloud.go
package cloud

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
)

func CloudCmd(
	configService config.ConfigService,
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cloud",
		Short: "Execute commands related to making cloud API calls",
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add Cloud-related commands
	cmd.AddCommand(HealthCheckCmd(configService))
	cmd.AddCommand(PublicUserLookupCmd(configService))

	return cmd
}
