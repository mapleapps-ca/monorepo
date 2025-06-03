// cmd/collections/collections.go
package collections

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/collections/share"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsharing"
)

func CollectionsCmd(
	createService collection.CreateService,
	listService collection.ListService,
	softDeleteService collection.SoftDeleteService,
	sharingService collectionsharing.CollectionSharingService,
	getMembersService collectionsharing.CollectionSharingGetMembersService,
	listSharedService collectionsharing.ListSharedCollectionsService,
	removeMemberService collectionsharing.CollectionSharingRemoveMembersService,

	synchronizedSharingService collectionsharing.SynchronizedCollectionSharingService,
	originalSharingService collectionsharing.CollectionSharingService,
	logger *zap.Logger,
) *cobra.Command {
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
	cmd.AddCommand(createRootCollectionCmd(createService, logger))
	cmd.AddCommand(createSubCollectionCmd(createService, logger))
	cmd.AddCommand(listCollectionsCmd(listService, logger))
	cmd.AddCommand(softDeleteCmd(softDeleteService, logger))
	cmd.AddCommand(archiveCmd(softDeleteService, logger))
	cmd.AddCommand(restoreCmd(softDeleteService, logger))
	cmd.AddCommand(listByStateCmd(listService, logger))

	// Add sharing subcommands
	cmd.AddCommand(share.ShareCmdWithSync(synchronizedSharingService, originalSharingService, logger))
	cmd.AddCommand(share.UnshareCmd(removeMemberService, logger))
	cmd.AddCommand(share.MembersCmd(getMembersService, logger))
	cmd.AddCommand(share.ListSharedCmd(listSharedService, logger))

	return cmd
}
