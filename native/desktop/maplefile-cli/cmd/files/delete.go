// cmd/files/delete.go - Clean unified delete command
package files

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
)

// deleteFileCmd creates a unified command for deleting files
func deleteFileCmd(
	logger *zap.Logger,
	localOnlyDeleteService localfile.LocalOnlyDeleteService,
	cloudOnlyDeleteService filesyncer.CloudOnlyDeleteService,
) *cobra.Command {
	var localOnly bool
	var cloudOnly bool
	var password string
	var force bool

	var cmd = &cobra.Command{
		Use:   "delete FILE_ID",
		Short: "Delete a file",
		Long: `
Delete a file from local storage, cloud storage, or both.

By default, deletes the file completely (both local and cloud copies).
Use flags to control what gets deleted:

  • --local-only: Delete only local copy, keep in cloud
  • --cloud-only: Delete only cloud copy, keep local copy
  • (no flags): Delete completely from both local and cloud

Examples:
  # Delete file completely (both local and cloud)
  maplefile-cli files delete 507f1f77bcf86cd799439011 --password mypass

  # Delete only local copy (keep in cloud for other devices)
  maplefile-cli files delete 507f1f77bcf86cd799439011 --local-only

  # Delete only from cloud (keep local copy)
  maplefile-cli files delete 507f1f77bcf86cd799439011 --cloud-only --password mypass

  # Skip confirmation
  maplefile-cli files delete 507f1f77bcf86cd799439011 --force --password mypass
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fileID := args[0]

			// Validate mutually exclusive flags
			if localOnly && cloudOnly {
				fmt.Println("❌ Error: Cannot use both --local-only and --cloud-only flags.")
				fmt.Println("Choose either local-only deletion, cloud-only deletion, or complete deletion (no flags).")
				return
			}

			// Validate password requirement for cloud operations
			if (cloudOnly || (!localOnly && !cloudOnly)) && password == "" {
				fmt.Println("❌ Error: Password is required for cloud deletion operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Convert to ObjectID for local operations
			var fileObjectID gocql.UUID
			var err error
			if localOnly || (!localOnly && !cloudOnly) {
				fileObjectID, err = primitive.ObjectIDFromHex(fileID)
				if err != nil {
					fmt.Printf("❌ Error: Invalid file ID format: %v\n", err)
					return
				}
			}

			// Determine operation type
			var operation string
			var operationIcon string

			if localOnly {
				operation = "delete local copy"
				operationIcon = "📱"
			} else if cloudOnly {
				operation = "delete cloud copy"
				operationIcon = "☁️"
			} else {
				operation = "delete completely"
				operationIcon = "🗑️"
			}

			// Confirmation prompt (unless --force)
			if !force {
				fmt.Printf("%s You are about to %s of file: %s\n", operationIcon, operation, fileID)

				if localOnly {
					fmt.Printf("💡 The file will remain in the cloud and on other devices.\n")
				} else if cloudOnly {
					fmt.Printf("💡 The file will remain on this device but be removed from cloud.\n")
				} else {
					fmt.Printf("⚠️  This will permanently delete the file from both local and cloud storage!\n")
				}

				fmt.Print("\nAre you sure? (y/N): ")
				var response string
				fmt.Scanln(&response)

				if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
					fmt.Println("❌ Deletion cancelled.")
					return
				}
			}

			// Execute the appropriate deletion operation
			if localOnly {
				// Local-only deletion
				fmt.Printf("📱 Deleting local copy: %s\n", fileID)

				input := &localfile.LocalOnlyDeleteInput{
					ID: fileObjectID,
				}

				err := localOnlyDeleteService.Execute(cmd.Context(), input)
				if err != nil {
					fmt.Printf("❌ Error deleting local file: %v\n", err)
					if strings.Contains(err.Error(), "not found") {
						fmt.Printf("💡 File may not exist locally or already deleted.\n")
					}
					return
				}

				fmt.Printf("✅ Successfully deleted local copy!\n")
				fmt.Printf("🆔 File ID: %s\n", fileID)
				fmt.Printf("☁️ Cloud copy remains available\n")
				fmt.Printf("💡 Download again with: maplefile-cli files get %s\n", fileID)

			} else if cloudOnly {
				// Cloud-only deletion
				fmt.Printf("☁️ Deleting cloud copy: %s\n", fileID)

				input := &filesyncer.CloudOnlyDeleteInput{
					FileID:       fileID,
					UserPassword: password,
				}

				output, err := cloudOnlyDeleteService.DeleteFromCloud(cmd.Context(), input)
				if err != nil {
					fmt.Printf("❌ Error deleting cloud file: %v\n", err)
					if strings.Contains(err.Error(), "incorrect password") {
						fmt.Printf("💡 Check your password and try again.\n")
					} else if strings.Contains(err.Error(), "local-only") {
						fmt.Printf("💡 File is local-only and doesn't exist in the cloud.\n")
					} else if strings.Contains(err.Error(), "not found") {
						fmt.Printf("💡 File not found in cloud storage.\n")
					}
					return
				}

				fmt.Printf("✅ Successfully deleted cloud copy!\n")
				fmt.Printf("🆔 File ID: %s\n", output.FileID.Hex())
				fmt.Printf("📊 Status: %v → %v\n", output.PreviousStatus, output.NewStatus)
				fmt.Printf("📱 Local copy remains available\n")

			} else {
				// Complete deletion (both local and cloud)
				fmt.Printf("🗑️ Deleting file completely: %s\n", fileID)

				// First delete from cloud
				fmt.Printf("Step 1/2: Deleting from cloud...\n")
				cloudInput := &filesyncer.CloudOnlyDeleteInput{
					FileID:       fileID,
					UserPassword: password,
				}

				_, cloudErr := cloudOnlyDeleteService.DeleteFromCloud(cmd.Context(), cloudInput)
				if cloudErr != nil {
					if !strings.Contains(cloudErr.Error(), "local-only") && !strings.Contains(cloudErr.Error(), "not found") {
						fmt.Printf("⚠️  Warning: Failed to delete from cloud: %v\n", cloudErr)
						fmt.Printf("Continuing with local deletion...\n")
					}
				} else {
					fmt.Printf("✅ Deleted from cloud\n")
				}

				// Then delete locally
				fmt.Printf("Step 2/2: Deleting locally...\n")
				localInput := &localfile.LocalOnlyDeleteInput{
					ID: fileObjectID,
				}

				localErr := localOnlyDeleteService.Execute(cmd.Context(), localInput)
				if localErr != nil {
					fmt.Printf("⚠️  Warning: Failed to delete locally: %v\n", localErr)
				} else {
					fmt.Printf("✅ Deleted locally\n")
				}

				// Show final result
				if cloudErr == nil && localErr == nil {
					fmt.Printf("\n✅ File completely deleted!\n")
				} else {
					fmt.Printf("\n⚠️  File partially deleted (see warnings above)\n")
				}
				fmt.Printf("🆔 File ID: %s\n", fileID)
			}

			// Log the operation
			logger.Info("File deletion completed",
				zap.String("fileID", fileID),
				zap.String("operation", operation),
				zap.Bool("localOnly", localOnly),
				zap.Bool("cloudOnly", cloudOnly))
		},
	}

	// Define flags
	cmd.Flags().BoolVar(&localOnly, "local-only", false, "Delete only local copy (keep in cloud)")
	cmd.Flags().BoolVar(&cloudOnly, "cloud-only", false, "Delete only cloud copy (keep local)")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for cloud operations)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
