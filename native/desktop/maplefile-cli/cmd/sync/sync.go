// native/desktop/maplefile-cli/cmd/sync/sync.go
package sync

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
)

// SyncCmd creates a command for sync operations
func SyncCmd(
	syncCollectionService svc_sync.SyncCollectionService,
	syncFileService svc_sync.SyncFileService,
	syncFullService svc_sync.SyncFullService,
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sync",
		Short: "Synchronize data between local storage and cloud backend",
		Long: `Synchronize collections and files between local storage and the cloud backend.

This command provides several subcommands for different types of synchronization:
- collections: Sync only collections
- files: Sync only file metadata
- full: Sync both collections and files
- reset: Reset the sync state to start fresh

Examples:
  # Sync all data (collections and files)
  maplefile-cli sync full

  # Sync only collections
  maplefile-cli sync collections

  # Sync only file metadata
  maplefile-cli sync files

  # Reset sync state
  maplefile-cli sync reset`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add sync subcommands
	cmd.AddCommand(collectionsCmd(syncCollectionService, logger))
	cmd.AddCommand(filesCmd(syncFileService, logger))
	cmd.AddCommand(fullCmd(syncFullService, logger))
	// cmd.AddCommand(resetCmd(syncService, logger))

	return cmd
}
