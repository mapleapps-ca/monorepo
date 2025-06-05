// cmd/collections/delete.go - Clean unified delete/restore commands
package collections

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	svc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
)

// deleteCmd creates a unified command for deleting/archiving collections
func deleteCmd(
	softDeleteService svc_collection.SoftDeleteService,
	logger *zap.Logger,
) *cobra.Command {
	var archive bool
	var withChildren bool
	var force bool

	var cmd = &cobra.Command{
		Use:   "delete COLLECTION_ID",
		Short: "Delete or archive a collection",
		Long: `
Delete or archive a collection.

By default, performs a soft delete (can be restored). Use --archive to archive
instead of deleting.

Examples:
  # Soft delete a collection (can be restored)
  maplefile-cli collections delete 507f1f77bcf86cd799439011

  # Archive a collection (hidden but not deleted)
  maplefile-cli collections delete 507f1f77bcf86cd799439011 --archive

  # Delete a collection and all its children
  maplefile-cli collections delete 507f1f77bcf86cd799439011 --with-children

  # Skip confirmation prompt
  maplefile-cli collections delete 507f1f77bcf86cd799439011 --force

  # Restore a deleted collection
  maplefile-cli collections restore 507f1f77bcf86cd799439011
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			collectionID := args[0]

			// Determine operation type
			var operation string
			var operationIcon string

			if archive {
				operation = "archive"
				operationIcon = "📦"
			} else {
				operation = "soft delete"
				operationIcon = "🗑️"
			}

			// Confirmation prompt (unless --force)
			if !force {
				fmt.Printf("%s You are about to %s collection: %s\n", operationIcon, operation, collectionID)
				if withChildren {
					fmt.Printf("⚠️  This will also %s ALL child collections!\n", operation)
				}
				fmt.Printf("💡 This can be reversed with: maplefile-cli collections restore %s\n", collectionID)

				fmt.Print("\nAre you sure? (y/N): ")
				var response string
				fmt.Scanln(&response)

				if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
					fmt.Println("❌ Operation cancelled.")
					return
				}
			}

			// Execute operation
			var err error

			if archive {
				fmt.Printf("📦 Archiving collection: %s\n", collectionID)
				err = softDeleteService.Archive(cmd.Context(), collectionID)
			} else {
				if withChildren {
					fmt.Printf("🗑️ Soft deleting collection and children: %s\n", collectionID)
					err = softDeleteService.SoftDeleteWithChildren(cmd.Context(), collectionID)
				} else {
					fmt.Printf("🗑️ Soft deleting collection: %s\n", collectionID)
					err = softDeleteService.SoftDelete(cmd.Context(), collectionID)
				}
			}

			if err != nil {
				fmt.Printf("🐞 Error performing %s: %v\n", operation, err)
				if strings.Contains(err.Error(), "invalid state transition") {
					fmt.Printf("💡 The collection may already be deleted/archived.\n")
				} else if strings.Contains(err.Error(), "not found") {
					fmt.Printf("💡 Collection not found. Check the ID and try again.\n")
				} else if strings.Contains(err.Error(), "permission") {
					fmt.Printf("💡 You don't have permission to delete this collection.\n")
				}
				return
			}

			// Success message
			if archive {
				fmt.Printf("✅ Successfully archived collection!\n")
			} else {
				if withChildren {
					fmt.Printf("✅ Successfully soft deleted collection and its children!\n")
				} else {
					fmt.Printf("✅ Successfully soft deleted collection!\n")
				}
			}

			fmt.Printf("🆔 Collection ID: %s\n", collectionID)
			fmt.Printf("💡 To restore: maplefile-cli collections restore %s\n", collectionID)

			logger.Info("Collection deleted/archived successfully",
				zap.String("collectionID", collectionID),
				zap.String("operation", operation),
				zap.Bool("withChildren", withChildren))
		},
	}

	cmd.Flags().BoolVar(&archive, "archive", false, "Archive the collection instead of deleting")
	cmd.Flags().BoolVar(&withChildren, "with-children", false, "Also delete/archive all child collections")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}

// restoreCmd creates a command for restoring deleted/archived collections
func restoreCmd(
	softDeleteService svc_collection.SoftDeleteService,
	logger *zap.Logger,
) *cobra.Command {
	var force bool

	var cmd = &cobra.Command{
		Use:   "restore COLLECTION_ID",
		Short: "Restore a deleted or archived collection",
		Long: `
Restore a collection that was previously soft deleted or archived.
This marks the collection as active again.

Examples:
  # Restore a collection
  maplefile-cli collections restore 507f1f77bcf86cd799439011

  # Restore without confirmation
  maplefile-cli collections restore 507f1f77bcf86cd799439011 --force
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			collectionID := args[0]

			// Confirmation prompt (unless --force)
			if !force {
				fmt.Printf("🔄 You are about to restore collection: %s\n", collectionID)
				fmt.Printf("💡 This will mark the collection as active again.\n")

				fmt.Print("\nContinue? (y/N): ")
				var response string
				fmt.Scanln(&response)

				if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
					fmt.Println("❌ Restore cancelled.")
					return
				}
			}

			fmt.Printf("🔄 Restoring collection: %s\n", collectionID)

			err := softDeleteService.Restore(cmd.Context(), collectionID)
			if err != nil {
				fmt.Printf("🐞 Error restoring collection: %v\n", err)
				if strings.Contains(err.Error(), "invalid state transition") {
					fmt.Printf("💡 The collection may already be active.\n")
				} else if strings.Contains(err.Error(), "not found") {
					fmt.Printf("💡 Collection not found.\n")
				}
				return
			}

			fmt.Printf("✅ Successfully restored collection!\n")
			fmt.Printf("🆔 Collection ID: %s\n", collectionID)
			fmt.Printf("📊 Status: Active\n")

			fmt.Printf("\n💡 Next steps:\n")
			fmt.Printf("   • View collection: maplefile-cli collections list\n")
			fmt.Printf("   • Add files: maplefile-cli files add PATH --collection %s\n", collectionID)

			logger.Info("Collection restored successfully",
				zap.String("collectionID", collectionID))
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
