// native/desktop/maplefile-cli/cmd/files/misc/download.go
package misc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filedownload"
)

// downloadFileCmd creates a command for downloading and decrypting a file from the cloud
func downloadFileCmd(
	logger *zap.Logger,
	downloadService filedownload.DownloadService,
) *cobra.Command {
	var fileID string
	var outputPath string
	var password string
	var urlDuration time.Duration

	var cmd = &cobra.Command{
		Use:   "download",
		Short: "Download and decrypt a file from the cloud",
		Long: `
Download and decrypt a file from MapleFile Cloud to your local system.

The file will be downloaded from cloud storage, decrypted using your password,
and saved to the specified location with its original filename and content.

Examples:
  # Download file to current directory with original filename
  maplefile-cli files download --file-id 507f1f77bcf86cd799439011 --password 1234567890

  # Download file to specific path
  maplefile-cli files download --file-id 507f1f77bcf86cd799439011 --output ~/Downloads/ --password 1234567890

  # Download with custom URL duration
  maplefile-cli files download --file-id 507f1f77bcf86cd799439011 --duration 2h --password 1234567890
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate required inputs
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to download.")
				return
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE decryption.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Convert to ObjectID
			fileObjectID, err := gocql.ParseUUID(fileID)
			if err != nil {
				fmt.Printf("❌ Error: Invalid file ID format: %v\n", err)
				return
			}

			// Set default URL duration if not provided
			if urlDuration == 0 {
				urlDuration = 1 * time.Hour
			}

			// Download and decrypt file
			fmt.Printf("🔄 Downloading and decrypting file: %s\n", fileID)
			fmt.Println("📡 Step 1/4: Getting presigned download URLs...")

			result, err := downloadService.DownloadAndDecryptFile(ctx, fileObjectID, password, urlDuration)
			if err != nil {
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else if strings.Contains(err.Error(), "permission") {
					fmt.Printf("❌ Error: You don't have permission to download this file.\n")
				} else if strings.Contains(err.Error(), "not found") {
					fmt.Printf("❌ Error: File not found.\n")
				} else {
					fmt.Printf("❌ Error downloading file: %v\n", err)
				}
				return
			}

			fmt.Println("📥 Step 2/4: Downloaded encrypted content...")
			fmt.Println("🔓 Step 3/4: Decrypted file content...")

			// Determine output directory and filename
			var outputDir string
			var filename string

			if outputPath != "" {
				// Check if outputPath is a directory or file
				if stat, err := os.Stat(outputPath); err == nil && stat.IsDir() {
					// outputPath is an existing directory
					outputDir = outputPath
					filename = result.DecryptedMetadata.Name
				} else {
					// outputPath is a file path (may not exist yet)
					outputDir = filepath.Dir(outputPath)
					filename = filepath.Base(outputPath)
				}
			} else {
				// Use current directory with original filename
				outputDir = "."
				filename = result.DecryptedMetadata.Name
			}

			// Ensure output directory exists
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Printf("❌ Error creating output directory: %v\n", err)
				return
			}

			// Create final output path
			finalOutputPath := filepath.Join(outputDir, filename)

			// Handle file conflicts
			if _, err := os.Stat(finalOutputPath); err == nil {
				// File exists, create a unique name
				ext := filepath.Ext(filename)
				nameWithoutExt := strings.TrimSuffix(filename, ext)
				counter := 1
				for {
					newFilename := fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
					newPath := filepath.Join(outputDir, newFilename)
					if _, err := os.Stat(newPath); os.IsNotExist(err) {
						finalOutputPath = newPath
						filename = newFilename
						break
					}
					counter++
				}
			}

			// Write decrypted file data
			if err := os.WriteFile(finalOutputPath, result.DecryptedData, 0644); err != nil {
				fmt.Printf("❌ Error saving decrypted file: %v\n", err)
				return
			}

			fmt.Println("💾 Step 4/4: Saved decrypted file...")

			// Write thumbnail if present
			var thumbnailPath string
			if result.ThumbnailData != nil && len(result.ThumbnailData) > 0 {
				ext := filepath.Ext(filename)
				nameWithoutExt := strings.TrimSuffix(filename, ext)
				thumbnailFilename := fmt.Sprintf("%s_thumbnail.jpg", nameWithoutExt)
				thumbnailPath = filepath.Join(outputDir, thumbnailFilename)

				if err := os.WriteFile(thumbnailPath, result.ThumbnailData, 0644); err != nil {
					fmt.Printf("⚠️  Warning: Failed to save thumbnail: %v\n", err)
				}
			}

			// Display success information
			fmt.Printf("\n✅ File successfully downloaded and decrypted!\n")
			fmt.Printf("🆔 File ID: %s\n", fileID)
			fmt.Printf("📁 Original Name: %s\n", result.DecryptedMetadata.Name)
			fmt.Printf("💾 Saved to: %s\n", finalOutputPath)
			// fmt.Printf("📏 File Size: %s (%d bytes)\n", formatFileSize(result.OriginalSize), result.OriginalSize)
			fmt.Printf("🗂️  MIME Type: %s\n", result.DecryptedMetadata.MimeType)

			// if thumbnailPath != "" {
			// 	fmt.Printf("🖼️  Thumbnail: %s (%s)\n", thumbnailPath, formatFileSize(result.ThumbnailSize))
			// }

			// Show file creation time if available
			if result.DecryptedMetadata.Created > 0 {
				createdTime := time.Unix(result.DecryptedMetadata.Created, 0)
				fmt.Printf("📅 Original Created: %s\n", createdTime.Format("2006-01-02 15:04:05"))
			}

			fmt.Printf("\n🎉 Download and decryption completed successfully!\n")
			fmt.Printf("🔐 Your file has been decrypted and is ready to use.\n")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to download (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory or file path (defaults to current directory with original filename)")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE decryption)")
	cmd.MarkFlagRequired("password")
	cmd.Flags().DurationVar(&urlDuration, "duration", 1*time.Hour, "Duration for presigned URLs (e.g., 1h, 30m)")

	return cmd
}
