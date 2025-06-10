// native/desktop/maplefile-cli/cmd/filesync/offload.go
package filesync

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
)

// offloadCmd creates a command for offloading files to cloud storage
func offloadCmd(
	offloadService filesyncer.OffloadService,
	logger *zap.Logger,
) *cobra.Command {
	var fileID string
	var password string

	var cmd = &cobra.Command{
		Use:   "offload",
		Short: "Offload a file to cloud-only storage",
		Long: `
Offload a local file to cloud-only storage. This will:
- Upload the file if it hasn't been uploaded yet
- Remove the local decrypted copy if already uploaded
- Set the file's sync status to cloud-only

Examples:
  maplefile-cli filesync offload --file-id 507f1f77bcf86cd799439011 --password 1234567890
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required fields
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to offload.")
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
			input := &filesyncer.OffloadInput{
				FileID:       fileObjectID,
				UserPassword: password,
			}

			// Execute offload
			fmt.Printf("🔄 Offloading file: %s\n", fileID)

			output, err := offloadService.Offload(cmd.Context(), input)
			if err != nil {
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else {
					fmt.Printf("❌ Error offloading file: %v\n", err)
				}
				return
			}

			// Display success information
			fmt.Printf("\n✅ File successfully offloaded!\n")
			fmt.Printf("🆔 File ID: %s\n", output.FileID.String())
			fmt.Printf("🔄 Action: %s\n", output.Action)
			fmt.Printf("📊 Status: %v → %v\n", output.PreviousStatus, output.NewStatus)
			fmt.Printf("💬 Message: %s\n", output.Message)

			if output.UploadResult != nil && output.UploadResult.Success {
				fmt.Printf("🆔 File ID: %s\n", output.UploadResult.FileID.String())
				fmt.Printf("📏 Uploaded Size: %d bytes\n", output.UploadResult.FileSizeBytes)
			}

			fmt.Printf("\n🎉 Your file is now stored in cloud-only mode!\n")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to offload (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")
	cmd.MarkFlagRequired("password")

	return cmd
}
