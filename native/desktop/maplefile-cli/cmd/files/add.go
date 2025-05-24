// native/desktop/maplefile-cli/cmd/files/add.go
package files

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// addFileCmd creates a command for importing a file into the MapleFile Cloud.
func addFileCmd(
	logger *zap.Logger,
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
  # Create a file with encrypted-only storage (most secure)
  maplefile-cli localfile create --file /path/to/file.pdf --collection 507f1f77bcf86cd799439011

  # Create a file with both encrypted and decrypted copies
  maplefile-cli localfile create --file /path/to/file.pdf --collection 507f1f77bcf86cd799439011 --storage-mode=hybrid

  # Create a file without encryption (not recommended)
  maplefile-cli localfile create --file /path/to/file.pdf --collection 507f1f77bcf86cd799439011 --storage-mode=decrypted_only

  # Create a file with a custom name
  maplefile-cli localfile create --file /path/to/file.pdf --collection 507f1f77bcf86cd799439011 --name "Project Document.pdf"
`,
		Run: func(cmd *cobra.Command, args []string) {

			// Validate required fields
			if filePath == "" {
				fmt.Println("Error: File path is required.")
				fmt.Println("Use --file flag to specify the path to the file.")
				return
			}

			if collectionID == "" {
				fmt.Println("Error: Collection ID is required.")
				fmt.Println("Use --collection flag to specify the collection ID.")
				return
			}

			// Check if file exists
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				fmt.Printf("🐞 Error accessing file: %v\n", err)
				return
			}

			if fileInfo.IsDir() {
				fmt.Println("Error: The specified path is a directory, not a file.")
				return
			}

			// If name not provided, use the original filename
			if name == "" {
				name = fileInfo.Name()
			}

			// Validate storage mode
			if storageMode != dom_file.StorageModeEncryptedOnly &&
				storageMode != dom_file.StorageModeDecryptedOnly &&
				storageMode != dom_file.StorageModeHybrid {
				fmt.Println("Error: Invalid storage mode. Must be 'encrypted_only', 'hybrid', or 'decrypted_only'.")
				return
			}

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
