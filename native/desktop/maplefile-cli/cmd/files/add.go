// native/desktop/maplefile-cli/cmd/files/add.go
package files

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
)

// addFileCmd creates a command for importing a file into the MapleFile Cloud.
func addFileCmd(
	logger *zap.Logger,
	addService localfile.AddService,
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
			// Step 1: Validate required flags
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

			// Convert collection ID string to ObjectID
			collectionObjectID, err := primitive.ObjectIDFromHex(collectionID)
			if err != nil {
				fmt.Printf("❌ Error: Invalid collection ID format: %v\n", err)
				return
			}

			// Set default storage mode if not provided
			if storageMode == "" {
				storageMode = dom_file.StorageModeEncryptedOnly
			}

			// ========================================
			// Step 2: Prepare service input
			// ========================================
			// TODO: Get actual owner ID from authenticated user
			// For now, using a placeholder ObjectID
			ownerID := primitive.NewObjectID()

			input := &localfile.AddInput{
				FilePath:     filePath,
				CollectionID: collectionObjectID,
				OwnerID:      ownerID,
				Name:         name,
				StorageMode:  storageMode,
			}

			// ========================================
			// Step 3: Execute the service
			// ========================================
			fmt.Printf("🔄 Processing file: %s\n", filePath)

			output, err := addService.Add(ctx, input)
			if err != nil {
				fmt.Printf("❌ Error adding file: %v\n", err)
				return
			}

			// ========================================
			// Step 4: Display success information
			// ========================================
			fmt.Printf("\n✅ File successfully added to MapleFile!\n")
			fmt.Printf("🆔 File ID: %s\n", output.File.ID.Hex())
			fmt.Printf("📁 File Name: %s\n", output.File.Name)
			fmt.Printf("📏 File Size: %d bytes\n", output.File.FileSize)
			fmt.Printf("🗂️  MIME Type: %s\n", output.File.MimeType)
			fmt.Printf("🔐 Storage Mode: %s\n", output.File.StorageMode)
			fmt.Printf("📂 Collection ID: %s\n", output.File.CollectionID.Hex())
			fmt.Printf("💾 Copied to: %s\n", output.CopiedFilePath)
			fmt.Printf("🔄 Sync Status: %s\n", output.File.SyncStatus)
			fmt.Printf("📅 Created: %s\n", output.File.CreatedAt.Format("2006-01-02 15:04:05"))

			fmt.Printf("\n🎉 Your file is now safely stored in MapleFile and ready for synchronization!\n")
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
