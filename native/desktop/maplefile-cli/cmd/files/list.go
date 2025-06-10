// cmd/files/list.go - Clean unified list command
package files

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
)

// listFilesCmd creates a command for listing files with various filters
func listFilesCmd(
	logger *zap.Logger,
	listService localfile.ListService,
) *cobra.Command {
	var collectionID string
	var verbose bool

	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List files in collections",
		Long: `
List files stored in your collections.

By default, lists all files across all collections. Use --collection to filter
by a specific collection.

Examples:
  # List all files across all collections
  maplefile-cli files list

  # List files in a specific collection
  maplefile-cli files list --collection 507f1f77bcf86cd799439011

  # List with detailed information
  maplefile-cli files list --collection 507f1f77bcf86cd799439011 --verbose
`,
		Run: func(cmd *cobra.Command, args []string) {
			if collectionID != "" {
				// Convert collection ID
				collectionObjectID, err := gocql.ParseUUID(collectionID)
				if err != nil {
					fmt.Printf("❌ Error: Invalid collection ID format: %v\n", err)
					return
				}

				// List files in specific collection
				input := &localfile.ListInput{
					CollectionID: collectionObjectID,
				}

				fmt.Printf("📂 Listing files in collection: %s\n\n", collectionObjectID.String())

				output, err := listService.ListByCollection(cmd.Context(), input)
				if err != nil {
					fmt.Printf("❌ Error listing files: %v\n", err)
					if strings.Contains(err.Error(), "invalid collection ID format") {
						fmt.Printf("💡 Tip: Check the collection ID format.\n")
					} else if strings.Contains(err.Error(), "collection not found") {
						fmt.Printf("💡 Tip: Check collection exists with: maplefile-cli collections list\n")
					}
					return
				}

				displayFileResults(output.Files, output.Count, verbose, collectionID)
			} else {
				// List all files (would need service enhancement)
				fmt.Printf("📋 Listing all files across collections...\n\n")
				fmt.Printf("⚠️  Listing all files requires service enhancement.\n")
				fmt.Printf("💡 For now, specify a collection: maplefile-cli files list --collection COLLECTION_ID\n")
				fmt.Printf("💡 View collections: maplefile-cli collections list\n")
				return
			}
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&collectionID, "collection", "c", "", "Collection ID to list files from")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed file information")

	return cmd
}

// displayFileResults shows file listing results
func displayFileResults(files []*dom_file.File, count int, verbose bool, collectionID string) {
	if count == 0 {
		fmt.Println("📭 No files found in this collection.")
		fmt.Printf("\n💡 Add your first file:\n")
		fmt.Printf("   maplefile-cli files add FILE_PATH --collection %s --password PASSWORD\n", collectionID)
		return
	}

	fmt.Printf("📋 Found %d file(s):\n\n", count)

	if verbose {
		displayDetailedFileList(files)
	} else {
		displaySimpleFileList(files)
	}

	// Show helpful next steps
	fmt.Printf("\n💡 Commands you can try:\n")
	fmt.Printf("   • Download file: maplefile-cli files get FILE_ID\n")
	fmt.Printf("   • Add more files: maplefile-cli files add FILE_PATH --collection %s\n", collectionID)
	fmt.Printf("   • Delete file: maplefile-cli files delete FILE_ID\n")
}

// displaySimpleFileList shows a compact table of files
func displaySimpleFileList(files []*dom_file.File) {
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
			status, name, size, syncStr, file.ID.String())
	}
}

// displayDetailedFileList shows comprehensive file information
func displayDetailedFileList(files []*dom_file.File) {
	for i, file := range files {
		if i > 0 {
			fmt.Println(strings.Repeat("-", 50))
		}

		fmt.Printf("🆔 ID: %s\n", file.ID.String())
		fmt.Printf("📄 Name: %s\n", file.Name)
		fmt.Printf("📏 Size: %s (%d bytes)\n", formatFileSize(file.FileSize), file.FileSize)
		fmt.Printf("🏷️  MIME Type: %s\n", file.MimeType)
		fmt.Printf("🔄 Sync Status: %s %s\n", getSyncStatusIcon(file.SyncStatus), getSyncStatusString(file.SyncStatus))
		fmt.Printf("💾 Storage Mode: %s\n", file.StorageMode)
		fmt.Printf("📅 Created: %s\n", file.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("📝 Modified: %s\n", file.ModifiedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("📊 Version: %d\n", file.Version)

		// Show storage details based on storage mode
		switch file.StorageMode {
		case dom_file.StorageModeEncryptedOnly:
			fmt.Printf("🔐 Storage: Encrypted only (most secure)\n")
		case dom_file.StorageModeHybrid:
			fmt.Printf("🔐 Storage: Hybrid (encrypted + decrypted)\n")
		case dom_file.StorageModeDecryptedOnly:
			fmt.Printf("📄 Storage: Decrypted only (not encrypted)\n")
		}

		fmt.Println()
	}
}

// Helper functions for file status display
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
