// native/desktop/maplefile-cli/cmd/filesync/cloud_delete.go
package filesync

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
)

// cloudDeleteCmd creates a command for deleting files from cloud storage
func cloudDeleteCmd(
	cloudDeleteService filesyncer.CloudDeleteService,
	logger *zap.Logger,
) *cobra.Command {
	var fileID string
	var password string

	var cmd = &cobra.Command{
		Use:   "cloud-delete",
		Short: "Delete a file from cloud storage",
		Long: `
Delete a file from cloud storage and set the local sync status to local-only.

This command will:
- Delete the file from the cloud backend
- Update the local file sync status to "local-only"
- Keep the local copy of the file intact

The file must be in a synchronized state (synced, cloud-only, or modified locally)
to be eligible for cloud deletion. Local-only files cannot be deleted from the cloud
as they don't exist there.

Examples:
  maplefile-cli filesync cloud-delete --file-id 507f1f77bcf86cd799439011 --password 1234567890
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required fields
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to delete from cloud.")
				return
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required for authentication.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Create service input
			input := &filesyncer.CloudDeleteInput{
				FileID:       fileID,
				UserPassword: password,
			}

			// Execute cloud delete
			fmt.Printf("🔄 Deleting file from cloud: %s\n", fileID)

			output, err := cloudDeleteService.DeleteFromCloud(cmd.Context(), input)
			if err != nil {
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else if strings.Contains(err.Error(), "local-only") {
					fmt.Printf("❌ Error: File is local-only and does not exist in the cloud.\n")
				} else if strings.Contains(err.Error(), "not found") {
					fmt.Printf("❌ Error: File not found. Please check the file ID and try again.\n")
				} else {
					fmt.Printf("❌ Error deleting file from cloud: %v\n", err)
				}
				return
			}

			// Display success information
			fmt.Printf("\n✅ File successfully deleted from cloud!\n")
			fmt.Printf("🆔 File ID: %s\n", output.FileID.Hex())
			fmt.Printf("🔄 Action: %s\n", output.Action)
			fmt.Printf("📊 Status: %v → %v\n", output.PreviousStatus, output.NewStatus)
			fmt.Printf("💬 Message: %s\n", output.Message)

			if output.PreviousStatus == 2 { // SyncStatusModifiedLocally
				fmt.Printf("\n⚠️  Warning: The file had local modifications that are now out of sync with the cloud.\n")
			}

			fmt.Printf("\n🎉 Your file is now in local-only mode!\n")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to delete from cloud (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for authentication)")
	cmd.MarkFlagRequired("password")

	return cmd
}
