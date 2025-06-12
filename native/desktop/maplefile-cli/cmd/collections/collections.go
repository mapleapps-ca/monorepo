// cmd/collections/collections.go - Clean main collections command
package collections

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/collections/share"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsharing"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsyncer"
)

func CollectionsCmd(
	createService collection.CreateService,
	listService collection.ListService,
	softDeleteService collection.SoftDeleteService,
	listFromCloudService collectionsyncer.ListFromCloudService,
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
		Long: `
Create and manage collections of encrypted files.

Collections are secure containers for organizing your files. You can create
root-level collections or sub-collections within existing collections.

Available commands:
  create    Create new collections (root or sub-collections)
  list      List collections with various filters
  delete    Delete or archive collections (can be restored)
  restore   Restore deleted/archived collections
  share     Share collections with other users

Examples:
  # Create a new collection
  maplefile-cli collections create "My Documents" --password PASSWORD

  # List all collections
  maplefile-cli collections list

  # Create a sub-collection
  maplefile-cli collections create "Project Files" --parent PARENT_ID --password PASSWORD

  # Delete a collection
  maplefile-cli collections delete COLLECTION_ID

  # Share a collection
  maplefile-cli collections share COLLECTION_ID --email user@example.com --permission read_write --password PASSWORD

For detailed help: maplefile-cli collections COMMAND --help
`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Core collection management commands
	cmd.AddCommand(createCmd(createService, logger))
	cmd.AddCommand(listCmd(listService, logger))
	cmd.AddCommand(deleteCmd(softDeleteService, logger))
	cmd.AddCommand(restoreCmd(softDeleteService, logger))

	// Sharing commands (keep as-is - well designed)
	cmd.AddCommand(share.ShareCmdWithSync(synchronizedSharingService, originalSharingService, logger))
	cmd.AddCommand(share.UnshareCmd(removeMemberService, logger))
	cmd.AddCommand(share.MembersCmd(getMembersService, logger))
	cmd.AddCommand(share.ListSharedCmd(listSharedService, logger))

	return cmd
}
