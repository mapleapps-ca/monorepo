// monorepo/native/desktop/maplefile-cli/cmd/files/files.go
package files

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filedownload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/fileupload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
)

// FilesCmd creates a command for local file operations
func FilesCmd(
	logger *zap.Logger,
	addService localfile.AddService,
	uploadService fileupload.UploadService,
	listService localfile.ListService,
	localOnlyDeleteService localfile.LocalOnlyDeleteService,
	downloadService filedownload.DownloadService,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "files",
		Short: "Manage local files",
		Long:  `Import and manage files on the local filesystem without synchronization.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add file management subcommands
	cmd.AddCommand(addFileCmd(
		logger,
		addService,
	))
	cmd.AddCommand(uploadFileCmd(
		logger,
		uploadService,
	))
	cmd.AddCommand(listFilesCmd(
		logger,
		listService,
	))
	cmd.AddCommand(localOnlyDeleteFilesCmd(
		logger,
		localOnlyDeleteService,
	))
	cmd.AddCommand(downloadFileCmd(
		logger,
		downloadService,
	))
	return cmd
}
