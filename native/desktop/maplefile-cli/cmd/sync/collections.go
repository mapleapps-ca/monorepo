// native/desktop/maplefile-cli/cmd/sync/collections.go
package sync

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
)

// collectionsCmd creates a command for syncing collections
func collectionsCmd(
	syncCollectionService svc_sync.SyncCollectionService,
	logger *zap.Logger,
) *cobra.Command {
	var batchSize int64
	var maxBatches int

	var cmd = &cobra.Command{
		Use:   "collections",
		Short: "Sync collections from cloud backend",
		Long: `Synchronize collections between local storage and the cloud backend.

This command will:
- Fetch collection changes from the cloud backend
- Create new collections that don't exist locally
- Update existing collections that have been modified on the server
- Delete collections that have been deleted on the server
- Update the local sync state

The sync process is incremental, only processing changes since the last sync.

Examples:
  # Sync collections with default settings
  maplefile-cli sync collections

  # Sync collections with custom batch size
  maplefile-cli sync collections --batch-size 25

  # Sync collections with limited batches
  maplefile-cli sync collections --max-batches 50`,
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()

			fmt.Println("🔄 Starting collection synchronization...")
			fmt.Println("📡 Connecting to cloud backend...")

			// Create input for sync service
			input := &svc_sync.SyncCollectionsInput{
				BatchSize:  batchSize,
				MaxBatches: maxBatches,
			}

			// Execute collection sync
			result, err := syncCollectionService.Execute(cmd.Context(), input)
			if err != nil {
				fmt.Printf("❌ Collection sync failed: %v\n", err)
				return
			}

			// Display results
			duration := time.Since(startTime)

			fmt.Println("\n✅ Collection synchronization completed!")
			fmt.Printf("⏱️  Duration: %v\n", duration.Round(time.Millisecond))
			fmt.Printf("📊 Summary:\n")
			fmt.Printf("   • Processed: %d collections\n", result.CollectionsProcessed)

			if result.CollectionsAdded > 0 {
				fmt.Printf("   • ➕ Added: %d collections\n", result.CollectionsAdded)
			}

			if result.CollectionsUpdated > 0 {
				fmt.Printf("   • 🔄 Updated: %d collections\n", result.CollectionsUpdated)
			}

			if result.CollectionsDeleted > 0 {
				fmt.Printf("   • 🗑️  Deleted: %d collections\n", result.CollectionsDeleted)
			}

			if len(result.Errors) > 0 {
				fmt.Printf("   • ⚠️  Errors: %d\n", len(result.Errors))
				fmt.Printf("\n⚠️  Errors encountered during sync:\n")
				for i, errMsg := range result.Errors {
					fmt.Printf("   %d. %s\n", i+1, errMsg)
				}
			}

			if result.CollectionsProcessed == 0 {
				fmt.Println("ℹ️  No collection changes found - already up to date!")
			} else {
				fmt.Println("\n🎉 Your collections are now synchronized!")
			}
		},
	}

	// Add command flags
	cmd.Flags().Int64Var(&batchSize, "batch-size", 50, "Number of collections to process per batch")
	cmd.Flags().IntVar(&maxBatches, "max-batches", 100, "Maximum number of batches to process")

	return cmd
}
