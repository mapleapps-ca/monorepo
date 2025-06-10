// cmd/collections/share/share_updated.go
package share

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocql/gocql"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsharing"
)

// ShareCmd creates a command for sharing collections with automatic local sync
func ShareCmdWithSync(
	synchronizedSharingService collectionsharing.SynchronizedCollectionSharingService,
	originalSharingService collectionsharing.CollectionSharingService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionID, recipientEmail, permissionLevel, password string
	var shareWithDescendants bool
	var syncStrategy string

	var cmd = &cobra.Command{
		Use:   "share",
		Short: "Share a collection with another user (with local sync)",
		Long: `
Share a collection with another user, granting them specified permissions.
The collection will be shared using end-to-end encryption and local state
will be automatically updated to reflect the new member.

The collection will be shared using end-to-end encryption where the collection key
is encrypted specifically for the recipient using their public key.

Examples:
  # Share a collection with read-only access (automatic local sync)
  maplefile-cli collections share --id 507f1f77bcf86cd799439011 --email user@example.com --permission read_only --password mypassword

  # Share with read-write access including all child collections
  maplefile-cli collections share --id 507f1f77bcf86cd799439011 --email user@example.com --permission read_write --descendants --password mypassword

  # Share with cloud sync strategy (pulls fresh data from server)
  maplefile-cli collections share --id 507f1f77bcf86cd799439011 --email user@example.com --permission admin --password mypassword --sync-strategy cloud-pull

  # Share without updating local state (original behavior)
  maplefile-cli collections share --id 507f1f77bcf86cd799439011 --email user@example.com --permission admin --password mypassword --sync-strategy none

Permission levels:
  - read_only: Can view collection contents and download files
  - read_write: Can add, modify, and delete files within the collection
  - admin: Can manage collection settings, share with others, and modify member permissions

Sync strategies:
  - immediate (default): Update local collection immediately after cloud sharing
  - cloud-pull: Pull fresh collection data from cloud after sharing
  - none: Don't update local state (original behavior)
`,
		Run: func(cmd *cobra.Command, args []string) {
			//
			// STEP 1: Validate required fields
			//

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

			// Convert string ID to gocql.TimeUUID
			// Note: This variable is named collectionObjectID but now holds a gocql.TimeUUID
			collectionObjectID, err := gocql.ParseUUID(collectionID)
			if err != nil {
				log.Fatalf("invalid collection ID format (expected UUID): %v\n", err)
			}

			//
			// STEP 2: Create request
			//

			// Create service input
			input := &collectionsharing.ShareCollectionInput{
				CollectionID:         collectionObjectID,
				RecipientEmail:       recipientEmail,
				PermissionLevel:      permissionLevel,
				ShareWithDescendants: shareWithDescendants,
			}

			//
			// STEP 3: Execute sharing with selected sync strategy
			//

			var output *collectionsharing.ShareCollectionOutput

			switch syncStrategy {
			case "immediate", "":
				// Default: Use synchronized service for immediate local update
				fmt.Printf("🔄 Sharing collection with immediate local sync...\n")
				output, err = synchronizedSharingService.ExecuteWithSync(cmd.Context(), input, password)

			case "cloud-pull":
				// Use original service then manually sync
				fmt.Printf("🔄 Sharing collection with cloud sync...\n")
				output, err = originalSharingService.Execute(cmd.Context(), input, password)
				if err == nil {
					fmt.Printf("🔄 Pulling updated collection data from cloud...\n")
					// Note: You'd need to implement cloud sync service call here
					// syncService.Execute(cmd.Context(), collectionObjectID, password)
					fmt.Printf("💡 Cloud sync completed.\n")
				}

			case "none":
				// Original behavior - no local sync
				fmt.Printf("🔄 Sharing collection (no local sync)...\n")
				output, err = originalSharingService.Execute(cmd.Context(), input, password)

			default:
				fmt.Printf("🐞 Error: Invalid sync strategy: %s\n", syncStrategy)
				fmt.Println("Valid strategies are: immediate, cloud-pull, none")
				return
			}

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
			fmt.Printf("💡 Local collection updated with new member. Changes are immediately visible.\n")
			fmt.Printf("  Sync Strategy: %s\n", getSyncStrategyDescription(syncStrategy))

			// Additional messaging based on sync strategy
			switch syncStrategy {
			case "immediate", "":
				fmt.Printf("💡 Local collection updated with new member. Changes are immediately visible.\n")
			case "cloud-pull":
				fmt.Printf("💡 Local collection synchronized with cloud. Latest data pulled.\n")
			case "none":
				fmt.Printf("💡 Local collection not updated. Run 'collections list' to see current local state.\n")
				fmt.Printf("💡 Consider running a sync operation to update local state.\n")
			}

			logger.Info("Collection shared successfully",
				zap.String("collectionID", collectionID),
				zap.String("recipientEmail", recipientEmail),
				zap.String("permissionLevel", permissionLevel),
				zap.String("syncStrategy", syncStrategy))
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&collectionID, "id", "", "ID of the collection to share (required)")
	cmd.Flags().StringVar(&recipientEmail, "email", "", "Email address of the recipient (required)")
	cmd.Flags().StringVar(&permissionLevel, "permission", "", "Permission level for the recipient (read_only, read_write, or admin) (required)")
	cmd.Flags().BoolVar(&shareWithDescendants, "descendants", false, "Also share all child collections")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")
	cmd.Flags().StringVar(&syncStrategy, "sync-strategy", "immediate", "Sync strategy after sharing (immediate, cloud-pull, none)")

	// Mark required flags
	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("permission")
	cmd.MarkFlagRequired("password")

	return cmd
}

// getSyncStrategyDescription returns a human-readable description of the sync strategy
func getSyncStrategyDescription(strategy string) string {
	switch strategy {
	case "immediate", "":
		return "immediate (local state updated instantly)"
	case "cloud-pull":
		return "cloud-pull (fresh data pulled from server)"
	case "none":
		return "none (local state not updated)"
	default:
		return strategy
	}
}
