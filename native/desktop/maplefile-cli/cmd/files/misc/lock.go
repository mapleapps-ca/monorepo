// native/desktop/maplefile-cli/cmd/files/misc/lock.go
package misc

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
	var password string

	var cmd = &cobra.Command{
		Use:   "lock",
		Short: "Lock a file to encrypted-only mode using E2EE",
		Long: `
Lock a file to encrypted-only mode by encrypting the decrypted content and
removing the decrypted version from disk. This provides maximum security as
only the encrypted version remains on disk.

This operation uses end-to-end encryption (E2EE) with the complete key chain:
password → key encryption key → master key → collection key → file key

The file must have a local decrypted version available for encryption.
Cloud-only files cannot be locked as they don't have local versions.

Examples:
  # Lock a file to encrypted-only mode
  maplefile-cli files lock --file-id 507f1f77bcf86cd799439011 --password 1234567890
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required inputs
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to lock.")
				return
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Create service input
			input := &localfile.LockInput{
				FileID:   fileID,
				Password: password,
			}

			// Execute lock operation
			fmt.Printf("🔒 Locking file: %s\n", fileID)
			fmt.Println("🔄 Switching to encrypted-only mode...")

			output, err := lockService.Lock(cmd.Context(), input)
			if err != nil {
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else if strings.Contains(err.Error(), "cloud-only") {
					fmt.Printf("❌ Error: Cannot lock cloud-only files. Use 'filesync onload' first to download the file locally.\n")
				} else if strings.Contains(err.Error(), "file not found") {
					fmt.Printf("❌ Error: File not found. Please check the file ID.\n")
				} else if strings.Contains(err.Error(), "no decrypted file") {
					fmt.Printf("❌ Error: No decrypted version available. Cannot lock file without decrypted version.\n")
				} else if strings.Contains(err.Error(), "does not exist on disk") {
					fmt.Printf("❌ Error: Required file missing from disk. File may be corrupted.\n")
				} else if strings.Contains(err.Error(), "failed to decrypt") {
					fmt.Printf("❌ Error: Failed to decrypt E2EE keys. Please verify your password.\n")
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

			fmt.Printf("\n🔒 Your file is now locked using E2EE for maximum security!\n")
			fmt.Printf("🔓 Use 'maplefile-cli files unlock' with your password to access the content when needed.\n")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to lock (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")
	cmd.MarkFlagRequired("password")

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
