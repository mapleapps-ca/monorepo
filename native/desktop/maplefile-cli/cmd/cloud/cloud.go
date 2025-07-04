// monorepo/native/desktop/maplefile-cli/cmd/cloud/cloud.go
package cloud

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	uc_publiclookupdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/publiclookupdto"
)

func CloudCmd(
	configService config.ConfigService,
	getPublicLookupFromCloudUseCase uc_publiclookupdto.GetPublicLookupFromCloudUseCase,
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
	cmd.AddCommand(PublicUserLookupCmd(configService, getPublicLookupFromCloudUseCase))

	return cmd
}
