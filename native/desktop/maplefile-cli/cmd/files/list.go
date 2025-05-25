// native/desktop/maplefile-cli/cmd/files/list.go
package files

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
)

// listFilesCmd creates a command for listing local files by collection
func listFilesCmd(
	logger *zap.Logger,
	listService localfile.ListService,
) *cobra.Command {
	var collectionID string
	var showDetails bool

	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List local files in a collection",
		Long: `
List files stored locally in a specific collection.

Examples:
  # List files in a collection
  maplefile-cli files list --collection 507f1f77bcf86cd799439011

  # List with detailed information
  maplefile-cli files list --collection 507f1f77bcf86cd799439011 --details
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required collection ID
			if collectionID == "" {
				fmt.Println("❌ Error: Collection ID is required.")
				fmt.Println("Use --collection flag to specify the collection ID.")
				return
			}

			// Create service input
			input := &localfile.ListInput{
				CollectionID: collectionID,
			}

			// Execute list service
			fmt.Printf("📂 Listing files in collection: %s\n\n", collectionID)

			output, err := listService.ListByCollection(cmd.Context(), input)
			if err != nil {
				if strings.Contains(err.Error(), "invalid collection ID format") {
					fmt.Printf("❌ Error: Invalid collection ID format. Please check the ID and try again.\n")
				} else {
					fmt.Printf("❌ Error listing files: %v\n", err)
				}
				return
			}

			// Display results
			if output.Count == 0 {
				fmt.Println("📭 No files found in this collection.")
				return
			}

			fmt.Printf("📋 Found %d file(s):\n\n", output.Count)

			if showDetails {
				displayDetailedList(output.Files)
			} else {
				displaySimpleList(output.Files)
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&collectionID, "collection", "c", "", "Collection ID to list files from (required)")
	cmd.MarkFlagRequired("collection")
	cmd.Flags().BoolVarP(&showDetails, "details", "d", false, "Show detailed information")

	return cmd
}

// displaySimpleList shows a compact table of files
func displaySimpleList(files []*dom_file.File) {
	fmt.Printf("%-8s %-30s %-12s %-15s %s\n", "STATUS", "NAME", "SIZE", "SYNC", "ID")
	fmt.Println(strings.Repeat("-", 80))

	for _, file := range files {
		status := getSyncStatusIcon(file.SyncStatus)
		size := formatFileSize(file.FileSize)
		syncStr := getSyncStatusString(file.SyncStatus)

		// Truncate long names
		name := file.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		fmt.Printf("%-8s %-30s %-12s %-15s %s\n",
			status, name, size, syncStr, file.ID.Hex())
	}
}

// displayDetailedList shows comprehensive file information
func displayDetailedList(files []*dom_file.File) {
	for i, file := range files {
		if i > 0 {
			fmt.Println(strings.Repeat("-", 50))
		}

		fmt.Printf("🆔 ID: %s\n", file.ID.Hex())
		fmt.Printf("📄 Name: %s\n", file.Name)
		fmt.Printf("📏 Size: %s (%d bytes)\n", formatFileSize(file.FileSize), file.FileSize)
		fmt.Printf("🏷️  MIME Type: %s\n", file.MimeType)
		fmt.Printf("🔄 Sync Status: %s %s\n", getSyncStatusIcon(file.SyncStatus), getSyncStatusString(file.SyncStatus))
		fmt.Printf("💾 Storage Mode: %s\n", file.StorageMode)
		fmt.Printf("📅 Created: %s\n", file.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("📝 Modified: %s\n", file.ModifiedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("📊 Version: %d\n", file.Version)
		fmt.Println()
	}
}

// Helper functions
func getSyncStatusIcon(status dom_file.SyncStatus) string {
	switch status {
	case dom_file.SyncStatusLocalOnly:
		return "📱"
	case dom_file.SyncStatusCloudOnly:
		return "☁️"
	case dom_file.SyncStatusSynced:
		return "✅"
	case dom_file.SyncStatusModifiedLocally:
		return "📝"
	default:
		return "❓"
	}
}

func getSyncStatusString(status dom_file.SyncStatus) string {
	switch status {
	case dom_file.SyncStatusLocalOnly:
		return "Local Only"
	case dom_file.SyncStatusCloudOnly:
		return "Cloud Only"
	case dom_file.SyncStatusSynced:
		return "Synced"
	case dom_file.SyncStatusModifiedLocally:
		return "Modified"
	default:
		return "Unknown"
	}
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
