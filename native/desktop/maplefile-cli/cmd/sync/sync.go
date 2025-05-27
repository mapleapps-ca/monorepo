// native/desktop/maplefile-cli/cmd/sync/sync.go
package sync

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	// svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
)

// SyncCmd creates a command for sync operations
func SyncCmd(
	// syncService svc_sync.SyncService,
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sync",
		Short: "Synchronize data between local storage and cloud backend",
		Long: `Synchronize collections and files between local storage and the cloud backend.

This command provides several subcommands for different types of synchronization:
- collections: Sync only collections
- files: Sync only files
- full: Sync both collections and files
- reset: Reset the sync state to start fresh

Examples:
  # Sync all data (collections and files)
  maplefile-cli sync full

  # Sync only collections
  maplefile-cli sync collections

  # Sync only files
  maplefile-cli sync files

  # Reset sync state
  maplefile-cli sync reset`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add sync subcommands
	cmd.AddCommand(collectionsCmd(logger))
	// cmd.AddCommand(filesCmd(syncService, logger))
	// cmd.AddCommand(fullCmd(syncService, logger))
	// cmd.AddCommand(resetCmd(syncService, logger))

	return cmd
}
