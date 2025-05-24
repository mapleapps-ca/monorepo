// native/desktop/maplefile-cli/cmd/files/upload.go
package files

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/fileupload"
)

// uploadFileCmd creates a command for uploading a file to the cloud
func uploadFileCmd(
	logger *zap.Logger,
	uploadService fileupload.UploadService,
) *cobra.Command {
	var fileID string

	var cmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload a local file to the cloud",
		Long: `
Uploads a local file to MapleFile Cloud using three-step upload process.

The file must have a sync status of "local_only" to be eligible for upload.
The upload process includes:
1. Creating a pending file record in the cloud
2. Uploading the encrypted file content
3. Completing the upload and marking the file as synced

Example:
  maplefile-cli files upload --file-id 507f1f77bcf86cd799439011
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate file ID
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to upload.")
				return
			}

			// Convert to ObjectID
			fileObjectID, err := primitive.ObjectIDFromHex(fileID)
			if err != nil {
				fmt.Printf("❌ Error: Invalid file ID format: %v\n", err)
				return
			}

			// Upload file
			fmt.Printf("🔄 Uploading file: %s\n", fileID)
			fmt.Println("📡 Step 1/3: Creating pending file record...")

			result, err := uploadService.UploadFile(ctx, fileObjectID)
			if err != nil {
				fmt.Printf("❌ Error uploading file: %v\n", err)
				return
			}

			if !result.Success {
				fmt.Printf("❌ Upload failed: %v\n", result.Error)
				return
			}

			// Display success
			fmt.Printf("\n✅ File successfully uploaded to MapleFile Cloud!\n")
			fmt.Printf("🆔 Local File ID: %s\n", result.FileID.Hex())
			fmt.Printf("☁️  Cloud File ID: %s\n", result.CloudFileID.Hex())
			fmt.Printf("📏 Uploaded Size: %d bytes\n", result.FileSizeBytes)
			if result.ThumbnailSizeBytes > 0 {
				fmt.Printf("🖼️  Thumbnail Size: %d bytes\n", result.ThumbnailSizeBytes)
			}
			fmt.Printf("📅 Uploaded At: %s\n", result.UploadedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("\n🎉 Your file is now synced with the cloud!\n")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to upload (required)")
	cmd.MarkFlagRequired("file-id")

	return cmd
}
