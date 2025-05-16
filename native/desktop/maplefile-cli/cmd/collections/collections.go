// monorepo/native/desktop/maplefile-cli/cmd/collections/collections.go
package collections

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	collectionService "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
)

func CollectionsCmd(collectionSvc collectionService.CollectionService, logger *zap.Logger) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "collections",
		Short: "Manage collections of files",
		Long:  `Create and manage collections of encrypted files.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add collection subcommands
	cmd.AddCommand(createRootCollectionCmd(collectionSvc, logger))
	cmd.AddCommand(createSubCollectionCmd(collectionSvc, logger))

	return cmd
}
