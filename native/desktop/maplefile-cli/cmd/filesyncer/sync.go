// native/desktop/maplefile-cli/cmd/filesyncer/sync.go
package filesyncer

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
)

// syncFileCmd creates a command for syncing a file by file ID
func syncFileCmd(
	syncService filesyncer.SyncFileService,
	logger *zap.Logger,
) *cobra.Command {
	var fileID string
	var force bool

	var cmd = &cobra.Command{
		Use:   "sync",
		Short: "Synchronize a file between local and remote",
		Long: `
Synchronize a file between local and remote storage.

This command finds a file by its ID and synchronizes it between
local and remote storage. The sync direction is determined automatically based
on file existence and modification times.

Sync behavior:
* If file exists only locally: Upload to remote
* If file exists only remotely: Download to local
* If file exists in both places: Sync newest version
* If both versions are identical: No action needed

Examples:
  # Auto-sync a file (recommended)
  maplefile-cli filesyncer sync --file-id 507f1f77bcf86cd799439011

  # Force sync without confirmation
  maplefile-cli filesyncer sync --file-id 507f1f77bcf86cd799439011 --force
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate required fields
			if fileID == "" {
				fmt.Println("Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file ID.")
				return
			}

			// Convert string ID to ObjectID
			objectID, err := primitive.ObjectIDFromHex(fileID)
			if err != nil {
				fmt.Printf("🐞 Error: Invalid file ID format: %v\n", err)
				return
			}

			// Confirm sync
			if !force {
				fmt.Printf("🚀 Ready to auto-sync file with ID: %s\n", fileID)
				if !confirmAction("Do you want to continue? (y/n): ") {
					fmt.Println("Sync cancelled.")
					return
				}
			}

			fmt.Println("\n🔄 Synchronizing file...")

			logger.Debug("Syncing file",
				zap.String("fileID", fileID),
				zap.Bool("force", force))

			// Prepare sync input
			syncInput := filesyncer.SyncFileInput{
				FileID: objectID,
			}

			// Sync the file
			result, err := syncService.Execute(ctx, syncInput)
			if err != nil {
				fmt.Printf("🐞 Error syncing file: %v\n", err)
				return
			}

			// Display success message
			fmt.Println("\n✅ File synchronized successfully!")
			fmt.Printf("Action: %s\n", result.SynchronizationLog)
			fmt.Printf("Sync Direction: %s\n", result.SyncDirection)

			if result.LocalFile != nil {
				fmt.Printf("Local File ID: %s\n", result.LocalFile.ID.Hex())
				fmt.Printf("File Name: %s\n", result.LocalFile.DecryptedName)
				fmt.Printf("Sync Status: %s\n", getSyncStatusText(result.LocalFile.SyncStatus))
			}

			if result.RemoteFile != nil {
				fmt.Printf("Remote File ID: %s\n", result.RemoteFile.ID.Hex())
			}

			fmt.Printf("Upload Status: %t\n", result.UploadedToRemote)
			fmt.Printf("Download Status: %t\n", result.DownloadedToLocal)

			fmt.Printf("\n📊 Summary:\n")
			switch result.SyncDirection {
			case "upload":
				fmt.Printf("  - Local file uploaded to remote backend\n")
			case "download":
				fmt.Printf("  - Remote file downloaded to local storage\n")
			case "none":
				fmt.Printf("  - Files are already in sync\n")
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&fileID, "file-id", "", "File ID to sync (required)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompts")

	// Mark required flags
	cmd.MarkFlagRequired("file-id")

	return cmd
}
