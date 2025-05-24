// native/desktop/maplefile-cli/cmd/files/add.go
package files

import (
	"context"
	"fmt"
	"mime"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// addFileCmd creates a command for importing a file into the MapleFile Cloud.
func addFileCmd(
	logger *zap.Logger,
	readFileUseCase localfile.ReadFileUseCase,
	checkFileExistsUseCase localfile.CheckFileExistsUseCase,
	getFileInfoUseCase localfile.GetFileInfoUseCase,
	pathUtilsUseCase localfile.PathUtilsUseCase,
) *cobra.Command {
	var filePath string
	var collectionID string
	var name string
	var storageMode string

	var cmd = &cobra.Command{
		Use:   "add",
		Short: "Add a file into your account",
		Long: `
Adds a file from the filesystem into MapleFile Cloud.

This command takes a local file and imports it into the MapleFile Cloud.
You can control how the file is stored using the --storage-mode flag:

* "encrypted_only": Only keep the encrypted version (most secure)
* "hybrid": Keep both encrypted and decrypted versions (convenient)
* "decrypted_only": Keep only the decrypted version (not recommended, no encryption)

Examples:
  # Windows examples
  maplefile-cli files add --file "C:\Users\John\Documents\report.pdf" --collection 507f1f77bcf86cd799439011
  maplefile-cli files add --file "D:\Projects\MyApp\config.json" --collection 507f1f77bcf86cd799439011 --storage-mode=hybrid

  # Linux examples
  maplefile-cli files add --file "/home/john/documents/report.pdf" --collection 507f1f77bcf86cd799439011
  maplefile-cli files add --file "/var/log/application.log" --collection 507f1f77bcf86cd799439011 --name "App Log"

  # macOS examples
  maplefile-cli files add --file "/Users/john/Desktop/presentation.pptx" --collection 507f1f77bcf86cd799439011
  maplefile-cli files add --file "/Applications/MyApp/data.db" --collection 507f1f77bcf86cd799439011 --storage-mode=encrypted_only
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// ========================================
			// Step 1: Validate and clean the file path
			// ========================================
			if filePath == "" {
				fmt.Println("❌ Error: File path is required.")
				fmt.Println("Use --file flag to specify the path to the file.")
				return
			}

			if collectionID == "" {
				fmt.Println("❌ Error: Collection ID is required.")
				fmt.Println("Use --collection flag to specify the collection ID.")
				return
			}

			// Clean the path (handles . and .. and redundant separators)
			cleanFilePath := pathUtilsUseCase.Clean(ctx, filePath)
			logger.Debug("Cleaned file path",
				zap.String("original", filePath),
				zap.String("cleaned", cleanFilePath))

			// ========================================
			// Step 2: Check if file exists (cross-platform)
			// ========================================
			fmt.Printf("🔍 Checking if file exists: %s\n", cleanFilePath)

			exists, err := checkFileExistsUseCase.Execute(ctx, cleanFilePath)
			if err != nil {
				fmt.Printf("❌ Error checking file existence: %v\n", err)
				return
			}

			if !exists {
				fmt.Printf("❌ Error: File does not exist: %s\n", cleanFilePath)
				return
			}

			// ========================================
			// Step 3: Get file information (cross-platform)
			// ========================================
			fmt.Printf("📋 Getting file information...\n")

			fileInfo, err := getFileInfoUseCase.Execute(ctx, cleanFilePath)
			if err != nil {
				fmt.Printf("❌ Error getting file info: %v\n", err)
				return
			}

			if fileInfo.IsDirectory {
				fmt.Println("❌ Error: The specified path is a directory, not a file.")
				return
			}

			// Display file information
			fmt.Printf("📁 File: %s\n", fileInfo.Name)
			fmt.Printf("📏 Size: %d bytes\n", fileInfo.Size)
			fmt.Printf("📅 Modified: %s\n", fileInfo.ModifiedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("🔐 Permissions: %s\n", fileInfo.Permissions)

			// ========================================
			// Step 4: Read file content (cross-platform)
			// ========================================
			fmt.Printf("📖 Reading file content...\n")

			fileContent, err := readFileUseCase.Execute(ctx, cleanFilePath)
			if err != nil {
				fmt.Printf("❌ Error reading file: %v\n", err)
				return
			}

			fmt.Printf("✅ Successfully read %d bytes from file\n", len(fileContent))

			// ========================================
			// Step 5: Process file metadata
			// ========================================

			// If name not provided, use the original filename
			if name == "" {
				name = pathUtilsUseCase.GetFileName(ctx, cleanFilePath)
			}

			// Detect MIME type from file extension
			fileExtension := pathUtilsUseCase.GetFileExtension(ctx, cleanFilePath)
			mimeType := mime.TypeByExtension(fileExtension)
			if mimeType == "" {
				mimeType = "application/octet-stream" // Default for unknown types
			}

			// Validate storage mode
			if storageMode != dom_file.StorageModeEncryptedOnly &&
				storageMode != dom_file.StorageModeDecryptedOnly &&
				storageMode != dom_file.StorageModeHybrid {
				fmt.Println("❌ Error: Invalid storage mode. Must be 'encrypted_only', 'hybrid', or 'decrypted_only'.")
				return
			}

			// ========================================
			// Step 6: Display processing summary
			// ========================================
			fmt.Printf("\n📊 File Processing Summary:\n")
			fmt.Printf("  Original Path: %s\n", filePath)
			fmt.Printf("  Clean Path: %s\n", cleanFilePath)
			fmt.Printf("  File Name: %s\n", name)
			fmt.Printf("  File Size: %d bytes\n", fileInfo.Size)
			fmt.Printf("  MIME Type: %s\n", mimeType)
			fmt.Printf("  Storage Mode: %s\n", storageMode)
			fmt.Printf("  Collection ID: %s\n", collectionID)

			// ========================================
			// Step 7: Cross-platform path examples
			// ========================================
			fmt.Printf("\n🌐 Cross-Platform Path Info:\n")

			// Check if path is absolute
			isAbsolute := pathUtilsUseCase.IsAbsolute(ctx, cleanFilePath)
			fmt.Printf("  Is Absolute Path: %t\n", isAbsolute)

			// Get directory and filename separately
			directory := pathUtilsUseCase.GetDirectory(ctx, cleanFilePath)
			filename := pathUtilsUseCase.GetFileName(ctx, cleanFilePath)
			fmt.Printf("  Directory: %s\n", directory)
			fmt.Printf("  Filename: %s\n", filename)

			// Show path in different formats
			unixStylePath := pathUtilsUseCase.ToSlash(ctx, cleanFilePath)
			fmt.Printf("  Unix-style path: %s\n", unixStylePath)

			// ========================================
			// Step 8: Create domain file object (placeholder)
			// ========================================
			// Here you would create your domain file object and save it
			// This is where you'd integrate with your file creation use case

			collectionObjectID, err := primitive.ObjectIDFromHex(collectionID)
			if err != nil {
				fmt.Printf("❌ Error: Invalid collection ID format: %v\n", err)
				return
			}

			// Example of creating a domain file object
			domainFile := &dom_file.File{
				ID:           primitive.NewObjectID(),
				CollectionID: collectionObjectID,
				OwnerID:      primitive.NewObjectID(), // You'd get this from current user
				Name:         name,
				MimeType:     mimeType,
				FilePath:     cleanFilePath,
				FileSize:     fileInfo.Size,
				StorageMode:  storageMode,
				CreatedAt:    time.Now(),
				ModifiedAt:   time.Now(),
				SyncStatus:   dom_file.SyncStatusLocalOnly,
				// You would set encrypted fields here based on your encryption logic
			}

			fmt.Printf("\n✅ File successfully processed and ready for import!\n")
			fmt.Printf("🆔 Generated File ID: %s\n", domainFile.ID.Hex())

			// TODO: Integrate with your file creation use case here
			// err = createFileUseCase.Execute(ctx, domainFile)
			// if err != nil {
			//     fmt.Printf("❌ Error creating file: %v\n", err)
			//     return
			// }
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the file to import (required)")
	cmd.Flags().StringVarP(&collectionID, "collection", "c", "", "ID of the collection to store the file in (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Custom name for the file (defaults to original filename)")
	cmd.Flags().StringVar(&storageMode, "storage-mode", dom_file.StorageModeEncryptedOnly,
		"Storage mode: 'encrypted_only', 'hybrid', or 'decrypted_only'")

	// Mark required flags
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("collection")

	return cmd
}
