// monorepo/native/desktop/maplefile-cli/cmd/remote/remote.go
package remote

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/remotecollection"
)

func RemoteCmd(
	configService config.ConfigService,
	remoteListService remotecollection.ListService,
	logger *zap.Logger,
) *cobra.Command {
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
	cmd.AddCommand(RemoteListCollectionsCmd(configService, remoteListService, logger))

	return cmd
}
