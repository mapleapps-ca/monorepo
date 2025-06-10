// cmd/collections/list.go - Clean unified list command
package collections

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	svc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
)

// listCmd creates a unified command for listing collections with various filters
func listCmd(
	listService svc_collection.ListService,
	logger *zap.Logger,
) *cobra.Command {
	var parentID string
	var state string
	var showModified bool
	var verbose bool

	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List collections",
		Long: `
List collections with various filtering options.

By default, lists all active root-level collections. Use flags to filter by:
  • Parent collection (sub-collections)
  • State (active, deleted, archived)
  • Modification status (locally modified)

Examples:
  # List all root collections
  maplefile-cli collections list

  # List sub-collections of a specific parent
  maplefile-cli collections list --parent 507f1f77bcf86cd799439011

  # List collections by state
  maplefile-cli collections list --state active
  maplefile-cli collections list --state deleted
  maplefile-cli collections list --state archived

  # List locally modified collections
  maplefile-cli collections list --modified

  # Detailed listing with verbose output
  maplefile-cli collections list --verbose

  # Combine filters
  maplefile-cli collections list --parent 507f1f77bcf86cd799439011 --verbose
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var output *svc_collection.ListOutput
			var err error
			var filterDescription string

			parenObjectID, err := gocql.ParseUUID(parentID)
			if err != nil {
				log.Fatalf("invalid parent ID format (expected UUID): %v\n", err)
			}

			// Determine which listing method to use
			if showModified {
				filterDescription = "locally modified collections"
				output, err = listService.ListModifiedLocally(ctx)
			} else if state != "" {
				if err := collection.ValidateState(state); err != nil {
					fmt.Printf("🐞 Error: Invalid state '%s'\n", state)
					fmt.Printf("Valid states: %s, %s, %s\n",
						collection.CollectionStateActive,
						collection.CollectionStateDeleted,
						collection.CollectionStateArchived)
					return
				}
				filterDescription = fmt.Sprintf("%s collections", state)

				// Note: This requires extending ListService for state filtering
				if state == collection.CollectionStateActive {
					output, err = listService.ListRoots(ctx)
				} else {
					fmt.Printf("⚠️  State filtering for '%s' requires service layer enhancement.\n", state)
					return
				}
			} else if parentID != "" {
				filterDescription = fmt.Sprintf("sub-collections under parent %s", parenObjectID.String())
				output, err = listService.ListByParent(ctx, parenObjectID)
			} else {
				filterDescription = "root collections"
				output, err = listService.ListRoots(ctx)
			}

			if err != nil {
				fmt.Printf("🐞 Error listing %s: %v\n", filterDescription, err)
				return
			}

			// Display results
			if output.Count == 0 {
				fmt.Printf("📭 No %s found.\n", filterDescription)

				if parentID != "" {
					fmt.Printf("💡 Create a sub-collection: maplefile-cli collections create 'Name' --parent %s\n", parentID)
				} else {
					fmt.Printf("💡 Create your first collection: maplefile-cli collections create 'My Collection'\n")
				}
				return
			}

			fmt.Printf("📋 Found %d %s:\n\n", output.Count, filterDescription)

			if verbose {
				displayDetailedList(output.Collections)
			} else {
				displaySimpleList(output.Collections)
			}

			// Next steps
			fmt.Printf("\n💡 Commands you can try:\n")
			if parentID == "" {
				fmt.Printf("   • View sub-collections: maplefile-cli collections list --parent COLLECTION_ID\n")
			}
			fmt.Printf("   • Create new collection: maplefile-cli collections create 'Collection Name'\n")
			fmt.Printf("   • Add files: maplefile-cli files add PATH --collection COLLECTION_ID\n")
		},
	}

	cmd.Flags().StringVarP(&parentID, "parent", "p", "", "List sub-collections of this parent collection")
	cmd.Flags().StringVarP(&state, "state", "s", "", "Filter by state (active, deleted, archived)")
	cmd.Flags().BoolVarP(&showModified, "modified", "m", false, "Show only locally modified collections")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")

	return cmd
}

// displaySimpleList shows a compact table of collections
func displaySimpleList(collections []*collection.Collection) {
	fmt.Printf("%-8s %-30s %-12s %-15s %s\n", "TYPE", "NAME", "STATE", "SYNC", "ID")
	fmt.Println(strings.Repeat("-", 80))

	for _, coll := range collections {
		typeIcon := getCollectionTypeIcon(coll.CollectionType)
		name := coll.Name
		if name == "" {
			name = "[Encrypted]"
		}
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		state := string(coll.State)
		if len(state) > 10 {
			state = state[:10]
		}

		syncStatus := getSyncStatusString(coll.SyncStatus)
		if len(syncStatus) > 13 {
			syncStatus = syncStatus[:13]
		}

		fmt.Printf("%-8s %-30s %-12s %-15s %s\n",
			typeIcon, name, state, syncStatus, coll.ID.String())
	}
}

// displayDetailedList shows comprehensive collection information
func displayDetailedList(collections []*collection.Collection) {
	for i, coll := range collections {
		if i > 0 {
			fmt.Println(strings.Repeat("-", 50))
		}

		displayName := coll.Name
		if displayName == "" {
			displayName = "[Encrypted]"
		}

		fmt.Printf("🆔 ID: %s\n", coll.ID.String())
		fmt.Printf("📁 Name: %s\n", displayName)
		fmt.Printf("🏷️  Type: %s %s\n", getCollectionTypeIcon(coll.CollectionType), coll.CollectionType)
		fmt.Printf("📊 State: %s\n", coll.State)
		fmt.Printf("🔄 Sync Status: %s\n", getSyncStatusString(coll.SyncStatus))

		if !(coll.ParentID.String() == "") {
			fmt.Printf("📂 Parent ID: %s\n", coll.ParentID.String())
		} else {
			fmt.Printf("📊 Level: Root collection\n")
		}

		fmt.Printf("📅 Created: %s\n", coll.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("📝 Modified: %s\n", coll.ModifiedAt.Format("2006-01-02 15:04:05"))

		fmt.Println()
	}
}

// Helper functions
func getCollectionTypeIcon(collectionType string) string {
	switch collectionType {
	case "folder":
		return "📁"
	case "album":
		return "🖼️"
	default:
		return "📦"
	}
}

func getSyncStatusString(status collection.SyncStatus) string {
	switch status {
	case collection.SyncStatusLocalOnly:
		return "📱 Local Only"
	case collection.SyncStatusCloudOnly:
		return "☁️ Cloud Only"
	case collection.SyncStatusSynced:
		return "✅ Synced"
	case collection.SyncStatusModifiedLocally:
		return "📝 Modified"
	default:
		return "❓ Unknown"
	}
}
