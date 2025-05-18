// monorepo/native/desktop/maplefile-cli/cmd/remote/listcollections.go
package remote

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/remotecollection"
)

// RemoteListCollectionsCmd creates a command for listing remote collections
func RemoteListCollectionsCmd(
	configService config.ConfigService,
	remoteListService remotecollection.ListService,
	logger *zap.Logger,
) *cobra.Command {
	var parentID string
	var verbose bool

	var cmd = &cobra.Command{
		Use:   "list-collections",
		Short: "List remote collections",
		Long: `
List collections stored in the cloud.

By default, this command lists root-level collections. Use the --parent flag
to list sub-collections of a specific parent collection.

Examples:
  # List root collections from the cloud
  maplefile-cli remote list-collections

  # List sub-collections of a specific parent from the cloud
  maplefile-cli remote list-collections --parent 507f1f77bcf86cd799439011

  # Show detailed information with verbose mode
  maplefile-cli remote list-collections --verbose
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Check if user is authenticated
			email, err := configService.GetEmail(ctx)
			if err != nil || email == "" {
				fmt.Println("❌ You must be logged in to list remote collections. Please login first.")
				return
			}

			var output *remotecollection.ListOutput

			if parentID != "" {
				// List collections under a specific parent
				logger.Debug("Listing remote sub-collections", zap.String("parentID", parentID))
				output, err = remoteListService.ListByParent(ctx, parentID)
				if err != nil {
					fmt.Printf("❌ Error listing remote sub-collections: %v\n", err)
					return
				}
			} else {
				// List root collections
				logger.Debug("Listing remote root collections")
				output, err = remoteListService.ListRoots(ctx)
				if err != nil {
					fmt.Printf("❌ Error listing remote root collections: %v\n", err)
					return
				}
			}

			// Display collections
			if output.Count == 0 {
				fmt.Println("No remote collections found.")
				return
			}

			fmt.Printf("\nFound %d remote collections:\n\n", output.Count)
			for i, collection := range output.Collections {
				// Display collection name - if decrypted name is available, use it
				// Otherwise show as encrypted
				displayName := collection.DecryptedName
				if displayName == "" {
					displayName = "[Encrypted Name]"
				}

				fmt.Printf("%d. %s (ID: %s, Type: %s)\n", i+1, displayName, collection.ID.Hex(), collection.Type)

				if verbose {
					fmt.Printf("   Created: %s\n", collection.CreatedAt.Format("2006-01-02 15:04:05"))
					fmt.Printf("   Modified: %s\n", collection.ModifiedAt.Format("2006-01-02 15:04:05"))
					if !collection.ParentID.IsZero() {
						fmt.Printf("   Parent ID: %s\n", collection.ParentID.Hex())
					}
					if len(collection.AncestorIDs) > 0 {
						fmt.Printf("   Ancestor IDs: ")
						for j, id := range collection.AncestorIDs {
							if j > 0 {
								fmt.Print(", ")
							}
							fmt.Print(id.Hex())
						}
						fmt.Println()
					}
					fmt.Println()
				}
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&parentID, "parent", "p", "", "ID of the parent collection")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information about each collection")

	return cmd
}
