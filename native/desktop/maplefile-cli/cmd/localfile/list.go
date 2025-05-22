// monorepo/native/desktop/maplefile-cli/cmd/localfile/list.go
package localfile

import (
	"context"
	"fmt"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	svc_localfile "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// listLocalFileCmd creates a command for listing local files
func listLocalFileCmd(
	listService svc_localfile.ListService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionID string
	var nameContains string
	var mimeType string
	var modifiedOnly bool
	var format string

	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List local files",
		Long: `
List local files stored in the MapleFile system.

This command displays information about files stored locally, including their names,
sizes, storage modes, and sync status. You can filter the results using various options.

Examples:
  # List all local files
  maplefile-cli localfile list

  # List files in a specific collection
  maplefile-cli localfile list --collection 507f1f77bcf86cd799439011

  # List files with names containing "document"
  maplefile-cli localfile list --name-contains document

  # List PDF files only
  maplefile-cli localfile list --mime-type application/pdf

  # List only locally modified files
  maplefile-cli localfile list --modified-only

  # Output in JSON format
  maplefile-cli localfile list --format json

  # Combine filters
  maplefile-cli localfile list --collection 507f1f77bcf86cd799439011 --name-contains report --format table
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate format
			if format != "table" && format != "json" && format != "simple" {
				fmt.Println("Error: Invalid format. Must be 'table', 'json', or 'simple'.")
				return
			}

			var output *svc_localfile.ListOutput
			var err error

			// Execute appropriate list operation based on flags
			switch {
			case modifiedOnly:
				output, err = listService.ListModifiedLocally(ctx)
			case collectionID != "":
				output, err = listService.ListByCollection(ctx, collectionID)
			case nameContains != "" || mimeType != "":
				output, err = listService.Search(ctx, nameContains, mimeType)
			default:
				output, err = listService.ListAll(ctx)
			}

			if err != nil {
				fmt.Printf("🐞 Error listing files: %v\n", err)
				return
			}

			if output.Count == 0 {
				fmt.Println("No files found matching the criteria.")
				return
			}

			// Display results based on format
			switch format {
			case "json":
				displayJSON(output)
			case "simple":
				displaySimple(output)
			default: // table
				displayTable(output)
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVar(&collectionID, "collection", "", "Filter by collection ID")
	cmd.Flags().StringVar(&nameContains, "name-contains", "", "Filter by files containing this text in the name")
	cmd.Flags().StringVar(&mimeType, "mime-type", "", "Filter by MIME type")
	cmd.Flags().BoolVar(&modifiedOnly, "modified-only", false, "Show only locally modified files")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: 'table', 'json', or 'simple'")

	return cmd
}

// displayTable displays files in a formatted table
func displayTable(output *svc_localfile.ListOutput) {
	fmt.Printf("\n📁 Found %d local file(s):\n\n", output.Count)

	// Header
	fmt.Printf("%-8s %-30s %-12s %-15s %-12s %-20s\n",
		"STATUS", "NAME", "SIZE", "STORAGE MODE", "MIME TYPE", "MODIFIED")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────────")

	for _, file := range output.Files {
		// Status indicator
		status := getSyncStatusIcon(file.SyncStatus)

		// Format file size
		sizeStr := formatFileSize(file.FileSize)

		// Truncate long names
		name := file.DecryptedName
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		// Truncate mime type
		mime := file.DecryptedMimeType
		if len(mime) > 10 {
			mime = mime[:7] + "..."
		}

		// Format modified time
		modifiedStr := file.ModifiedAt.Format("2006-01-02 15:04")

		fmt.Printf("%-8s %-30s %-12s %-15s %-12s %-20s\n",
			status, name, sizeStr, file.StorageMode, mime, modifiedStr)
	}

	fmt.Println()
	displaySummary(output)
}

// displaySimple displays files in a simple list format
func displaySimple(output *svc_localfile.ListOutput) {
	fmt.Printf("Found %d file(s):\n\n", output.Count)

	for _, file := range output.Files {
		fmt.Printf("%s %s (%s)\n",
			getSyncStatusIcon(file.SyncStatus),
			file.DecryptedName,
			formatFileSize(file.FileSize))
	}
}

// displayJSON displays files in JSON format
func displayJSON(output *svc_localfile.ListOutput) {
	// Create a simplified structure for JSON output
	type jsonFile struct {
		ID           string    `json:"id"`
		Name         string    `json:"name"`
		Size         int64     `json:"size"`
		MimeType     string    `json:"mime_type"`
		StorageMode  string    `json:"storage_mode"`
		SyncStatus   string    `json:"sync_status"`
		CollectionID string    `json:"collection_id"`
		ModifiedAt   time.Time `json:"modified_at"`
		IsModified   bool      `json:"is_modified_locally"`
	}

	type jsonOutput struct {
		Count int        `json:"count"`
		Files []jsonFile `json:"files"`
	}

	jsonFiles := make([]jsonFile, len(output.Files))
	for i, file := range output.Files {
		jsonFiles[i] = jsonFile{
			ID:           file.ID.Hex(),
			Name:         file.DecryptedName,
			Size:         file.FileSize,
			MimeType:     file.DecryptedMimeType,
			StorageMode:  file.StorageMode,
			SyncStatus:   getSyncStatusString(file.SyncStatus),
			CollectionID: file.CollectionID.Hex(),
			ModifiedAt:   file.ModifiedAt,
			IsModified:   file.IsModifiedLocally,
		}
	}

	result := jsonOutput{
		Count: output.Count,
		Files: jsonFiles,
	}

	// Simple JSON marshaling (in a real app, you might want to use json.MarshalIndent)
	fmt.Printf(`{
  "count": %d,
  "files": [`, result.Count)

	for i, file := range result.Files {
		if i > 0 {
			fmt.Print(",")
		}
		fmt.Printf(`
    {
      "id": "%s",
      "name": "%s",
      "size": %d,
      "mime_type": "%s",
      "storage_mode": "%s",
      "sync_status": "%s",
      "collection_id": "%s",
      "modified_at": "%s",
      "is_modified_locally": %t
    }`, file.ID, file.Name, file.Size, file.MimeType, file.StorageMode,
			file.SyncStatus, file.CollectionID, file.ModifiedAt.Format(time.RFC3339), file.IsModified)
	}

	fmt.Print(`
  ]
}
`)
}

// displaySummary shows a summary of the file listing
func displaySummary(output *svc_localfile.ListOutput) {
	var totalSize int64
	statusCounts := make(map[localfile.SyncStatus]int)
	modeCounts := make(map[string]int)

	for _, file := range output.Files {
		totalSize += file.FileSize
		statusCounts[file.SyncStatus]++
		modeCounts[file.StorageMode]++
	}

	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   Total files: %d\n", output.Count)
	fmt.Printf("   Total size: %s\n", formatFileSize(totalSize))

	if len(statusCounts) > 0 {
		fmt.Printf("   By sync status:\n")
		for status, count := range statusCounts {
			fmt.Printf("     %s %s: %d\n", getSyncStatusIcon(status), getSyncStatusString(status), count)
		}
	}

	if len(modeCounts) > 0 {
		fmt.Printf("   By storage mode:\n")
		for mode, count := range modeCounts {
			fmt.Printf("     %s: %d\n", mode, count)
		}
	}
}

// getSyncStatusIcon returns an icon for the sync status
func getSyncStatusIcon(status localfile.SyncStatus) string {
	switch status {
	case localfile.SyncStatusLocalOnly:
		return "📱"
	case localfile.SyncStatusCloudOnly:
		return "☁️"
	case localfile.SyncStatusSynced:
		return "✅"
	case localfile.SyncStatusModifiedLocally:
		return "🔄"
	default:
		return "❓"
	}
}

// getSyncStatusString returns a string representation of the sync status
func getSyncStatusString(status localfile.SyncStatus) string {
	switch status {
	case localfile.SyncStatusLocalOnly:
		return "local_only"
	case localfile.SyncStatusCloudOnly:
		return "cloud_only"
	case localfile.SyncStatusSynced:
		return "synced"
	case localfile.SyncStatusModifiedLocally:
		return "modified_locally"
	default:
		return "unknown"
	}
}

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

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
