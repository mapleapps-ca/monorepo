// cmd/files/add.go - Clean unified add command
package files

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/fileupload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
)

// addFileCmd creates a unified command for adding files with auto-upload
func addFileCmd(
	logger *zap.Logger,
	addService localfile.LocalFileAddService,
	uploadService fileupload.FileUploadService,
) *cobra.Command {
	var collectionID string
	var name string
	var storageMode string
	var password string
	var localOnly bool

	var cmd = &cobra.Command{
		Use:   "add FILE_PATH",
		Short: "Add a file to your collection",
		Long: `
Add a file from your filesystem to a MapleFile collection.

By default, the file is encrypted locally and automatically uploaded to the cloud.
Use --local-only to skip the upload step.

Storage modes control how files are stored:
  • encrypted_only: Only encrypted version kept locally (most secure)
  • hybrid: Both encrypted and decrypted versions (convenient, default)
  • decrypted_only: Only decrypted version (not recommended)

Examples:
  # Add file with auto-upload (recommended)
  maplefile-cli files add "/path/to/document.pdf" --collection 507f1f77bcf86cd799439011 --password mypass

  # Add file locally only (upload later)
  maplefile-cli files add "/path/to/photo.jpg" --collection 507f1f77bcf86cd799439011 --local-only --password mypass

  # Add with custom name
  maplefile-cli files add "/path/to/file.txt" --collection 507f1f77bcf86cd799439011 --name "My Document" --password mypass

  # Add with encrypted-only storage (most secure)
  maplefile-cli files add "/path/to/secret.pdf" --collection 507f1f77bcf86cd799439011 --storage-mode encrypted_only --password mypass
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			filePath := args[0]

			// Validate required parameters
			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			if collectionID == "" {
				fmt.Println("❌ Error: Collection ID is required.")
				fmt.Println("Use --collection flag to specify the target collection.")
				return
			}

			// Convert collection ID
			collectionObjectID, err := primitive.ObjectIDFromHex(collectionID)
			if err != nil {
				fmt.Printf("❌ Error: Invalid collection ID format: %v\n", err)
				return
			}

			// Set default storage mode
			if storageMode == "" {
				storageMode = dom_file.StorageModeHybrid // More convenient default
			}

			// Validate storage mode
			validModes := []string{dom_file.StorageModeEncryptedOnly, dom_file.StorageModeHybrid, dom_file.StorageModeDecryptedOnly}
			isValid := false
			for _, mode := range validModes {
				if storageMode == mode {
					isValid = true
					break
				}
			}
			if !isValid {
				fmt.Printf("❌ Error: Invalid storage mode '%s'\n", storageMode)
				fmt.Printf("Valid modes: %s\n", strings.Join(validModes, ", "))
				return
			}

			// Step 1: Add file locally
			fmt.Printf("📁 Adding file to collection: %s\n", filePath)

			input := &localfile.LocalFileAddInput{
				FilePath:     filePath,
				CollectionID: collectionObjectID,
				OwnerID:      primitive.NewObjectID(), // Service will use authenticated user ID
				Name:         name,
				StorageMode:  storageMode,
			}

			output, err := addService.Add(ctx, input, password)
			if err != nil {
				fmt.Printf("❌ Error adding file: %v\n", err)
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("💡 Tip: Check your password and try again.\n")
				} else if strings.Contains(err.Error(), "file not found") {
					fmt.Printf("💡 Tip: Check the file path is correct.\n")
				} else if strings.Contains(err.Error(), "collection not found") {
					fmt.Printf("💡 Tip: Check the collection ID with: maplefile-cli collections list\n")
				}
				return
			}

			fmt.Printf("✅ File added locally!\n")
			fmt.Printf("🆔 File ID: %s\n", output.File.ID.Hex())
			fmt.Printf("📁 Name: %s\n", output.File.Name)
			fmt.Printf("📏 Size: %s\n", formatFileSize(output.File.FileSize))
			fmt.Printf("🔐 Storage Mode: %s\n", output.File.StorageMode)

			// Step 2: Auto-upload (unless --local-only)
			if !localOnly {
				fmt.Printf("\n📤 Uploading to cloud...\n")

				uploadResult, err := uploadService.Execute(ctx, output.File.ID, password)
				if err != nil {
					fmt.Printf("⚠️  File added locally but upload failed: %v\n", err)
					fmt.Printf("💡 Upload later with: maplefile-cli files upload %s --password PASSWORD\n", output.File.ID.Hex())
					return
				}

				if uploadResult.Success {
					fmt.Printf("✅ File uploaded successfully!\n")
					fmt.Printf("📤 Uploaded: %s\n", formatFileSize(uploadResult.FileSizeBytes))
					if uploadResult.ThumbnailSizeBytes > 0 {
						fmt.Printf("🖼️  Thumbnail: %s\n", formatFileSize(uploadResult.ThumbnailSizeBytes))
					}
					fmt.Printf("🔄 Status: Synced with cloud\n")
				} else {
					fmt.Printf("⚠️  Upload completed with issues: %v\n", uploadResult.Error)
				}
			} else {
				fmt.Printf("\n📱 File stored locally only (--local-only specified)\n")
				fmt.Printf("💡 Upload later with: maplefile-cli files upload %s --password PASSWORD\n", output.File.ID.Hex())
			}

			// Show next steps
			fmt.Printf("\n🎉 File successfully added to MapleFile!\n")
			fmt.Printf("💡 Next steps:\n")
			fmt.Printf("   • View files: maplefile-cli files list --collection %s\n", collectionID)
			fmt.Printf("   • Download file: maplefile-cli files get %s\n", output.File.ID.Hex())
			if !localOnly {
				fmt.Printf("   • Share collection: maplefile-cli collections share %s --email user@example.com\n", collectionID)
			}

			logger.Info("File added successfully",
				zap.String("fileID", output.File.ID.Hex()),
				zap.String("name", output.File.Name),
				zap.String("collectionID", collectionID),
				zap.Bool("uploaded", !localOnly))
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&collectionID, "collection", "c", "", "Collection ID to store the file in (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Custom name for the file (defaults to filename)")
	cmd.Flags().StringVar(&storageMode, "storage-mode", dom_file.StorageModeHybrid,
		"Storage mode: encrypted_only, hybrid, decrypted_only")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")
	cmd.Flags().BoolVar(&localOnly, "local-only", false, "Add locally without uploading to cloud")

	// Mark required flags
	cmd.MarkFlagRequired("collection")
	cmd.MarkFlagRequired("password")

	return cmd
}

// Helper function for file size formatting
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
