// native/desktop/maplefile-cli/cmd/files/unlock.go
package files

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
)

// unlockFileCmd creates a command for unlocking a file
func unlockFileCmd(
	logger *zap.Logger,
	unlockService localfile.UnlockService,
) *cobra.Command {
	var fileID string
	var storageMode string

	var cmd = &cobra.Command{
		Use:   "unlock",
		Short: "Unlock a file to access decrypted content",
		Long: `
Unlock a file to access its decrypted content. You can choose between
two storage modes:

* "decrypted_only": Keep only the decrypted version (removes encrypted version)
* "hybrid": Keep both encrypted and decrypted versions (recommended)

The file must have a local decrypted version available. Cloud-only files
cannot be unlocked directly - use 'filesync onload' first to download them.

Examples:
  # Unlock to decrypted-only mode (removes encrypted version)
  maplefile-cli files unlock --file-id 507f1f77bcf86cd799439011 --mode decrypted_only

  # Unlock to hybrid mode (keeps both versions) - RECOMMENDED
  maplefile-cli files unlock --file-id 507f1f77bcf86cd799439011 --mode hybrid

  # Unlock to hybrid mode (default if no mode specified)
  maplefile-cli files unlock --file-id 507f1f77bcf86cd799439011
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required inputs
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to unlock.")
				return
			}

			// Set default storage mode if not provided
			if storageMode == "" {
				storageMode = dom_file.StorageModeHybrid // Safer default
			}

			// Validate storage mode
			if storageMode != dom_file.StorageModeDecryptedOnly && storageMode != dom_file.StorageModeHybrid {
				fmt.Printf("❌ Error: Invalid storage mode '%s'. Must be 'decrypted_only' or 'hybrid'.\n", storageMode)
				return
			}

			// Create service input
			input := &localfile.UnlockInput{
				FileID:      fileID,
				StorageMode: storageMode,
			}

			// Execute unlock operation
			fmt.Printf("🔓 Unlocking file: %s\n", fileID)
			if storageMode == dom_file.StorageModeHybrid {
				fmt.Println("🔄 Switching to hybrid mode (keeping both encrypted and decrypted versions)...")
			} else {
				fmt.Println("🔄 Switching to decrypted-only mode (removing encrypted version)...")
			}

			output, err := unlockService.Unlock(cmd.Context(), input)
			if err != nil {
				if strings.Contains(err.Error(), "cloud-only") {
					fmt.Printf("❌ Error: Cannot unlock cloud-only files. Use 'filesync onload' first to download the file locally.\n")
				} else if strings.Contains(err.Error(), "file not found") {
					fmt.Printf("❌ Error: File not found. Please check the file ID.\n")
				} else if strings.Contains(err.Error(), "no decrypted file") {
					fmt.Printf("❌ Error: No decrypted version available. Cannot unlock file without decrypted version.\n")
				} else if strings.Contains(err.Error(), "does not exist on disk") {
					fmt.Printf("❌ Error: Decrypted file missing from disk. File may be corrupted.\n")
				} else if strings.Contains(err.Error(), "storage mode must be") {
					fmt.Printf("❌ Error: Invalid storage mode. Use 'decrypted_only' or 'hybrid'.\n")
				} else {
					fmt.Printf("❌ Error unlocking file: %v\n", err)
				}
				return
			}

			// Display success information
			fmt.Printf("\n✅ File successfully unlocked!\n")
			fmt.Printf("🆔 File ID: %s\n", output.FileID.Hex())
			fmt.Printf("📊 Storage Mode: %s → %s\n", output.PreviousMode, output.NewMode)

			if output.DeletedPath != "" {
				fmt.Printf("🗑️  Deleted: %s\n", getPathDisplayName(output.DeletedPath))
			}
			fmt.Printf("💾 Available: %s\n", getPathDisplayName(output.RemainingPath))
			fmt.Printf("💬 Status: %s\n", output.Message)

			if storageMode == dom_file.StorageModeHybrid {
				fmt.Printf("\n🔓 Your file is now unlocked in hybrid mode!\n")
				fmt.Printf("📁 Both encrypted and decrypted versions are available for flexibility.\n")
			} else {
				fmt.Printf("\n🔓 Your file is now unlocked in decrypted-only mode!\n")
				fmt.Printf("⚠️  The encrypted version has been removed. Use with caution.\n")
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to unlock (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVarP(&storageMode, "mode", "m", "hybrid", "Storage mode: 'decrypted_only' or 'hybrid' (default: hybrid)")

	return cmd
}
