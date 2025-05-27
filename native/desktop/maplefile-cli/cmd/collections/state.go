// cmd/collections/state.go
package collections

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	svc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
)

// softDeleteCmd creates a command for soft deleting a collection
func softDeleteCmd(
	softDeleteService svc_collection.SoftDeleteService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionID string
	var withChildren bool

	var cmd = &cobra.Command{
		Use:   "soft-delete",
		Short: "Soft delete a collection (mark as deleted)",
		Long: `
Soft delete a collection by marking it as deleted instead of permanently removing it.
Soft deleted collections can be restored later.

Examples:
  # Soft delete a collection
  maplefile-cli collections soft-delete --id 507f1f77bcf86cd799439011

  # Soft delete a collection and all its children
  maplefile-cli collections soft-delete --id 507f1f77bcf86cd799439011 --with-children
`,
		Run: func(cmd *cobra.Command, args []string) {
			if collectionID == "" {
				fmt.Println("🐞 Error: Collection ID is required.")
				fmt.Println("Use --id flag to specify the collection ID.")
				return
			}

			var err error
			if withChildren {
				err = softDeleteService.SoftDeleteWithChildren(cmd.Context(), collectionID)
			} else {
				err = softDeleteService.SoftDelete(cmd.Context(), collectionID)
			}

			if err != nil {
				fmt.Printf("🐞 Error soft deleting collection: %v\n", err)
				if strings.Contains(err.Error(), "invalid state transition") {
					fmt.Printf("💡 The collection may already be deleted or in an invalid state for deletion.\n")
				}
				logger.Error("Failed to soft delete collection",
					zap.String("collectionID", collectionID),
					zap.Bool("withChildren", withChildren),
					zap.Error(err))
				return
			}

			if withChildren {
				fmt.Printf("✅ Successfully soft deleted collection and its children!\n")
			} else {
				fmt.Printf("✅ Successfully soft deleted collection!\n")
			}
			fmt.Printf("Collection ID: %s\n", collectionID)
			fmt.Printf("💡 Use 'restore' command to restore the collection if needed.\n")

			logger.Info("Collection soft deleted successfully",
				zap.String("collectionID", collectionID),
				zap.Bool("withChildren", withChildren))
		},
	}

	cmd.Flags().StringVar(&collectionID, "id", "", "ID of the collection to soft delete (required)")
	cmd.Flags().BoolVar(&withChildren, "with-children", false, "Also soft delete all child collections")
	cmd.MarkFlagRequired("id")

	return cmd
}

