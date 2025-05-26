// native/desktop/maplefile-cli/cmd/files/download.go
package files

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filedownload"
)

// downloadFileCmd creates a command for downloading a file from the cloud
func downloadFileCmd(
	logger *zap.Logger,
	downloadService filedownload.DownloadService,
) *cobra.Command {
	var fileID string
	var outputPath string
	var urlDuration time.Duration

	var cmd = &cobra.Command{
		Use:   "download",
		Short: "Download a file from the cloud",
		Long: `
Download a file from MapleFile Cloud to your local system.

The file will be downloaded from cloud storage using presigned URLs and saved to the specified location.

Examples:
  # Download file to current directory
  maplefile-cli files download --file-id 507f1f77bcf86cd799439011

  # Download file to specific path
  maplefile-cli files download --file-id 507f1f77bcf86cd799439011 --output ~/Downloads/myfile.pdf

  # Download with custom URL duration
  maplefile-cli files download --file-id 507f1f77bcf86cd799439011 --duration 2h
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate required inputs
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to download.")
				return
			}

			// Convert to ObjectID
			fileObjectID, err := primitive.ObjectIDFromHex(fileID)
			if err != nil {
				fmt.Printf("❌ Error: Invalid file ID format: %v\n", err)
				return
			}

			// Set default URL duration if not provided
			if urlDuration == 0 {
				urlDuration = 1 * time.Hour
			}

			// Download file
			fmt.Printf("🔄 Downloading file: %s\n", fileID)
			fmt.Println("📡 Getting presigned download URLs...")

			result, err := downloadService.DownloadFile(ctx, fileObjectID, urlDuration)
			if err != nil {
				if strings.Contains(err.Error(), "permission") {
					fmt.Printf("❌ Error: You don't have permission to download this file.\n")
				} else if strings.Contains(err.Error(), "not found") {
					fmt.Printf("❌ Error: File not found.\n")
				} else {
					fmt.Printf("❌ Error downloading file: %v\n", err)
				}
				return
			}

			// Determine output path
			var finalOutputPath string
			if outputPath != "" {
				finalOutputPath = outputPath
			} else {
				// Use current directory with file ID as filename
				finalOutputPath = fmt.Sprintf("./%s.bin", fileID)
			}

			// Create output directory if needed
			if err := os.MkdirAll(filepath.Dir(finalOutputPath), 0755); err != nil {
				fmt.Printf("❌ Error creating output directory: %v\n", err)
				return
			}

			// Write file data
			if err := os.WriteFile(finalOutputPath, result.FileData, 0644); err != nil {
				fmt.Printf("❌ Error saving file: %v\n", err)
				return
			}

			// Write thumbnail if present
			var thumbnailPath string
			if result.ThumbnailData != nil && len(result.ThumbnailData) > 0 {
				ext := filepath.Ext(finalOutputPath)
				name := strings.TrimSuffix(finalOutputPath, ext)
				thumbnailPath = fmt.Sprintf("%s_thumbnail.jpg", name)

				if err := os.WriteFile(thumbnailPath, result.ThumbnailData, 0644); err != nil {
					fmt.Printf("⚠️  Warning: Failed to save thumbnail: %v\n", err)
				}
			}

			// Display success
			fmt.Printf("\n✅ File successfully downloaded!\n")
			fmt.Printf("🆔 File ID: %s\n", fileID)
			fmt.Printf("📁 Saved to: %s\n", finalOutputPath)
			fmt.Printf("📏 File Size: %d bytes\n", result.FileSize)
			if thumbnailPath != "" {
				fmt.Printf("🖼️  Thumbnail: %s (%d bytes)\n", thumbnailPath, result.ThumbnailSize)
			}
			fmt.Printf("\n🎉 Download completed successfully!\n")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to download (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (defaults to current directory)")
	cmd.Flags().DurationVar(&urlDuration, "duration", 1*time.Hour, "Duration for presigned URLs (e.g., 1h, 30m)")

	return cmd
}
