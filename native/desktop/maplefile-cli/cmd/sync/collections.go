// native/desktop/maplefile-cli/cmd/sync/collections.go
package sync

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	// svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
)

// collectionsCmd creates a command for syncing collections
func collectionsCmd(
	// syncService svc_sync.SyncService,
	logger *zap.Logger,
) *cobra.Command {
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
  # Sync collections
  maplefile-cli sync collections`,
		Run: func(cmd *cobra.Command, args []string) {
			// startTime := time.Now()

			// fmt.Println("🔄 Starting collection synchronization...")
			// fmt.Println("📡 Connecting to cloud backend...")

			// // Execute collection sync
			// result, err := syncService.SyncCollections(cmd.Context())
			// if err != nil {
			// 	fmt.Printf("❌ Collection sync failed: %v\n", err)
			// 	return
			// }

			// // Display results
			// duration := time.Since(startTime)

			// fmt.Println("\n✅ Collection synchronization completed!")
			// fmt.Printf("⏱️  Duration: %v\n", duration.Round(time.Millisecond))
			// fmt.Printf("📊 Summary:\n")
			// fmt.Printf("   • Processed: %d collections\n", result.CollectionsProcessed)

			// if result.CollectionsAdded > 0 {
			// 	fmt.Printf("   • ➕ Added: %d collections\n", result.CollectionsAdded)
			// }

			// if result.CollectionsUpdated > 0 {
			// 	fmt.Printf("   • 🔄 Updated: %d collections\n", result.CollectionsUpdated)
			// }

			// if result.CollectionsDeleted > 0 {
			// 	fmt.Printf("   • 🗑️  Deleted: %d collections\n", result.CollectionsDeleted)
			// }

			// if len(result.Errors) > 0 {
			// 	fmt.Printf("   • ⚠️  Errors: %d\n", len(result.Errors))
			// 	fmt.Printf("\n⚠️  Errors encountered during sync:\n")
			// 	for i, errMsg := range result.Errors {
			// 		fmt.Printf("   %d. %s\n", i+1, errMsg)
			// 	}
			// }

			// if result.CollectionsProcessed == 0 {
			// 	fmt.Println("ℹ️  No collection changes found - already up to date!")
			// } else {
			// 	fmt.Println("\n🎉 Your collections are now synchronized!")
			// }
		},
	}

	return cmd
}
