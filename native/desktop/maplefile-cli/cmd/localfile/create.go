// monorepo/native/desktop/maplefile-cli/cmd/localfile/create.go
package localfile

import (
	"context"
	"fmt"
	"os"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	svc_localfile "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// createLocalFileCmd creates a command for importing a file to local storage only
func createLocalFileCmd(
	importService svc_localfile.ImportService,
	logger *zap.Logger,
) *cobra.Command {
	var filePath string
	var collectionID string
	var name string
	var encrypt bool

	var cmd = &cobra.Command{
		Use:   "create",
		Short: "Create a local file from the filesystem",
		Long: `
Create a local file from the filesystem without uploading to remote server.

This command takes a local file and imports it into the MapleFile system locally,
either encrypting its contents or keeping it as-is based on the --encrypt flag.
You must specify the local file path and the collection ID where the file should be stored.

Examples:
  # Create a local file (encrypted by default)
  maplefile-cli localfile create --file /path/to/file.pdf --collection 507f1f77bcf86cd799439011

  # Create a local file without encryption
  maplefile-cli localfile create --file /path/to/file.pdf --collection 507f1f77bcf86cd799439011 --encrypt=false

  # Create a local file with a custom name
  maplefile-cli localfile create --file /path/to/file.pdf --collection 507f1f77bcf86cd799439011 --name "Project Document.pdf"
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

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

			// Generate a random encrypted file ID
			encryptedFileID, err := generateRandomHexID(16)
			if err != nil {
				fmt.Printf("🐞 Error generating file ID: %v\n", err)
				return
			}

			var encryptedFileKey keys.EncryptedFileKey
			var encryptionVersion string
			var metadata string
			var localFileState string

			if encrypt {
				// Generate a file encryption key for encrypted file
				fileKey, err := crypto.GenerateRandomBytes(crypto.SecretBoxKeySize)
				if err != nil {
					fmt.Printf("🐞 Error generating file key: %v\n", err)
					return
				}

				// Create metadata (in a real implementation, this would be JSON and encrypted)
				// For this example, we'll use a simple base64 encoding to simulate encryption
				metadata = base64Encode(fmt.Sprintf(`{"name":"%s","mime":"%s"}`, name, getMimeType(name)))

				// Simulate encrypting the file key with the collection key
				// In a real implementation, this would use proper E2EE techniques
				ciphertext, nonce, err := crypto.EncryptWithSecretBox(fileKey, fileKey)
				if err != nil {
					fmt.Printf("🐞 Error encrypting file key: %v\n", err)
					return
				}

				encryptedFileKey = keys.EncryptedFileKey{
					Ciphertext: ciphertext,
					Nonce:      nonce,
				}
				encryptionVersion = "1.0"
				localFileState = localfile.LocalFileStateLocalAndEncrypted
			} else {
				// For unencrypted files, we still need some metadata
				// But we don't actually encrypt it
				metadata = base64Encode(fmt.Sprintf(`{"name":"%s","mime":"%s"}`, name, getMimeType(name)))

				// Create empty encryption structures to satisfy the API
				emptyKey := make([]byte, 0)
				encryptedFileKey = keys.EncryptedFileKey{
					Ciphertext: emptyKey,
					Nonce:      emptyKey,
				}
				encryptionVersion = "unencrypted"
				localFileState = localfile.LocalFileStateLocalAndDecrypted
			}

			logger.Debug("Creating local file",
				zap.String("filepath", filePath),
				zap.String("collectionID", collectionID),
				zap.String("name", name),
				zap.Bool("encrypted", encrypt))

			// Prepare import input
			importInput := svc_localfile.ImportInput{
				FilePath:          filePath,
				CollectionID:      collectionID,
				EncryptedFileID:   encryptedFileID,
				EncryptedMetadata: metadata,
				DecryptedName:     name,
				DecryptedMimeType: getMimeType(name),
				EncryptedFileKey:  encryptedFileKey,
				EncryptionVersion: encryptionVersion,
				LocalFileState:    localFileState,
				// You might want to add thumbnail data generation here for certain file types
			}

			// Import the file
			result, err := importService.Import(ctx, importInput)
			if err != nil {
				fmt.Printf("🐞 Error creating local file: %v\n", err)
				return
			}

			// Display success message for import
			fmt.Println("\n✅ File created successfully!")
			fmt.Printf("Local File ID: %s\n", result.File.ID.Hex())
			fmt.Printf("File Name: %s\n", result.File.DecryptedName)
			fmt.Printf("File Path: %s\n", result.File.LocalFilePath)
			if encrypt {
				fmt.Println("File is stored encrypted")
			} else {
				fmt.Println("File is stored unencrypted")
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the file to import (required)")
	cmd.Flags().StringVarP(&collectionID, "collection", "c", "", "ID of the collection to store the file in (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Custom name for the file (defaults to original filename)")
	cmd.Flags().BoolVarP(&encrypt, "encrypt", "e", true, "Encrypt the file during import (defaults to true)")

	// Mark required flags
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("collection")

	return cmd
}
