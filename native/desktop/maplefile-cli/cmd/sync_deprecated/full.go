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
	syncFullService svc_sync.SyncFullService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionBatchSize int64
	var fileBatchSize int64
	var maxBatches int
	var password string

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

Note: This command syncs file metadata only. The actual file content remains
in the cloud until you explicitly download (onload) files.

Examples:
  # Perform full synchronization with default settings
  maplefile-cli sync full --password mypassword

  # Full sync with custom batch sizes
  maplefile-cli sync full --collection-batch-size 25 --file-batch-size 30 --password mypassword

  # Full sync with limited batches
  maplefile-cli sync full --max-batches 50 --password mypassword`,
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			fmt.Println("🔄 Starting full synchronization...")
			fmt.Println("📡 Connecting to cloud backend...")

			// Create input for sync service
			input := &svc_sync.FullSyncInput{
				CollectionBatchSize: collectionBatchSize,
				FileBatchSize:       fileBatchSize,
				MaxBatches:          maxBatches,
				Password:            password,
			}

			// Execute full sync
			result, err := syncFullService.Execute(cmd.Context(), input)
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

	// Add command flags
	cmd.Flags().Int64Var(&collectionBatchSize, "collection-batch-size", 50, "Number of collections to process per batch")
	cmd.Flags().Int64Var(&fileBatchSize, "file-batch-size", 50, "Number of files to process per batch")
	cmd.Flags().IntVar(&maxBatches, "max-batches", 100, "Maximum number of batches to process")
	cmd.Flags().StringVarP(&password, "password", "", "", "User password for E2EE operations")

	return cmd
}
