// native/desktop/maplefile-cli/cmd/files/lock.go
package files

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
)

// lockFileCmd creates a command for locking a file (encrypted-only mode)
func lockFileCmd(
	logger *zap.Logger,
	lockService localfile.LockService,
) *cobra.Command {
	var fileID string

	var cmd = &cobra.Command{
		Use:   "lock",
		Short: "Lock a file to encrypted-only mode",
		Long: `
Lock a file to encrypted-only mode by deleting the local decrypted version
and keeping only the encrypted version. This provides maximum security as
the decrypted content is not stored on disk.

The file must have a local encrypted version available. Cloud-only files
cannot be locked as they don't have local encrypted versions.

Examples:
  # Lock a file to encrypted-only mode
  maplefile-cli files lock --file-id 507f1f77bcf86cd799439011
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required inputs
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to lock.")
				return
			}

			// Create service input
			input := &localfile.LockInput{
				FileID: fileID,
			}

			// Execute lock operation
			fmt.Printf("🔒 Locking file: %s\n", fileID)
			fmt.Println("🔄 Switching to encrypted-only mode...")

			output, err := lockService.Lock(cmd.Context(), input)
			if err != nil {
				if strings.Contains(err.Error(), "cloud-only") {
					fmt.Printf("❌ Error: Cannot lock cloud-only files. Use 'filesync onload' first to download the file locally.\n")
				} else if strings.Contains(err.Error(), "file not found") {
					fmt.Printf("❌ Error: File not found. Please check the file ID.\n")
				} else if strings.Contains(err.Error(), "no encrypted file") {
					fmt.Printf("❌ Error: No encrypted version available. Cannot lock file without encrypted version.\n")
				} else if strings.Contains(err.Error(), "does not exist on disk") {
					fmt.Printf("❌ Error: Encrypted file missing from disk. File may be corrupted.\n")
				} else {
					fmt.Printf("❌ Error locking file: %v\n", err)
				}
				return
			}

			// Display success information
			fmt.Printf("\n✅ File successfully locked!\n")
			fmt.Printf("🆔 File ID: %s\n", output.FileID.Hex())
			fmt.Printf("📊 Storage Mode: %s → %s\n", output.PreviousMode, output.NewMode)
			fmt.Printf("🗑️  Deleted: %s\n", getPathDisplayName(output.DeletedPath))
			fmt.Printf("💾 Remaining: %s\n", getPathDisplayName(output.RemainingPath))
			fmt.Printf("💬 Status: %s\n", output.Message)

			fmt.Printf("\n🔒 Your file is now in encrypted-only mode for maximum security!\n")
			fmt.Printf("🔓 Use 'maplefile-cli files unlock' to access the decrypted content when needed.\n")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to lock (required)")
	cmd.MarkFlagRequired("file-id")

	return cmd
}

// Helper function to display path names nicely
func getPathDisplayName(path string) string {
	if path == "" {
		return "(none)"
	}
	// Extract just the filename for cleaner display
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}
