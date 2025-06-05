// cmd/sync/sync.go - Clean unified sync command
package sync

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	dom_syncdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
)

// syncCmd creates a unified command for synchronizing data
func syncCmd(
	syncCollectionService svc_sync.SyncCollectionService,
	syncFileService svc_sync.SyncFileService,
	logger *zap.Logger,
) *cobra.Command {
	var collections bool
	var files bool
	var collectionBatchSize int64
	var fileBatchSize int64
	var maxBatches int
	var password string

	var cmd = &cobra.Command{
		Use:   "sync",
		Short: "Synchronize with cloud backend",
		Long: `
Synchronize your collections and files with the MapleFile cloud backend.

By default, syncs both collections and files. Use flags to sync only specific types:
  --collections    Sync only collections
  --files          Sync only file metadata

The sync process is incremental, only processing changes since the last sync.
File content remains in the cloud until explicitly downloaded.

Examples:
  # Sync everything (collections + files) - recommended
  maplefile-cli sync --password mypass

  # Sync only collections
  maplefile-cli sync --collections --password mypass

  # Sync only file metadata
  maplefile-cli sync --files --password mypass

  # Sync both explicitly (same as default)
  maplefile-cli sync --collections --files --password mypass

  # Custom batch sizes for large datasets
  maplefile-cli sync --collection-batch-size 25 --file-batch-size 30 --password mypass
`,
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Determine what to sync
			syncCollections := collections
			syncFiles := files

			// If no specific flags are set, sync both (default behavior)
			if !collections && !files {
				syncCollections = true
				syncFiles = true
			}

			fmt.Println("🔄 Starting synchronization...")
			fmt.Println("📡 Connecting to cloud backend...")

			var totalErrors []string
			var collectionsResult *dom_syncdto.SyncResult
			var filesResult *dom_syncdto.SyncResult

			// Sync collections if requested
			if syncCollections {
				fmt.Println("\n📁 Synchronizing collections...")

				collectionInput := &svc_sync.SyncCollectionsInput{
					BatchSize:  collectionBatchSize,
					MaxBatches: maxBatches,
					Password:   password,
				}

				var err error
				collectionsResult, err = syncCollectionService.Execute(cmd.Context(), collectionInput)
				if err != nil {
					fmt.Printf("❌ Collection sync failed: %v\n", err)
					totalErrors = append(totalErrors, fmt.Sprintf("Collections: %v", err))
				} else {
					fmt.Printf("✅ Collections synchronized!\n")
					fmt.Printf("   • Processed: %d collections\n", collectionsResult.CollectionsProcessed)
					if collectionsResult.CollectionsAdded > 0 {
						fmt.Printf("   • ➕ Added: %d\n", collectionsResult.CollectionsAdded)
					}
					if collectionsResult.CollectionsUpdated > 0 {
						fmt.Printf("   • 🔄 Updated: %d\n", collectionsResult.CollectionsUpdated)
					}
					if collectionsResult.CollectionsDeleted > 0 {
						fmt.Printf("   • 🗑️  Deleted: %d\n", collectionsResult.CollectionsDeleted)
					}

					if len(collectionsResult.Errors) > 0 {
						fmt.Printf("   • ⚠️  Errors: %d\n", len(collectionsResult.Errors))
						totalErrors = append(totalErrors, collectionsResult.Errors...)
					}
				}
			}

			// Sync files if requested
			if syncFiles {
				fmt.Println("\n📄 Synchronizing file metadata...")

				fileInput := &svc_sync.SyncFilesInput{
					BatchSize:  fileBatchSize,
					MaxBatches: maxBatches,
					Password:   password,
				}

				var err error
				filesResult, err = syncFileService.Execute(cmd.Context(), fileInput)
				if err != nil {
					fmt.Printf("❌ File sync failed: %v\n", err)
					totalErrors = append(totalErrors, fmt.Sprintf("Files: %v", err))
				} else {
					fmt.Printf("✅ File metadata synchronized!\n")
					fmt.Printf("   • Processed: %d files\n", filesResult.FilesProcessed)
					if filesResult.FilesAdded > 0 {
						fmt.Printf("   • ➕ Added: %d\n", filesResult.FilesAdded)
					}
					if filesResult.FilesUpdated > 0 {
						fmt.Printf("   • 🔄 Updated: %d\n", filesResult.FilesUpdated)
					}
					if filesResult.FilesDeleted > 0 {
						fmt.Printf("   • 🗑️  Deleted: %d\n", filesResult.FilesDeleted)
					}

					if len(filesResult.Errors) > 0 {
						fmt.Printf("   • ⚠️  Errors: %d\n", len(filesResult.Errors))
						totalErrors = append(totalErrors, filesResult.Errors...)
					}
				}
			}

			// Show final results
			duration := time.Since(startTime)
			fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")

			if len(totalErrors) > 0 {
				fmt.Printf("⚠️  Synchronization completed with %d error(s):\n", len(totalErrors))
				for i, err := range totalErrors {
					if i < 5 { // Show first 5 errors
						fmt.Printf("   %d. %s\n", i+1, err)
					}
				}
				if len(totalErrors) > 5 {
					fmt.Printf("   ... and %d more errors\n", len(totalErrors)-5)
				}
			} else {
				fmt.Printf("✅ Synchronization completed successfully!\n")
			}

			fmt.Printf("⏱️  Duration: %v\n", duration.Round(time.Millisecond))

			// Summary
			totalProcessed := 0
			if collectionsResult != nil {
				totalProcessed += collectionsResult.CollectionsProcessed
			}
			if filesResult != nil {
				totalProcessed += filesResult.FilesProcessed
			}

			if totalProcessed == 0 {
				fmt.Println("ℹ️  No changes found - everything is up to date!")
			} else {
				fmt.Printf("📊 Total items processed: %d\n", totalProcessed)
				if syncFiles && filesResult != nil && filesResult.FilesProcessed > 0 {
					fmt.Println("💡 Use 'maplefile-cli files get FILE_ID' to download file content locally.")
				}
			}

			// Show next steps
			fmt.Printf("\n💡 What's next:\n")
			fmt.Printf("   • View collections: maplefile-cli collections list\n")
			if syncFiles {
				fmt.Printf("   • View files: maplefile-cli files list --collection COLLECTION_ID\n")
			}
			fmt.Printf("   • Add new content: maplefile-cli files add FILE_PATH --collection COLLECTION_ID\n")

			logger.Info("Sync completed",
				zap.Bool("syncedCollections", syncCollections),
				zap.Bool("syncedFiles", syncFiles),
				zap.Int("totalErrors", len(totalErrors)),
				zap.Duration("duration", duration))
		},
	}

	// Define flags
	cmd.Flags().BoolVar(&collections, "collections", false, "Sync only collections")
	cmd.Flags().BoolVar(&files, "files", false, "Sync only file metadata")
	cmd.Flags().Int64Var(&collectionBatchSize, "collection-batch-size", 50, "Collections per batch")
	cmd.Flags().Int64Var(&fileBatchSize, "file-batch-size", 50, "Files per batch")
	cmd.Flags().IntVar(&maxBatches, "max-batches", 100, "Maximum batches to process")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")

	// Mark required flags
	cmd.MarkFlagRequired("password")

	return cmd
}
