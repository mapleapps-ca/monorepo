// cmd/files/get.go - Clean unified get/download command
package files

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
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
)

// getFileCmd creates a unified command for downloading and accessing files
func getFileCmd(
	logger *zap.Logger,
	downloadService filedownload.DownloadService,
	onloadService filesyncer.OnloadService,
) *cobra.Command {
	var outputPath string
	var password string
	var force bool

	var cmd = &cobra.Command{
		Use:   "get FILE_ID",
		Short: "Download and decrypt a file",
		Long: `
Download and decrypt a file from MapleFile.

This command automatically handles different file states:
  • Cloud-only files: Downloads from cloud and decrypts
  • Local files: Copies decrypted version to output location
  • Encrypted-only files: Decrypts in-place

Examples:
  # Download file to current directory with original name
  maplefile-cli files get 507f1f77bcf86cd799439011 --password mypass

  # Download to specific directory
  maplefile-cli files get 507f1f77bcf86cd799439011 --output ~/Downloads/ --password mypass

  # Download with custom filename
  maplefile-cli files get 507f1f77bcf86cd799439011 --output ~/Documents/my-file.pdf --password mypass

  # Overwrite existing file
  maplefile-cli files get 507f1f77bcf86cd799439011 --output existing-file.txt --force --password mypass
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			fileID := args[0]

			// Validate required inputs
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

			fmt.Printf("📥 Getting file: %s\n", fileID)

			// Try to determine the best approach based on file state
			// For now, we'll use the download service which handles most cases
			result, err := downloadService.DownloadAndDecryptFile(ctx, fileObjectID, password, 1*time.Hour)
			if err != nil {
				// If download fails, it might be a local-only file that needs onload
				if strings.Contains(err.Error(), "cloud-only") || strings.Contains(err.Error(), "not found in cloud") {
					fmt.Printf("📱 File appears to be local-only, attempting onload...\n")

					onloadInput := &filesyncer.OnloadInput{
						FileID:       fileObjectID,
						UserPassword: password,
					}

					onloadResult, onloadErr := onloadService.Onload(ctx, onloadInput)
					if onloadErr != nil {
						fmt.Printf("❌ Error accessing file: %v\n", onloadErr)
						return
					}

					fmt.Printf("✅ File made available locally!\n")
					fmt.Printf("💾 Local Path: %s\n", onloadResult.DecryptedPath)

					// Copy to output location if specified
					if outputPath != "" {
						err = copyFileToOutput(onloadResult.DecryptedPath, outputPath, result.DecryptedMetadata.Name, force)
						if err != nil {
							fmt.Printf("❌ Error copying to output location: %v\n", err)
							return
						}
					}

					return
				}

				// Handle other download errors
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else if strings.Contains(err.Error(), "permission") {
					fmt.Printf("❌ Error: You don't have permission to download this file.\n")
				} else if strings.Contains(err.Error(), "not found") {
					fmt.Printf("❌ Error: File not found.\n")
				} else {
					fmt.Printf("❌ Error getting file: %v\n", err)
				}
				return
			}

			fmt.Printf("✅ File downloaded and decrypted!\n")

			// Determine output location
			var finalOutputPath string
			var filename string

			if outputPath != "" {
				finalOutputPath, filename, err = determineOutputPath(outputPath, result.DecryptedMetadata.Name, force)
				if err != nil {
					fmt.Printf("❌ Error determining output path: %v\n", err)
					return
				}
			} else {
				// Use current directory with original filename
				filename = result.DecryptedMetadata.Name
				finalOutputPath = filepath.Join(".", filename)

				// Handle conflicts in current directory
				if _, err := os.Stat(finalOutputPath); err == nil && !force {
					finalOutputPath = generateUniqueFilename(finalOutputPath)
					filename = filepath.Base(finalOutputPath)
				}
			}

			// Write decrypted file data
			if err := os.WriteFile(finalOutputPath, result.DecryptedData, 0644); err != nil {
				fmt.Printf("❌ Error saving file: %v\n", err)
				return
			}

			// Display success information
			fmt.Printf("\n📋 File Details:\n")
			fmt.Printf("  🆔 File ID: %s\n", fileID)
			fmt.Printf("  📁 Original Name: %s\n", result.DecryptedMetadata.Name)
			fmt.Printf("  💾 Saved to: %s\n", finalOutputPath)
			fmt.Printf("  📏 Size: %s (%d bytes)\n", formatFileSize(result.OriginalSize), result.OriginalSize)
			fmt.Printf("  🏷️  MIME Type: %s\n", result.DecryptedMetadata.MimeType)

			// Save thumbnail if present
			if result.ThumbnailData != nil && len(result.ThumbnailData) > 0 {
				thumbnailPath := generateThumbnailPath(finalOutputPath)
				if err := os.WriteFile(thumbnailPath, result.ThumbnailData, 0644); err == nil {
					fmt.Printf("  🖼️  Thumbnail: %s (%s)\n", thumbnailPath, formatFileSize(result.ThumbnailSize))
				}
			}

			// Show file creation time if available
			if result.DecryptedMetadata.Created > 0 {
				createdTime := time.Unix(result.DecryptedMetadata.Created, 0)
				fmt.Printf("  📅 Created: %s\n", createdTime.Format("2006-01-02 15:04:05"))
			}

			fmt.Printf("\n🎉 File is ready to use!\n")

			logger.Info("File downloaded successfully",
				zap.String("fileID", fileID),
				zap.String("outputPath", finalOutputPath),
				zap.Int64("size", result.OriginalSize))
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory or file path (defaults to current directory)")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE decryption)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files without confirmation")

	// Mark required flags
	cmd.MarkFlagRequired("password")

	return cmd
}

// determineOutputPath figures out where to save the file
func determineOutputPath(outputPath, originalName string, force bool) (finalPath, filename string, err error) {
	// Check if outputPath is a directory or file
	if stat, statErr := os.Stat(outputPath); statErr == nil && stat.IsDir() {
		// outputPath is an existing directory
		filename = originalName
		finalPath = filepath.Join(outputPath, filename)
	} else {
		// outputPath is a file path (may not exist yet)
		dir := filepath.Dir(outputPath)
		filename = filepath.Base(outputPath)
		finalPath = outputPath

		// Ensure directory exists
		if err = os.MkdirAll(dir, 0755); err != nil {
			return "", "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Handle file conflicts
	if _, err := os.Stat(finalPath); err == nil && !force {
		finalPath = generateUniqueFilename(finalPath)
		filename = filepath.Base(finalPath)
	}

	return finalPath, filename, nil
}

// generateUniqueFilename creates a unique filename to avoid conflicts
func generateUniqueFilename(originalPath string) string {
	dir := filepath.Dir(originalPath)
	ext := filepath.Ext(originalPath)
	nameWithoutExt := strings.TrimSuffix(filepath.Base(originalPath), ext)

	counter := 1
	for {
		newFilename := fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
		newPath := filepath.Join(dir, newFilename)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

// generateThumbnailPath creates a path for saving thumbnails
func generateThumbnailPath(filePath string) string {
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	nameWithoutExt := strings.TrimSuffix(filepath.Base(filePath), ext)
	return filepath.Join(dir, fmt.Sprintf("%s_thumbnail.jpg", nameWithoutExt))
}

// copyFileToOutput copies a file from source to destination
func copyFileToOutput(sourcePath, outputPath, originalName string, force bool) error {
	// Read source file
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Determine final output path
	finalPath, _, err := determineOutputPath(outputPath, originalName, force)
	if err != nil {
		return err
	}

	// Write to destination
	if err := os.WriteFile(finalPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("✅ File copied to: %s\n", finalPath)
	return nil
}
