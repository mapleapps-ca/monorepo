// native/desktop/maplefile-cli/cmd/filesyncer/sync_collection.go
package filesyncer

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
)

// syncCollectionCmd creates a command for syncing all files in a collection
func syncCollectionCmd(
	syncCollectionService filesyncer.SyncCollectionService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionID string
	var force bool
	var verbose bool

	var cmd = &cobra.Command{
		Use:   "sync-collection",
		Short: "Synchronize all files in a collection",
		Long: `
Synchronize all files in a collection between local and remote storage.

This command processes all files in a collection and synchronizes them based on
their current state and modification times. It's useful for bulk synchronization
operations.

Sync behavior per file:
* Local-only files: Upload to remote
* Remote-only files: Download to local
* Files in both places: Sync newest version
* Already synced files: Skip (unless forced)

Examples:
  # Sync all files in a collection (bidirectional)
  maplefile-cli filesyncer sync-collection --collection-id 507f1f77bcf86cd799439011

  # Verbose output to see details of each file
  maplefile-cli filesyncer sync-collection --collection-id 507f1f77bcf86cd799439011 --verbose

  # Force sync without confirmation
  maplefile-cli filesyncer sync-collection --collection-id 507f1f77bcf86cd799439011 --force
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate required fields
			if collectionID == "" {
				fmt.Println("Error: Collection ID is required.")
				fmt.Println("Use --collection-id flag to specify the collection ID.")
				return
			}

			// Convert string ID to ObjectID
			collectionObjID, err := primitive.ObjectIDFromHex(collectionID)
			if err != nil {
				fmt.Printf("🐞 Error: Invalid collection ID format: %v\n", err)
				return
			}

			// Confirm sync
			if !force {
				fmt.Printf("🚀 Ready to sync all files in collection: %s\n", collectionID)
				if !confirmAction("Do you want to continue? (y/n): ") {
					fmt.Println("Collection sync cancelled.")
					return
				}
			}

			fmt.Println("\n🔄 Synchronizing collection files...")

			logger.Debug("Syncing collection",
				zap.String("collectionID", collectionID),
				zap.Bool("force", force),
				zap.Bool("verbose", verbose))

			// Prepare sync input
			syncInput := filesyncer.SyncCollectionInput{
				CollectionID: collectionObjID,
			}

			// Sync the collection
			result, err := syncCollectionService.Execute(ctx, syncInput)
			if err != nil {
				fmt.Printf("🐞 Error syncing collection: %v\n", err)
				return
			}

			// Display summary
			fmt.Println("\n✅ Collection synchronization completed!")
			fmt.Printf("Total Files: %d\n", result.TotalFiles)
			fmt.Printf("Successful Syncs: %d\n", result.SuccessfulSyncs)
			fmt.Printf("Failed Syncs: %d\n", result.FailedSyncs)
			fmt.Printf("Files Uploaded: %d\n", result.UploadedFiles)
			fmt.Printf("Files Downloaded: %d\n", result.DownloadedFiles)

			// Show detailed results if verbose
			if verbose && len(result.Details) > 0 {
				fmt.Printf("\n📋 Detailed Results:\n")
				for i, detail := range result.Details {
					fmt.Printf("  %d. %s\n", i+1, detail.SynchronizationLog)
					if detail.LocalFile != nil {
						fmt.Printf("     Local: %s (%s)\n", detail.LocalFile.DecryptedName, detail.LocalFile.ID.Hex())
					}
					if detail.RemoteFile != nil {
						fmt.Printf("     Remote: %s\n", detail.RemoteFile.ID.Hex())
					}
					fmt.Printf("     Direction: %s\n", detail.SyncDirection)
					fmt.Printf("     Uploaded: %t, Downloaded: %t\n", detail.UploadedToRemote, detail.DownloadedToLocal)
					fmt.Println()
				}
			}

			fmt.Printf("\n📊 Summary:\n")
			if result.SuccessfulSyncs > 0 {
				fmt.Printf("  - Successfully synchronized %d files\n", result.SuccessfulSyncs)
			}
			if result.FailedSyncs > 0 {
				fmt.Printf("  - Failed to synchronize %d files (check logs for details)\n", result.FailedSyncs)
			}
			if result.UploadedFiles > 0 {
				fmt.Printf("  - Uploaded %d files to remote\n", result.UploadedFiles)
			}
			if result.DownloadedFiles > 0 {
				fmt.Printf("  - Downloaded %d files to local\n", result.DownloadedFiles)
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&collectionID, "collection-id", "", "ID of the collection to sync (required)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompts")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed results for each file")

	// Mark required flags
	cmd.MarkFlagRequired("collection-id")

	return cmd
}
