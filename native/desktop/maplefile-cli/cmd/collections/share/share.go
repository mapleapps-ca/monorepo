// cmd/collections/share/share.go
package share

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsharing"
)

// ShareCmd creates a command for sharing collections
func ShareCmd(
	sharingService collectionsharing.CollectionSharingService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionID, recipientEmail, permissionLevel, password string
	var shareWithDescendants bool

	var cmd = &cobra.Command{
		Use:   "share",
		Short: "Share a collection with another user",
		Long: `
Share a collection with another user, granting them specified permissions.

The collection will be shared using end-to-end encryption where the collection key
is encrypted specifically for the recipient using their public key.

Examples:
  # Share a collection with read-only access
  maplefile-cli collections share --id 507f1f77bcf86cd799439011 --email user@example.com --permission read_only --password mypassword

  # Share a collection with read-write access including all child collections
  maplefile-cli collections share --id 507f1f77bcf86cd799439011 --email user@example.com --permission read_write --descendants --password mypassword

  # Share a collection with admin access
  maplefile-cli collections share --id 507f1f77bcf86cd799439011 --email user@example.com --permission admin --password mypassword

Permission levels:
  - read_only: Can view collection contents and download files
  - read_write: Can add, modify, and delete files within the collection
  - admin: Can manage collection settings, share with others, and modify member permissions
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
				fmt.Println("Use --email flag to specify the recipient's email address.")
				return
			}

			if permissionLevel == "" {
				fmt.Println("🐞 Error: Permission level is required.")
				fmt.Println("Use --permission flag to specify the permission level (read_only, read_write, or admin).")
				return
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Validate permission level
			if err := collectionsharingdto.ValidatePermissionLevel(permissionLevel); err != nil {
				fmt.Printf("🐞 Error: Invalid permission level: %s\n", permissionLevel)
				fmt.Println("Valid permission levels are: read_only, read_write, admin")
				return
			}

			// Create service input
			input := &collectionsharing.ShareCollectionInput{
				CollectionID:         collectionID,
				RecipientEmail:       recipientEmail,
				PermissionLevel:      permissionLevel,
				ShareWithDescendants: shareWithDescendants,
			}

			// Execute share operation
			output, err := sharingService.Execute(cmd.Context(), input, password)
			if err != nil {
				fmt.Printf("🐞 Error sharing collection: %v\n", err)
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else if strings.Contains(err.Error(), "permission") {
					fmt.Printf("❌ Error: You don't have permission to share this collection.\n")
				} else if strings.Contains(err.Error(), "not found") {
					fmt.Printf("❌ Error: Collection or recipient user not found.\n")
				} else if strings.Contains(err.Error(), "already has access") {
					fmt.Printf("❌ Error: Recipient already has access to this collection.\n")
				} else {
					logger.Error("Failed to share collection",
						zap.String("collectionID", collectionID),
						zap.String("recipientEmail", recipientEmail),
						zap.Error(err))
				}
				return
			}

			// Display success message
			fmt.Printf("✅ Successfully shared collection!\n\n")
			fmt.Printf("Share Details:\n")
			fmt.Printf("  Collection ID: %s\n", collectionID)
			fmt.Printf("  Recipient: %s\n", recipientEmail)
			fmt.Printf("  Permission Level: %s\n", permissionLevel)
			fmt.Printf("  Shared with Descendants: %t\n", shareWithDescendants)
			fmt.Printf("  Memberships Created: %d\n", output.MembershipsCreated)

			logger.Info("Collection shared successfully",
				zap.String("collectionID", collectionID),
				zap.String("recipientEmail", recipientEmail),
				zap.String("permissionLevel", permissionLevel))
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&collectionID, "id", "", "ID of the collection to share (required)")
	cmd.Flags().StringVar(&recipientEmail, "email", "", "Email address of the recipient (required)")
	cmd.Flags().StringVar(&permissionLevel, "permission", "", "Permission level for the recipient (read_only, read_write, or admin) (required)")
	cmd.Flags().BoolVar(&shareWithDescendants, "descendants", false, "Also share all child collections")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")

	// Mark required flags
	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("permission")
	cmd.MarkFlagRequired("password")

	return cmd
}
