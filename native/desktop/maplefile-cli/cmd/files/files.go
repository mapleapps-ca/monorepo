// monorepo/native/desktop/maplefile-cli/cmd/files/files.go
package files

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filedownload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/fileupload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// FilesCmd creates a command for local file operations
func FilesCmd(
	logger *zap.Logger,
	addService localfile.LocalFileAddService,
	fileUploadService fileupload.FileUploadService,
	listService localfile.ListService,
	localOnlyDeleteService localfile.LocalOnlyDeleteService,
	downloadService filedownload.DownloadService,
	lockService localfile.LockService,
	unlockService localfile.UnlockService,
	getFileUseCase uc_file.GetFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "files",
		Short: "Manage local files",
		Long:  `Import, manage, and download files with end-to-end encryption.`,
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
		fileUploadService,
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
	cmd.AddCommand(lockFileCmd(
		logger,
		lockService,
	))
	cmd.AddCommand(unlockFileCmd(
		logger,
		unlockService,
	))
	cmd.AddCommand(debugE2EECmd(
		logger,
		getFileUseCase,
		getUserByIsLoggedInUseCase,
		getCollectionUseCase,
	))

	return cmd
}
