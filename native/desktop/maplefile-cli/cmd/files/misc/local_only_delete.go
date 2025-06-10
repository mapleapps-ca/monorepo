// native/desktop/maplefile-cli/cmd/files/misc/local_only_delete.go
package misc

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocql/gocql"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
)

// localOnlyDeleteFilesCmd creates a command for localOnlyDeleteing local files by file ID
func localOnlyDeleteFilesCmd(
	logger *zap.Logger,
	localOnlyDeleteService localfile.LocalOnlyDeleteService,
) *cobra.Command {
	var fileID string

	var cmd = &cobra.Command{
		Use:   "local-only-delete",
		Short: "Deletes a local file",
		Long: `
		Deletes a local file (and it's content).

Examples:
  # Delete a file
  maplefile-cli files local-only-delete --id 507f1f77bcf86cd799439011
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required file ID
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file ID.")
				return
			}

			fileIDObj, err := gocql.ParseUUID(fileID)
			if err != nil {
				log.Fatalf("Failed converting string-to-object-id: %v\n", err)
			}

			// Create service input
			input := &localfile.LocalOnlyDeleteInput{
				ID: fileIDObj,
			}

			if err := localOnlyDeleteService.Execute(cmd.Context(), input); err != nil {
				if strings.Contains(err.Error(), "invalid ID format") {
					fmt.Printf("❌ Error: Invalid file ID format. Please check the ID and try again.\n")
				} else {
					fmt.Printf("❌ Error localOnlyDeleteing files: %v\n", err)
				}
				return
			}

			fmt.Printf("✅ Deleted local file %s\n\n", fileID)
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "File ID (required)")
	cmd.MarkFlagRequired("file-id")

	return cmd
}
