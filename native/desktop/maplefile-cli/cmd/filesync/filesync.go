// native/desktop/maplefile-cli/cmd/filesync/filesync.go
package filesync

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
)

// FileSyncCmd creates a command for file synchronization operations
func FileSyncCmd(
	offloadService filesyncer.OffloadService,
	onloadService filesyncer.OnloadService,
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "filesync",
		Short: "Synchronize files between local and cloud storage",
		Long:  `Offload files to cloud-only storage or onload cloud files to local storage.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add file sync subcommands
	cmd.AddCommand(offloadCmd(offloadService, logger))
	cmd.AddCommand(onloadCmd(onloadService, logger))

	return cmd
}

// offloadCmd creates a command for offloading files to cloud storage
func offloadCmd(
	offloadService filesyncer.OffloadService,
	logger *zap.Logger,
) *cobra.Command {
	var fileID string
	var password string

	var cmd = &cobra.Command{
		Use:   "offload",
		Short: "Offload a file to cloud-only storage",
		Long: `
Offload a local file to cloud-only storage. This will:
- Upload the file if it hasn't been uploaded yet
- Remove the local decrypted copy if already uploaded
- Set the file's sync status to cloud-only

Examples:
  maplefile-cli filesync offload --file-id 507f1f77bcf86cd799439011 --password 1234567890
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required fields
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to offload.")
				return
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Create service input
			input := &filesyncer.OffloadInput{
				FileID:       fileID,
				UserPassword: password,
			}

			// Execute offload
			fmt.Printf("🔄 Offloading file: %s\n", fileID)

			output, err := offloadService.Offload(cmd.Context(), input)
			if err != nil {
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else {
					fmt.Printf("❌ Error offloading file: %v\n", err)
				}
				return
			}

			// Display success information
			fmt.Printf("\n✅ File successfully offloaded!\n")
			fmt.Printf("🆔 File ID: %s\n", output.FileID.Hex())
			fmt.Printf("🔄 Action: %s\n", output.Action)
			fmt.Printf("📊 Status: %v → %v\n", output.PreviousStatus, output.NewStatus)
			fmt.Printf("💬 Message: %s\n", output.Message)

			if output.UploadResult != nil && output.UploadResult.Success {
				fmt.Printf("☁️  Cloud File ID: %s\n", output.UploadResult.CloudFileID.Hex())
				fmt.Printf("📏 Uploaded Size: %d bytes\n", output.UploadResult.FileSizeBytes)
			}

			fmt.Printf("\n🎉 Your file is now stored in cloud-only mode!\n")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to offload (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")
	cmd.MarkFlagRequired("password")

	return cmd
}

// onloadCmd creates a command for onloading files from cloud storage
func onloadCmd(
	onloadService filesyncer.OnloadService,
	logger *zap.Logger,
) *cobra.Command {
	var fileID string
	var password string

	var cmd = &cobra.Command{
		Use:   "onload",
		Short: "Onload a cloud-only file to local storage",
		Long: `
Onload a cloud-only file to local storage. This will:
- Download the encrypted file from cloud storage
- Decrypt the file using your encryption keys
- Save the decrypted file locally
- Set the file's sync status to synced

Examples:
  maplefile-cli filesync onload --file-id 507f1f77bcf86cd799439011 --password 1234567890
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required fields
			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				fmt.Println("Use --file-id flag to specify the file to onload.")
				return
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Create service input
			input := &filesyncer.OnloadInput{
				FileID:       fileID,
				UserPassword: password,
			}

			// Execute onload
			fmt.Printf("🔄 Onloading file: %s\n", fileID)
			fmt.Println("📡 Downloading and decrypting file from cloud...")

			output, err := onloadService.Onload(cmd.Context(), input)
			if err != nil {
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else if strings.Contains(err.Error(), "not cloud-only") {
					fmt.Printf("❌ Error: File is not in cloud-only mode. Only cloud-only files can be onloaded.\n")
				} else {
					fmt.Printf("❌ Error onloading file: %v\n", err)
				}
				return
			}

			// Display success information
			fmt.Printf("\n✅ File successfully onloaded!\n")
			fmt.Printf("🆔 File ID: %s\n", output.FileID.Hex())
			fmt.Printf("📊 Status: %v → %v\n", output.PreviousStatus, output.NewStatus)
			fmt.Printf("💾 Local Path: %s\n", output.DecryptedPath)
			fmt.Printf("📏 Downloaded Size: %d bytes\n", output.DownloadedSize)
			fmt.Printf("💬 Message: %s\n", output.Message)

			fmt.Printf("\n🎉 Your file is now available locally!\n")
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to onload (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")
	cmd.MarkFlagRequired("password")

	return cmd
}
