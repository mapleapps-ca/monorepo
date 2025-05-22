// native/desktop/maplefile-cli/cmd/filesyncer/filesyncer.go
package filesyncer

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/remotefile"
)

// FileSyncerCmd creates a command to handles synchronization between local and remote
func FileSyncerCmd(
	importService localfile.ImportService,
	deleteService localfile.DeleteService,
	getService localfile.GetService,
	uploadToRemoteService filesyncer.UploadToRemoteService,
	downloadToLocalService filesyncer.DownloadToLocalService,
	syncFileService filesyncer.SyncFileService,
	syncCollectionService filesyncer.SyncCollectionService,
	remoteFetchService remotefile.FetchService,
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "filesyncer",
		Short: "Manage remote/local file synchronization",
		Long: `
File Syncer manages synchronization between local and remote files.

This command provides various operations for keeping your local and remote files
in sync, including uploading, downloading, and bidirectional synchronization.

Available operations:
* upload: Upload local files to remote backend
* download: Download remote files to local storage
* sync: Intelligently sync individual files
* sync-collection: Bulk sync all files in a collection

Examples:
  # Upload a specific local file
  maplefile-cli filesyncer upload --file-id 507f1f77bcf86cd799439011

  # Download a specific remote file
  maplefile-cli filesyncer download --remote-id 507f1f77bcf86cd799439011

  # Auto-sync a file by encrypted ID
  maplefile-cli filesyncer sync --encrypted-id abc123def456

  # Sync all files in a collection
  maplefile-cli filesyncer sync-collection --collection-id 507f1f77bcf86cd799439011
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add file synchronization subcommands
	cmd.AddCommand(uploadLocalFileCmd(uploadToRemoteService, getService, logger))
	cmd.AddCommand(downloadRemoteFileCmd(downloadToLocalService, logger))
	cmd.AddCommand(syncFileCmd(syncFileService, logger))
	cmd.AddCommand(syncCollectionCmd(syncCollectionService, logger))

	return cmd
}

// confirmAction asks for user confirmation and returns true if the user confirms
func confirmAction(message string) bool {
	var response string
	fmt.Print(message)
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}
