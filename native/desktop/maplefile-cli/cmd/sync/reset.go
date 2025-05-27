// native/desktop/maplefile-cli/cmd/sync/reset.go
package sync

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
)

// resetCmd creates a command for resetting sync state
func resetCmd(
	syncService svc_sync.SyncService,
	logger *zap.Logger,
) *cobra.Command {
	var force bool

	var cmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset the synchronization state",
		Long: `Reset the local synchronization state to start fresh.

This command will:
- Clear all sync state timestamps
- Clear sync cursors and tracking information
- Force the next sync to be a full synchronization from the beginning

This is useful if:
- You suspect sync state corruption
- You want to force a complete re-sync
- You're troubleshooting sync issues

⚠️  Warning: This will cause the next sync to process all data from the server,
which may take longer than usual for large datasets.

Examples:
  # Reset sync state (with confirmation)
  maplefile-cli sync reset

  # Reset sync state without confirmation
  maplefile-cli sync reset --force`,
		Run: func(cmd *cobra.Command, args []string) {
			if !force {
				fmt.Println("⚠️  This will reset your sync state and force a full re-sync on the next run.")
				fmt.Print("Are you sure you want to continue? (y/N): ")

				var response string
				fmt.Scanln(&response)

				if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
					fmt.Println("❌ Reset cancelled.")
					return
				}
			}

			fmt.Println("🔄 Resetting sync state...")

			// Reset sync state
			err := syncService.ResetSync(cmd.Context())
			if err != nil {
				fmt.Printf("❌ Failed to reset sync state: %v\n", err)
				return
			}

			fmt.Println("✅ Sync state has been reset successfully!")
			fmt.Println("💡 The next sync will be a full synchronization from the beginning.")
			fmt.Println("💡 Run 'maplefile-cli sync full' to perform a complete sync.")
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
