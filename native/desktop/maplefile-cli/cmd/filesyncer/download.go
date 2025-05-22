// native/desktop/maplefile-cli/cmd/filesyncer/download.go
package filesyncer

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
)

// downloadRemoteFileCmd creates a command for downloading a remote file to local storage
func downloadRemoteFileCmd(
	downloadService filesyncer.DownloadToLocalService,
	logger *zap.Logger,
) *cobra.Command {
	var remoteFileID string
	var force bool

	var cmd = &cobra.Command{
		Use:   "download",
		Short: "Download a remote file to local storage",
		Long: `
Download a remote file to local storage.

This command downloads a remote file from the MapleFile backend and stores it locally
in encrypted form. The file can then be decrypted and accessed locally.

The behavior depends on whether the file already exists locally:
* If the file doesn't exist locally, it will be downloaded and created
* If the file already exists and is synced, you must use --force to re-download
* If the remote file is newer than the local version, it will be downloaded

Examples:
  # Download a remote file that doesn't exist locally
  maplefile-cli filesyncer download --remote-id 507f1f77bcf86cd799439011

  # Force re-download an already synced file
  maplefile-cli filesyncer download --remote-id 507f1f77bcf86cd799439011 --force
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate required fields
			if remoteFileID == "" {
				fmt.Println("Error: Remote file ID is required.")
				fmt.Println("Use --remote-id flag to specify the remote file ID.")
				return
			}

			// Convert string ID to ObjectID
			remoteID, err := primitive.ObjectIDFromHex(remoteFileID)
			if err != nil {
				fmt.Printf("🐞 Error: Invalid remote file ID format: %v\n", err)
				return
			}

			// Confirm download
			if !force {
				fmt.Printf("🚀 Ready to download remote file %s.\n", remoteFileID)
				if !confirmAction("Do you want to continue? (y/n): ") {
					fmt.Println("Download cancelled.")
					return
				}
			}

			fmt.Println("\n🔄 Downloading file from remote backend...")

			logger.Debug("Downloading remote file",
				zap.String("remoteFileID", remoteFileID),
				zap.Bool("force", force))

			// Prepare download input
			downloadInput := filesyncer.DownloadToLocalInput{
				RemoteID: remoteID,
			}

			// Download the file
			result, err := downloadService.Execute(ctx, downloadInput)
			if err != nil {
				fmt.Printf("🐞 Error downloading file: %v\n", err)
				return
			}

			// Display success message
			fmt.Println("\n✅ File downloaded successfully!")
			fmt.Printf("Action: %s\n", result.SynchronizationLog)
			fmt.Printf("Local File ID: %s\n", result.LocalFile.ID.Hex())
			fmt.Printf("Remote File ID: %s\n", result.RemoteFile.ID.Hex())
			fmt.Printf("File Name: %s\n", result.LocalFile.DecryptedName)
			fmt.Printf("Sync Status: %s\n", getSyncStatusText(result.LocalFile.SyncStatus))
			fmt.Printf("Download Status: %t\n", result.DownloadedToLocal)

			fmt.Printf("\n📊 Summary:\n")
			fmt.Printf("  - Remote file is now available locally\n")
			fmt.Printf("  - File size: %d bytes\n", result.LocalFile.DecryptedFileSize)
			fmt.Printf("  - Storage mode: %s\n", result.LocalFile.StorageMode)
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&remoteFileID, "remote-id", "", "ID of the remote file to download (required)")
	cmd.Flags().BoolVar(&force, "force", false, "Force download even if file is already synced")

	// Mark required flags
	cmd.MarkFlagRequired("remote-id")

	return cmd
}
