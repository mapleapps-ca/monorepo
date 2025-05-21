// monorepo/native/desktop/maplefile-cli/cmd/localfile/delete.go
package localfile

import (
	"context"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// deleteLocalFileCmd creates a command for deleting a local file
func deleteLocalFileCmd(
	deleteService localfile.DeleteService,
	logger *zap.Logger,
) *cobra.Command {
	var fileID string
	var deleteAllInCollection bool
	var collectionID string
	var force bool

	var cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a local file",
		Long: `
Delete a local file from the MapleFile system.

This command permanently removes a file from local storage. You can either:
- Delete a single file by specifying its ID
- Delete all files in a collection by using the --all-in-collection flag with a collection ID

Examples:
  # Delete a single file
  maplefile-cli localfile delete --file-id 507f1f77bcf86cd799439011

  # Delete all files in a collection
  maplefile-cli localfile delete --all-in-collection --collection 507f1f77bcf86cd799439011

  # Force deletion without confirmation
  maplefile-cli localfile delete --file-id 507f1f77bcf86cd799439011 --force
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate input combinations
			if fileID == "" && !deleteAllInCollection {
				fmt.Println("Error: Either --file-id or --all-in-collection must be specified.")
				return
			}

			if fileID != "" && deleteAllInCollection {
				fmt.Println("Error: Cannot specify both --file-id and --all-in-collection. Choose one.")
				return
			}

			if deleteAllInCollection && collectionID == "" {
				fmt.Println("Error: Collection ID is required when using --all-in-collection.")
				fmt.Println("Use --collection flag to specify the collection ID.")
				return
			}

			// Handle confirmation unless forced
			if !force {
				var confirmMessage string
				if fileID != "" {
					confirmMessage = fmt.Sprintf("Are you sure you want to delete file %s? This action cannot be undone (y/n): ", fileID)
				} else {
					confirmMessage = fmt.Sprintf("Are you sure you want to delete ALL files in collection %s? This action cannot be undone (y/n): ", collectionID)
				}

				if !confirmAction(confirmMessage) {
					fmt.Println("Operation cancelled.")
					return
				}
			}

			var result *localfile.DeleteOutput
			var err error

			// Perform deletion based on flags
			if fileID != "" {
				// Delete single file
				result, err = deleteService.Delete(ctx, fileID)
				if err != nil {
					fmt.Printf("🐞 Error deleting file: %v\n", err)
					return
				}

				fmt.Println("\n✅ File deleted successfully!")
			} else if deleteAllInCollection {
				// Delete all files in collection
				result, err = deleteService.DeleteByCollection(ctx, collectionID)
				if err != nil {
					fmt.Printf("🐞 Error deleting files in collection: %v\n", err)
					return
				}

				fmt.Printf("\n✅ Successfully deleted %d files from collection!\n", result.DeletedCount)
			}

			if result != nil {
				fmt.Println(result.Message)
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&fileID, "file-id", "", "ID of the file to delete")
	cmd.Flags().BoolVar(&deleteAllInCollection, "all-in-collection", false, "Delete all files in a collection")
	cmd.Flags().StringVar(&collectionID, "collection", "", "ID of the collection to delete files from (used with --all-in-collection)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}

// confirmAction asks for user confirmation and returns true if the user confirms
func confirmAction(message string) bool {
	var response string
	fmt.Print(message)
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}
