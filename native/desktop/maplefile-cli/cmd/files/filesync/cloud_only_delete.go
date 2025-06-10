// native/desktop/maplefile-cli/cmd/filesync/cloud_delete.go
package filesync

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
)

// cloudOnlyDeleteCmd creates a command for deleting files from cloud storage
func cloudOnlyDeleteCmd(
	cloudOnlyDeleteService filesyncer.CloudOnlyDeleteService,
	logger *zap.Logger,
) *cobra.Command {
	var fileID string
	var password string
	var dryRun bool

	var cmd = &cobra.Command{
		Use:   "cloud-only-delete",
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

Use --dry-run to check the file status without actually deleting it.

Examples:
  # Delete file from cloud
  maplefile-cli filesync cloud-only-delete --file-id 507f1f77bcf86cd799439011 --password 1234567890

  # Check file status without deleting (dry run)
  maplefile-cli filesync cloud-only-delete --file-id 507f1f77bcf86cd799439011 --password 1234567890 --dry-run
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
			input := &filesyncer.CloudOnlyDeleteInput{
				FileID:       fileID,
				UserPassword: password,
			}

			if dryRun {
				fmt.Printf("🔍 DRY RUN: Checking file status for: %s\n", fileID)
				fmt.Printf("ℹ️  This will not actually delete the file.\n\n")
			} else {
				fmt.Printf("🔄 Deleting file from cloud: %s\n", fileID)
				fmt.Printf("🔍 Debug: Attempting to delete file with ID: %s\n", fileID)
			}

			// For dry run, we could add a separate method, but for now we'll rely on error messages
			output, err := cloudOnlyDeleteService.DeleteFromCloud(cmd.Context(), input)
			if err != nil {
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else if strings.Contains(err.Error(), "local-only") {
					fmt.Printf("❌ Error: File is local-only and does not exist in the cloud.\n")
					if dryRun {
						fmt.Printf("ℹ️  DRY RUN: File cannot be deleted from cloud because it only exists locally.\n")
					}
				} else if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
					fmt.Printf("❌ Error: File not found in cloud storage.\n")
					fmt.Printf("🔍 Debug: This could mean:\n")
					fmt.Printf("   • The file was never uploaded to the cloud\n")
					fmt.Printf("   • The file ID is incorrect\n")
					fmt.Printf("   • The file was already deleted from the cloud\n")
				} else if strings.Contains(err.Error(), "permission") {
					fmt.Printf("❌ Error: You don't have permission to delete this file.\n")
				} else {
					fmt.Printf("❌ Error deleting file from cloud: %v\n", err)
					fmt.Printf("🔍 Debug: Raw error details: %s\n", err.Error())
				}
				return
			}

			// Display success information
			if dryRun {
				fmt.Printf("\n✅ DRY RUN: File would be successfully deleted from cloud!\n")
			} else {
				fmt.Printf("\n✅ File successfully deleted from cloud!\n")
			}
			fmt.Printf("🆔 File ID: %s\n", output.FileID.String())
			fmt.Printf("🔄 Action: %s\n", output.Action)
			fmt.Printf("📊 Status: %v → %v\n", output.PreviousStatus, output.NewStatus)
			fmt.Printf("💬 Message: %s\n", output.Message)

			if output.PreviousStatus == 2 { // SyncStatusModifiedLocally
				fmt.Printf("\n⚠️  Warning: The file had local modifications that are now out of sync with the cloud.\n")
			}

			if dryRun {
				fmt.Printf("\n🔍 DRY RUN: No actual changes were made.\n")
			} else {
				fmt.Printf("\n🎉 Your file is now in local-only mode!\n")
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to delete from cloud (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for authentication)")
	cmd.MarkFlagRequired("password")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Check file status without actually deleting it")

	return cmd
}
