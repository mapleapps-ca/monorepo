// native/desktop/maplefile-cli/cmd/sync/full.go
package sync

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
)

// fullCmd creates a command for full synchronization
func fullCmd(
	syncService svc_sync.SyncService,
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "full",
		Short: "Perform full synchronization of collections and files",
		Long: `Perform a comprehensive synchronization of both collections and files
between local storage and the cloud backend.

This command will:
- Sync all collections from the cloud backend
- Sync all file metadata from the cloud backend
- Create, update, or delete items as needed
- Update the local sync state for both collections and files

This is equivalent to running both 'sync collections' and 'sync files' commands.

The sync process is incremental, only processing changes since the last sync.

Examples:
  # Perform full synchronization
  maplefile-cli sync full`,
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()

			fmt.Println("🔄 Starting full synchronization...")
			fmt.Println("📡 Connecting to cloud backend...")

			// Execute full sync
			result, err := syncService.FullSync(cmd.Context())
			if err != nil {
				fmt.Printf("❌ Full sync failed: %v\n", err)
				return
			}

			// Display results
			duration := time.Since(startTime)

			fmt.Println("\n✅ Full synchronization completed!")
			fmt.Printf("⏱️  Duration: %v\n", duration.Round(time.Millisecond))

			// Collections summary
			if result.CollectionsProcessed > 0 || result.FilesProcessed > 0 {
				fmt.Printf("📊 Summary:\n")

				if result.CollectionsProcessed > 0 {
					fmt.Printf("\n📁 Collections:\n")
					fmt.Printf("   • Processed: %d\n", result.CollectionsProcessed)
					if result.CollectionsAdded > 0 {
						fmt.Printf("   • ➕ Added: %d\n", result.CollectionsAdded)
					}
					if result.CollectionsUpdated > 0 {
						fmt.Printf("   • 🔄 Updated: %d\n", result.CollectionsUpdated)
					}
					if result.CollectionsDeleted > 0 {
						fmt.Printf("   • 🗑️  Deleted: %d\n", result.CollectionsDeleted)
					}
				}

				if result.FilesProcessed > 0 {
					fmt.Printf("\n📄 Files:\n")
					fmt.Printf("   • Processed: %d\n", result.FilesProcessed)
					if result.FilesAdded > 0 {
						fmt.Printf("   • ➕ Added: %d\n", result.FilesAdded)
					}
					if result.FilesUpdated > 0 {
						fmt.Printf("   • 🔄 Updated: %d\n", result.FilesUpdated)
					}
					if result.FilesDeleted > 0 {
						fmt.Printf("   • 🗑️  Deleted: %d\n", result.FilesDeleted)
					}
				}
			}

			if len(result.Errors) > 0 {
				fmt.Printf("\n⚠️  Errors encountered during sync (%d total):\n", len(result.Errors))
				for i, errMsg := range result.Errors {
					if i < 10 { // Show first 10 errors
						fmt.Printf("   %d. %s\n", i+1, errMsg)
					}
				}
				if len(result.Errors) > 10 {
					fmt.Printf("   ... and %d more errors\n", len(result.Errors)-10)
				}
			}

			totalProcessed := result.CollectionsProcessed + result.FilesProcessed
			if totalProcessed == 0 {
				fmt.Println("ℹ️  No changes found - everything is already up to date!")
			} else {
				fmt.Println("\n🎉 Your data is now fully synchronized!")
				if result.FilesProcessed > 0 {
					fmt.Println("💡 Use 'maplefile-cli filesync onload' to download file content locally.")
				}
			}
		},
	}

	return cmd
}
