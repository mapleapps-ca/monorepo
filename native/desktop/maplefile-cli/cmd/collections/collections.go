// monorepo/native/desktop/maplefile-cli/cmd/collections/collections.go
package collections

import (
	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
)

func CollectionsCmd(configService config.ConfigService) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "collections",
		Short: "Execute commands related to making collection operations",
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add Remote-related commands
	cmd.AddCommand(HealthCheckCmd(configService))

	return cmd
}