// archiveCmd creates a command for archiving a collection
func archiveCmd(
	softDeleteService svc_collection.SoftDeleteService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionID string

	var cmd = &cobra.Command{
		Use:   "archive",
		Short: "Archive a collection",
		Long: `
Archive a collection by marking it as archived. Archived collections are
hidden from normal listings but can be restored.

Examples:
  # Archive a collection
  maplefile-cli collections archive --id 507f1f77bcf86cd799439011
`,
		Run: func(cmd *cobra.Command, args []string) {
			if collectionID == "" {
				fmt.Println("🐞 Error: Collection ID is required.")
				fmt.Println("Use --id flag to specify the collection ID.")
				return
			}

			err := softDeleteService.Archive(cmd.Context(), collectionID)
			if err != nil {
				fmt.Printf("🐞 Error archiving collection: %v\n", err)
				if strings.Contains(err.Error(), "invalid state transition") {
					fmt.Printf("💡 The collection may already be archived or in an invalid state.\n")
				}
				logger.Error("Failed to archive collection",
					zap.String("collectionID", collectionID),
					zap.Error(err))
				return
			}

			fmt.Printf("✅ Successfully archived collection!\n")
			fmt.Printf("Collection ID: %s\n", collectionID)
			fmt.Printf("💡 Use 'restore' command to restore the collection if needed.\n")

			logger.Info("Collection archived successfully",
				zap.String("collectionID", collectionID))
		},
	}

	cmd.Flags().StringVar(&collectionID, "id", "", "ID of the collection to archive (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

// restoreCmd creates a command for restoring a collection
func restoreCmd(
	softDeleteService svc_collection.SoftDeleteService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionID string

	var cmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore a deleted or archived collection",
		Long: `
Restore a collection that was previously soft deleted or archived.
This marks the collection as active again.

Examples:
  # Restore a collection
  maplefile-cli collections restore --id 507f1f77bcf86cd799439011
`,
		Run: func(cmd *cobra.Command, args []string) {
			if collectionID == "" {
				fmt.Println("🐞 Error: Collection ID is required.")
				fmt.Println("Use --id flag to specify the collection ID.")
				return
			}

			err := softDeleteService.Restore(cmd.Context(), collectionID)
			if err != nil {
				fmt.Printf("🐞 Error restoring collection: %v\n", err)
				if strings.Contains(err.Error(), "invalid state transition") {
					fmt.Printf("💡 The collection may already be active or cannot be restored from its current state.\n")
				}
				logger.Error("Failed to restore collection",
					zap.String("collectionID", collectionID),
					zap.Error(err))
				return
			}

			fmt.Printf("✅ Successfully restored collection!\n")
			fmt.Printf("Collection ID: %s\n", collectionID)

			logger.Info("Collection restored successfully",
				zap.String("collectionID", collectionID))
		},
	}

	cmd.Flags().StringVar(&collectionID, "id", "", "ID of the collection to restore (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

// listByStateCmd creates a command for listing collections by state
func listByStateCmd(
	listService svc_collection.ListService,
	logger *zap.Logger,
) *cobra.Command {
	var state string
	var verbose bool

	var cmd = &cobra.Command{
		Use:   "list-by-state",
		Short: "List collections by state",
		Long: `
List collections filtered by their state (active, deleted, archived).

Examples:
  # List active collections
  maplefile-cli collections list-by-state --state active

  # List deleted collections
  maplefile-cli collections list-by-state --state deleted

  # List archived collections with verbose output
  maplefile-cli collections list-by-state --state archived --verbose
`,
		Run: func(cmd *cobra.Command, args []string) {
			if state == "" {
				fmt.Println("🐞 Error: State is required.")
				fmt.Printf("Valid states: %s, %s, %s\n",
					collection.CollectionStateActive,
					collection.CollectionStateDeleted,
					collection.CollectionStateArchived)
				return
			}

			// Validate state
			if err := collection.ValidateState(state); err != nil {
				fmt.Printf("🐞 Error: Invalid state '%s'\n", state)
				fmt.Printf("Valid states: %s, %s, %s\n",
					collection.CollectionStateActive,
					collection.CollectionStateDeleted,
					collection.CollectionStateArchived)
				return
			}

			// For now, we'll use the existing list service and filter logic
			// In a real implementation, you'd want to extend the list service
			// to support state-based listing directly
			var output *svc_collection.ListOutput
			var err error

			switch state {
			case collection.CollectionStateActive:
				output, err = listService.ListRoots(cmd.Context())
			case collection.CollectionStateDeleted:
				// This would need to be implemented in the service layer
				fmt.Printf("🔧 Listing deleted collections is not yet implemented in the list service.\n")
				fmt.Printf("💡 This requires extending the ListService interface to support state filtering.\n")
				return
			case collection.CollectionStateArchived:
				// This would need to be implemented in the service layer
				fmt.Printf("🔧 Listing archived collections is not yet implemented in the list service.\n")
				fmt.Printf("💡 This requires extending the ListService interface to support state filtering.\n")
				return
			}

			if err != nil {
				fmt.Printf("🐞 Error listing collections by state: %v\n", err)
				return
			}

			// Display collections
			if output.Count == 0 {
				fmt.Printf("No %s collections found.\n", state)
				return
			}

			fmt.Printf("\nFound %d %s collections:\n\n", output.Count, state)
			for i, coll := range output.Collections {
				displayName := coll.Name
				if displayName == "" {
					displayName = "[Encrypted]"
				}

				fmt.Printf("%d. %s (ID: %s, Type: %s, State: %s)\n",
					i+1, displayName, coll.ID.Hex(), coll.CollectionType, coll.State)

				if verbose {
					fmt.Printf("   Created: %s\n", coll.CreatedAt.Format("2006-01-02 15:04:05"))
					fmt.Printf("   Modified: %s\n", coll.ModifiedAt.Format("2006-01-02 15:04:05"))
					if !coll.ParentID.IsZero() {
						fmt.Printf("   Parent ID: %s\n", coll.ParentID.Hex())
					}
					fmt.Println()
				}
			}
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "State to filter by (active, deleted, archived) (required)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information about each collection")
	cmd.MarkFlagRequired("state")

	return cmd
}
