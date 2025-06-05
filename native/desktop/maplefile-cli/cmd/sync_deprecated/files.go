// native/desktop/maplefile-cli/cmd/sync/files.go
package sync

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
)

// filesCmd creates a command for syncing files
func filesCmd(
	syncService svc_sync.SyncFileService,
	logger *zap.Logger,
) *cobra.Command {
	var batchSize int64
	var maxBatches int
	var password string

	var cmd = &cobra.Command{
		Use:   "files",
		Short: "Sync files from cloud backend",
		Long: `Synchronize files between local storage and the cloud backend.

This command will:
- Fetch file changes from the cloud backend
- Create new file records that don't exist locally
- Update existing file records that have been modified on the server
- Delete file records that have been deleted on the server
- Update the local sync state

Note: This command syncs file metadata only. The actual file content remains
in the cloud until you explicitly download (onload) files.

The sync process is incremental, only processing changes since the last sync.

Examples:
  # Sync files with default settings
  maplefile-cli sync files

  # Sync files with custom batch size
  maplefile-cli sync files --batch-size 25

  # Sync files with limited batches
  maplefile-cli sync files --max-batches 50`,
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			fmt.Println("🔄 Starting file synchronization...")
			fmt.Println("📡 Connecting to cloud backend...")

			// Create input for sync service
			input := &svc_sync.SyncFilesInput{
				BatchSize:  batchSize,
				MaxBatches: maxBatches,
				Password:   password,
			}

			// Execute file sync
			result, err := syncService.Execute(cmd.Context(), input)
			if err != nil {
				fmt.Printf("❌ File sync failed: %v\n", err)
				return
			}

			// Display results
			duration := time.Since(startTime)

			fmt.Println("\n✅ File synchronization completed!")
			fmt.Printf("⏱️  Duration: %v\n", duration.Round(time.Millisecond))
			fmt.Printf("📊 Summary:\n")
			fmt.Printf("   • Processed: %d files\n", result.FilesProcessed)

			if result.FilesAdded > 0 {
				fmt.Printf("   • ➕ Added: %d files\n", result.FilesAdded)
			}

			if result.FilesUpdated > 0 {
				fmt.Printf("   • 🔄 Updated: %d files\n", result.FilesUpdated)
			}

			if result.FilesDeleted > 0 {
				fmt.Printf("   • 🗑️  Deleted: %d files\n", result.FilesDeleted)
			}

			if len(result.Errors) > 0 {
				fmt.Printf("   • ⚠️  Errors: %d\n", len(result.Errors))
				fmt.Printf("\n⚠️  Errors encountered during sync:\n")
				for i, errMsg := range result.Errors {
					fmt.Printf("   %d. %s\n", i+1, errMsg)
				}
			}

			if result.FilesProcessed == 0 {
				fmt.Println("ℹ️  No file changes found - already up to date!")
			} else {
				fmt.Println("\n🎉 Your file metadata is now synchronized!")
				fmt.Println("💡 Use 'maplefile-cli filesync onload' to download file content locally.")
			}
		},
	}

	// Add command flags
	cmd.Flags().Int64Var(&batchSize, "batch-size", 50, "Number of files to process per batch")
	cmd.Flags().IntVar(&maxBatches, "max-batches", 100, "Maximum number of batches to process")
	cmd.Flags().StringVarP(&password, "password", "", "", "User password for decrypting collection names")

	return cmd
}
