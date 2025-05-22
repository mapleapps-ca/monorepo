// native/desktop/maplefile-cli/cmd/filesyncer/upload.go
package filesyncer

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	dom_localfile "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
)

// uploadLocalFileCmd creates a command for uploading a local file to the remote backend
func uploadLocalFileCmd(
	uploadService filesyncer.UploadService,
	getService localfile.GetService,
	logger *zap.Logger,
) *cobra.Command {
	var fileID string
	var force bool

	var cmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload a local file to the remote backend",
		Long: `
Upload a local file to the remote backend.

This command uploads the encrypted version of a local file to the remote MapleFile backend.
Only encrypted files can be uploaded for security reasons.

The behavior depends on the file's current sync status:
* If the file is local-only, a new remote file will be created
* If the file is already synced, you must use --force to re-upload
* If the file has local modifications, it will be uploaded to update the remote version

Storage mode requirements:
* "encrypted_only": File will be uploaded (recommended)
* "hybrid": Encrypted version will be uploaded
* "decrypted_only": Cannot be uploaded (not allowed for security)

Examples:
  # Upload a local-only file
  maplefile-cli filesyncer upload --file-id 507f1f77bcf86cd799439011

  # Force re-upload an already synced file
  maplefile-cli filesyncer upload --file-id 507f1f77bcf86cd799439011 --force

  # Upload a file with local modifications
  maplefile-cli filesyncer upload --file-id 507f1f77bcf86cd799439011
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate required fields
			if fileID == "" {
				fmt.Println("Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the local file ID.")
				return
			}

			// Get file information first to show user what will be uploaded
			fileOutput, err := getService.GetByID(ctx, fileID)
			if err != nil {
				fmt.Printf("🐞 Error getting file information: %v\n", err)
				return
			}

			if fileOutput == nil || fileOutput.File == nil {
				fmt.Println("Error: File not found.")
				return
			}

			file := fileOutput.File

			// Display file information
			fmt.Printf("📁 File: %s\n", file.DecryptedName)
			fmt.Printf("🔐 Storage Mode: %s\n", file.StorageMode)
			fmt.Printf("📊 Sync Status: %s\n", getSyncStatusText(file.SyncStatus))
			fmt.Printf("💾 File Size: %d bytes\n", file.FileSize)

			// Validate file can be uploaded
			if file.StorageMode == "decrypted_only" {
				fmt.Println("❌ Error: Cannot upload decrypted-only files.")
				fmt.Println("Only encrypted files can be uploaded to the remote backend for security.")
				return
			}

			if file.EncryptedFilePath == "" && (file.StorageMode == "encrypted_only" || file.StorageMode == "hybrid") {
				fmt.Println("❌ Error: File has no encrypted data available.")
				return
			}

			// Check if force is needed
			if file.SyncStatus == dom_localfile.SyncStatusSynced && !force {
				fmt.Println("⚠️  File is already synced with remote backend.")
				fmt.Println("Use --force flag to re-upload the file.")
				return
			}

			// Confirm upload
			if !force {
				var action string
				switch file.SyncStatus {
				case 0: // SyncStatusLocalOnly
					action = "create a new remote file"
				case 3: // SyncStatusModifiedLocally
					action = "update the existing remote file"
				default:
					action = "upload the file"
				}

				fmt.Printf("\n🚀 Ready to %s.\n", action)
				if !confirmAction("Do you want to continue? (y/n): ") {
					fmt.Println("Upload cancelled.")
					return
				}
			}

			fmt.Println("\n🔄 Uploading file to remote backend...")

			logger.Debug("Uploading local file",
				zap.String("fileID", fileID),
				zap.String("fileName", file.DecryptedName),
				zap.String("storageMode", file.StorageMode),
				zap.Bool("force", force))

			// Prepare upload input
			uploadInput := filesyncer.UploadInput{
				LocalFileID: fileID,
				ForceUpdate: force,
			}

			// Upload the file
			result, err := uploadService.Upload(ctx, uploadInput)
			if err != nil {
				fmt.Printf("🐞 Error uploading file: %v\n", err)
				return
			}

			// Display success message
			fmt.Println("\n✅ File uploaded successfully!")
			fmt.Printf("Action: %s\n", result.Action)
			fmt.Printf("Local File ID: %s\n", result.LocalFile.ID.Hex())
			fmt.Printf("Remote File ID: %s\n", result.RemoteFile.ID.Hex())
			fmt.Printf("File Name: %s\n", result.LocalFile.DecryptedName)
			fmt.Printf("Sync Status: %s\n", getSyncStatusText(result.LocalFile.SyncStatus))

			if result.RemoteFile.UploadURL != "" {
				fmt.Printf("Upload completed to: %s\n", result.RemoteFile.FileObjectKey)
			}

			fmt.Printf("\n📊 Summary:\n")
			fmt.Printf("  - Local file is now synced with remote backend\n")
			fmt.Printf("  - File size: %d bytes\n", result.RemoteFile.FileSize)
			fmt.Printf("  - Encryption version: %s\n", result.RemoteFile.EncryptionVersion)
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&fileID, "file-id", "", "ID of the local file to upload (required)")
	cmd.Flags().BoolVar(&force, "force", false, "Force upload even if file is already synced")

	// Mark required flags
	cmd.MarkFlagRequired("file-id")

	return cmd
}

// getSyncStatusText returns a human-readable sync status
func getSyncStatusText(status dom_localfile.SyncStatus) string {
	switch status {
	case dom_localfile.SyncStatusLocalOnly: // SyncStatusLocalOnly
		return "Local Only"
	case dom_localfile.SyncStatusCloudOnly: // SyncStatusCloudOnly
		return "Cloud Only"
	case dom_localfile.SyncStatusSynced: // SyncStatusSynced
		return "Synced"
	case dom_localfile.SyncStatusModifiedLocally: // SyncStatusModifiedLocally
		return "Modified Locally"
	default:
		return "Unknown"
	}
}
