// cmd/sync/main.go - Clean main sync command structure
package sync

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
)

// SyncCmd creates the main sync command with simplified structure
func SyncCmd(
	syncCollectionService svc_sync.SyncCollectionService,
	syncFileService svc_sync.SyncFileService,
	syncDebugService svc_sync.SyncDebugService,
	logger *zap.Logger,
) *cobra.Command {
	// Create the main sync command (unified)
	mainSyncCmd := syncCmd(syncCollectionService, syncFileService, logger)

	// Set up the parent command that can have subcommands
	var cmd = &cobra.Command{
		Use:   "sync",
		Short: "Synchronize data with cloud backend",
		Long: `
Synchronize your collections and files with the MapleFile cloud backend.

This command has two modes:

1. Direct sync (recommended):
   maplefile-cli sync [flags]

   Synchronizes both collections and files by default. Use --collections or --files
   to sync only specific types.

2. Debug mode:
   maplefile-cli sync debug [flags]

   Diagnoses sync issues and provides recommendations.

Examples:
  # Sync everything (recommended)
  maplefile-cli sync --password mypass

  # Sync only collections
  maplefile-cli sync --collections --password mypass

  # Sync only file metadata
  maplefile-cli sync --files --password mypass

  # Debug sync issues
  maplefile-cli sync debug --password mypass

  # Quick network check
  maplefile-cli sync debug --network

The sync process is incremental and only processes changes since the last sync.
`,
		Run: mainSyncCmd.Run, // Delegate to the main sync command by default
	}

	// Copy flags from main sync command to parent
	cmd.Flags().AddFlagSet(mainSyncCmd.Flags())

	// Add debug subcommand
	cmd.AddCommand(debugCmd(syncDebugService, logger))

	return cmd
}
