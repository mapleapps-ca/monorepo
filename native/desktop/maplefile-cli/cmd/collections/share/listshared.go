// cmd/collections/sharing.go
package share

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
)

// listSharedCmd creates a command for listing shared collections
func listSharedCmd(
	sharingService collection.SharingService,
	logger *zap.Logger,
) *cobra.Command {
	var verbose bool

	var cmd = &cobra.Command{
		Use:   "list-shared",
		Short: "List collections shared with you",
		Long: `
List all collections that have been shared with you by other users.

This command shows collections where you have been granted access (read_only, read_write, or admin)
but are not the owner.

Examples:
  # List shared collections
  maplefile-cli collections list-shared

  # List shared collections with detailed information
  maplefile-cli collections list-shared --verbose
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Execute list operation
			output, err := sharingService.ListSharedCollections(cmd.Context())
			if err != nil {
				fmt.Printf("🐞 Error listing shared collections: %v\n", err)
				logger.Error("Failed to list shared collections", zap.Error(err))
				return
			}

			// Display results
			if output.Count == 0 {
				fmt.Println("No shared collections found.")
				return
			}

			fmt.Printf("\nFound %d shared collections:\n\n", output.Count)
			for i, coll := range output.Collections {
				displayName := coll.Name
				if displayName == "" || displayName == "[Encrypted]" {
					displayName = "[Encrypted]"
				}

				fmt.Printf("%d. %s (ID: %s, Type: %s)\n",
					i+1, displayName, coll.ID.Hex(), coll.CollectionType)

				if verbose {
					fmt.Printf("   Owner ID: %s\n", coll.OwnerID.Hex())
					fmt.Printf("   Created: %s\n", coll.CreatedAt.Format("2006-01-02 15:04:05"))
					fmt.Printf("   Modified: %s\n", coll.ModifiedAt.Format("2006-01-02 15:04:05"))
					fmt.Printf("   Sync Status: %s\n", coll.SyncStatus.String())
					fmt.Println()
				}
			}
		},
	}

	// Define command flags
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information about each collection")

	return cmd
}
