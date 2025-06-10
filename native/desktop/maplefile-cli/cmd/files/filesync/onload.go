// native/desktop/maplefile-cli/cmd/filesync/onload.go
package filesync

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
)

// onloadCmd creates a command for onloading files from cloud storage
func onloadCmd(
	onloadService filesyncer.OnloadService,
	logger *zap.Logger,
) *cobra.Command {
	var fileID string
	var password string

	var cmd = &cobra.Command{
		Use:   "onload",
		Short: "Onload a cloud-only file to local storage",
		Long: `
Onload a cloud-only file to local storage. This will:
- Download the encrypted file from cloud storage
- Decrypt the file using your encryption keys
- Save the decrypted file locally
- Set the file's sync status to synced

The file must be in cloud-only status to be eligible for onloading.
This command uses the integrated download service to handle all E2EE
decryption automatically.

Examples:
  maplefile-cli filesync onload --file-id 507f1f77bcf86cd799439011 --password 1234567890
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required fields
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to onload.")
				return
			}

			// Convert to ObjectID
			fileObjectID, err := gocql.ParseUUID(fileID)
			if err != nil {
				fmt.Printf("❌ Error: Invalid file ID format: %v\n", err)
				return
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Create service input
			input := &filesyncer.OnloadInput{
				FileID:       fileObjectID,
				UserPassword: password,
			}

			// Execute onload
			fmt.Printf("🔄 Onloading file: %s\n", fileID)
			fmt.Println("📡 Downloading and decrypting file from cloud...")

			output, err := onloadService.Onload(cmd.Context(), input)
			if err != nil {
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else if strings.Contains(err.Error(), "not cloud-only") {
					fmt.Printf("❌ Error: File is not in cloud-only mode. Only cloud-only files can be onloaded.\n")
				} else if strings.Contains(err.Error(), "file not found") {
					fmt.Printf("❌ Error: File not found. Please check the file ID and try again.\n")
				} else if strings.Contains(err.Error(), "permission") {
					fmt.Printf("❌ Error: You don't have permission to access this file.\n")
				} else {
					fmt.Printf("❌ Error onloading file: %v\n", err)
				}
				return
			}

			// Display success information
			fmt.Printf("\n✅ File successfully onloaded!\n")
			fmt.Printf("🆔 File ID: %s\n", output.FileID.String())
			fmt.Printf("📊 Status: %v → %v\n", output.PreviousStatus, output.NewStatus)
			fmt.Printf("💾 Local Path: %s\n", output.DecryptedPath)
			fmt.Printf("📏 Downloaded Size: %d bytes\n", output.DownloadedSize)
			fmt.Printf("💬 Message: %s\n", output.Message)

			fmt.Printf("\n🎉 Your file is now available locally!\n")
			fmt.Printf("🔐 The file has been downloaded and decrypted using E2EE.\n")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to onload (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")
	cmd.MarkFlagRequired("password")

	return cmd
}
