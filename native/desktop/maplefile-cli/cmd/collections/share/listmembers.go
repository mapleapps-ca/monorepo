// cmd/collections/sharing.go
package share

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
)

// membersCmd creates a command for listing collection members
func membersCmd(
	sharingService collection.SharingService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionID string
	var verbose bool

	var cmd = &cobra.Command{
		Use:   "members",
		Short: "List members of a collection",
		Long: `
List all members who have access to a specific collection.

This command shows all users who have been granted access to the collection,
including their permission levels and when access was granted.

Examples:
  # List collection members
  maplefile-cli collections members --id 507f1f77bcf86cd799439011

  # List collection members with detailed information
  maplefile-cli collections members --id 507f1f77bcf86cd799439011 --verbose
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required fields
			if collectionID == "" {
				fmt.Println("🐞 Error: Collection ID is required.")
				fmt.Println("Use --id flag to specify the collection ID.")
				return
			}

			// Execute get members operation
			members, err := sharingService.GetCollectionMembers(cmd.Context(), collectionID)
			if err != nil {
				fmt.Printf("🐞 Error getting collection members: %v\n", err)
				if strings.Contains(err.Error(), "not found") {
					fmt.Printf("❌ Error: Collection not found.\n")
				} else if strings.Contains(err.Error(), "permission") {
					fmt.Printf("❌ Error: You don't have permission to view this collection.\n")
				} else {
					logger.Error("Failed to get collection members",
						zap.String("collectionID", collectionID),
						zap.Error(err))
				}
				return
			}

			// Display results
			if len(members) == 0 {
				fmt.Println("No members found for this collection.")
				return
			}

			fmt.Printf("\nCollection members (%d):\n\n", len(members))
			for i, member := range members {
				fmt.Printf("%d. %s (%s)\n",
					i+1, member.RecipientEmail, member.PermissionLevel)

				if verbose {
					fmt.Printf("   Member ID: %s\n", member.ID.Hex())
					fmt.Printf("   Recipient ID: %s\n", member.RecipientID.Hex())
					fmt.Printf("   Granted By: %s\n", member.GrantedByID.Hex())
					fmt.Printf("   Is Inherited: %t\n", member.IsInherited)
					if member.IsInherited && !member.InheritedFromID.IsZero() {
						fmt.Printf("   Inherited From: %s\n", member.InheritedFromID.Hex())
					}
					fmt.Printf("   Created: %s\n", member.CreatedAt.Format("2006-01-02 15:04:05"))
					fmt.Println()
				}
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&collectionID, "id", "", "ID of the collection (required)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information about each member")

	// Mark required flags
	cmd.MarkFlagRequired("id")

	return cmd
}
