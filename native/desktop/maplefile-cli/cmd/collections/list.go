// cmd/collections/list.go
package collections

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localcollection"
)

// listCollectionsCmd creates a command for listing collections
func listCollectionsCmd(
	listService localcollection.ListService,
	logger *zap.Logger,
) *cobra.Command {
	var parentID string
	var showModified bool
	var verbose bool

	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List collections",
		Long: `
List collections stored locally.

By default, this command lists root-level collections. Use the --parent flag
to list sub-collections of a specific parent collection.

Examples:
  # List root collections
  maplefile-cli collections list

  # List sub-collections of a specific parent
  maplefile-cli collections list --parent 507f1f77bcf86cd799439011

  # List locally modified collections
  maplefile-cli collections list --modified

  # Show detailed information with verbose mode
  maplefile-cli collections list --verbose
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var output *localcollection.ListOutput
			var err error

			if showModified {
				// List modified collections
				logger.Debug("Listing locally modified collections")
				output, err = listService.ListModifiedLocally(ctx)
				if err != nil {
					fmt.Printf("🐞 Error listing modified collections: %v\n", err)
					return
				}
			} else if parentID != "" {
				// List collections under a specific parent
				logger.Debug("Listing sub-collections", zap.String("parentID", parentID))
				output, err = listService.ListByParent(ctx, parentID)
				if err != nil {
					fmt.Printf("🐞 Error listing sub-collections: %v\n", err)
					return
				}
			} else {
				// List root collections
				logger.Debug("Listing root collections")
				output, err = listService.ListRoots(ctx)
				if err != nil {
					fmt.Printf("🐞 Error listing root collections: %v\n", err)
					return
				}
			}

			// Display collections
			if output.Count == 0 {
				fmt.Println("No collections found.")
				return
			}

			fmt.Printf("\nFound %d collections:\n\n", output.Count)
			for i, collection := range output.Collections {
				displayName := collection.DecryptedName
				if displayName == "" {
					displayName = "[Encrypted]"
				}

				fmt.Printf("%d. %s (Local ID: %s, Remote ID: %s, Type: %s)\n", i+1, displayName, collection.ID.Hex(), collection.RemoteID.Hex(), collection.Type)

				if verbose {
					fmt.Printf("   Created: %s\n", collection.CreatedAt.Format("2006-01-02 15:04:05"))
					fmt.Printf("   Modified: %s\n", collection.ModifiedAt.Format("2006-01-02 15:04:05"))
					if !collection.RemoteID.IsZero() {
						fmt.Printf("   Remote ID: %s\n", collection.RemoteID.Hex())
					}
					if !collection.ParentID.IsZero() {
						fmt.Printf("   Parent ID: %s\n", collection.ParentID.Hex())
					}
					if collection.IsModifiedLocally {
						fmt.Printf("   Status: Modified locally\n")
					} else if !collection.LastSyncedAt.IsZero() {
						fmt.Printf("   Last Synced: %s\n", collection.LastSyncedAt.Format("2006-01-02 15:04:05"))
					}
					fmt.Println()
				}
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&parentID, "parent", "p", "", "ID of the parent collection")
	cmd.Flags().BoolVarP(&showModified, "modified", "m", false, "Show only locally modified collections")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information about each collection")

	return cmd
}
