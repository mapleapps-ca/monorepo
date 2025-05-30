// cmd/collections/share/unshare.go
package share

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsharing"
)

// UnshareCmd creates a command for removing collection members
func UnshareCmd(
	removeMemberService collectionsharing.CollectionSharingRemoveMembersService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionID, recipientEmail string
	var removeFromDescendants bool

	var cmd = &cobra.Command{
		Use:   "unshare",
		Short: "Remove a user's access to a collection",
		Long: `
Remove a user's access to a collection.

This command revokes a user's access to the specified collection. If the
--descendants flag is used, it will also remove access from all child collections.

Examples:
  # Remove access to a specific collection
  maplefile-cli collections unshare --id 507f1f77bcf86cd799439011 --email user@example.com

  # Remove access to a collection and all its child collections
  maplefile-cli collections unshare --id 507f1f77bcf86cd799439011 --email user@example.com --descendants
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required fields
			if collectionID == "" {
				fmt.Println("🐞 Error: Collection ID is required.")
				fmt.Println("Use --id flag to specify the collection ID.")
				return
			}

			if recipientEmail == "" {
				fmt.Println("🐞 Error: Recipient email is required.")
				fmt.Println("Use --email flag to specify the email address of the user to remove.")
				return
			}

			// Create service input
			input := &collectionsharing.RemoveMemberInput{
				CollectionID:          collectionID,
				RecipientEmail:        recipientEmail,
				RemoveFromDescendants: removeFromDescendants,
			}

			// Execute remove operation
			_, err := removeMemberService.Execute(cmd.Context(), input)
			if err != nil {
				fmt.Printf("🐞 Error removing collection member: %v\n", err)
				if strings.Contains(err.Error(), "permission") {
					fmt.Printf("❌ Error: You don't have permission to remove members from this collection.\n")
				} else if strings.Contains(err.Error(), "not found") {
					fmt.Printf("❌ Error: Collection or user not found.\n")
				} else if strings.Contains(err.Error(), "does not have access") {
					fmt.Printf("❌ Error: User does not have access to this collection.\n")
				} else if strings.Contains(err.Error(), "cannot remove the collection owner") {
					fmt.Printf("❌ Error: Cannot remove the collection owner.\n")
				} else {
					logger.Error("Failed to remove collection member",
						zap.String("collectionID", collectionID),
						zap.String("recipientEmail", recipientEmail),
						zap.Error(err))
				}
				return
			}

			// Display success message
			fmt.Printf("✅ Successfully removed collection access!\n\n")
			fmt.Printf("Remove Details:\n")
			fmt.Printf("  Collection ID: %s\n", collectionID)
			fmt.Printf("  User Removed: %s\n", recipientEmail)
			fmt.Printf("  Removed from Descendants: %t\n", removeFromDescendants)

			logger.Info("Collection member removed successfully",
				zap.String("collectionID", collectionID),
				zap.String("recipientEmail", recipientEmail))
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&collectionID, "id", "", "ID of the collection (required)")
	cmd.Flags().StringVar(&recipientEmail, "email", "", "Email address of the user to remove (required)")
	cmd.Flags().BoolVar(&removeFromDescendants, "descendants", false, "Also remove access from all child collections")

	// Mark required flags
	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("email")

	return cmd
}
