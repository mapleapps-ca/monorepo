// monorepo/native/desktop/maplefile-cli/cmd/localfile/delete.go
package localfile

import (
	"context"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/remotefile"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// deleteLocalFileCmd creates a command for deleting a local file
func deleteLocalFileCmd(
	deleteService localfile.DeleteService,
	getService localfile.GetService,
	remoteFetchService remotefile.FetchService,
	logger *zap.Logger,
) *cobra.Command {
	var fileID string
	var force bool
	var forceIgnoreRemote bool

	var cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a local file",
		Long: `
Delete a local file from the MapleFile system.

This command permanently removes a file from local storage.
By default, files with corresponding remote copies will not be deleted to prevent sync issues.
Use --force-ignore-remote to override this behavior (not recommended).

Examples:
  # Delete a file
  maplefile-cli localfile delete --file-id 507f1f77bcf86cd799439011

  # Force deletion without confirmation
  maplefile-cli localfile delete --file-id 507f1f77bcf86cd799439011 --force

  # Delete even if remote file exists (not recommended)
  maplefile-cli localfile delete --file-id 507f1f77bcf86cd799439011 --force-ignore-remote
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate input
			if fileID == "" {
				fmt.Println("Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file ID.")
				return
			}

			// Convert string ID to ObjectID
			id, err := primitive.ObjectIDFromHex(fileID)
			if err != nil {
				fmt.Printf("🐞 Error: Invalid file ID format: %v\n", err)
				return
			}

			// Get the file to check its remote ID
			if !forceIgnoreRemote {
				fileOutput, err := getService.GetByID(ctx, id)
				if err != nil {
					fmt.Printf("🐞 Error retrieving file: %v\n", err)
					return
				}

				if fileOutput == nil || fileOutput.File == nil {
					fmt.Println("Error: File not found.")
					return
				}

				// Check if it has a remote ID
				if !fileOutput.File.RemoteID.IsZero() {
					// Check if the remote file still exists
					remoteOutput, err := remoteFetchService.FetchByID(ctx, fileOutput.File.RemoteID.Hex())
					if err == nil && remoteOutput != nil && remoteOutput.File != nil {
						// Remote file exists, refuse to delete local copy
						fmt.Println("Error: This file has a remote copy that still exists.")
						fmt.Println("Deleting the local copy would cause sync issues.")
						fmt.Println("Please delete the remote file first, or use --force-ignore-remote (not recommended).")
						return
					}
				}
			}

			// Handle confirmation unless forced
			if !force {
				confirmMessage := fmt.Sprintf("Are you sure you want to delete file %s? This action cannot be undone (y/n): ", fileID)
				if !confirmAction(confirmMessage) {
					fmt.Println("Operation cancelled.")
					return
				}
			}

			// Delete the file
			result, err := deleteService.Delete(ctx, fileID)
			if err != nil {
				fmt.Printf("🐞 Error deleting file: %v\n", err)
				return
			}

			fmt.Println("\n✅ File deleted successfully!")
			if result != nil && result.Message != "" {
				fmt.Println(result.Message)
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&fileID, "file-id", "", "ID of the file to delete (required)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&forceIgnoreRemote, "force-ignore-remote", false, "Delete local file even if remote copy exists (not recommended)")

	// Mark required flags
	cmd.MarkFlagRequired("file-id")

	return cmd
}

// confirmAction asks for user confirmation and returns true if the user confirms
func confirmAction(message string) bool {
	var response string
	fmt.Print(message)
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}
