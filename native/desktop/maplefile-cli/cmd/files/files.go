// cmd/files/files.go - Clean main files command
package files

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/files/filesync"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/files/misc"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filedownload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/fileupload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// FilesCmd creates the main files command with clean, simplified subcommands
func FilesCmd(
	logger *zap.Logger,
	addService localfile.LocalFileAddService,
	fileUploadService fileupload.FileUploadService,
	listService localfile.ListService,
	localOnlyDeleteService localfile.LocalOnlyDeleteService,
	downloadService filedownload.DownloadService,
	offloadService filesyncer.OffloadService,
	onloadService filesyncer.OnloadService,
	cloudOnlyDeleteService filesyncer.CloudOnlyDeleteService,
	lockService localfile.LockService,
	unlockService localfile.UnlockService,
	getFileUseCase uc_file.GetFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "files",
		Short: "Manage your files",
		Long: `
Manage files in your MapleFile collections with end-to-end encryption.

All file operations use E2EE to ensure your data remains private and secure.
Files can be stored locally, in the cloud, or both depending on your needs.

Available commands:
  add      Add files to collections (auto-uploads by default)
  list     List files in collections
  get      Download and decrypt files
  delete   Delete files (local, cloud, or both)

Examples:
  # Add a file to a collection (uploads automatically)
  maplefile-cli files add "/path/to/document.pdf" --collection COLLECTION_ID --password PASSWORD

  # List files in a collection
  maplefile-cli files list --collection COLLECTION_ID

  # Download a file
  maplefile-cli files get FILE_ID --password PASSWORD

  # Delete a file completely
  maplefile-cli files delete FILE_ID --password PASSWORD

  # Add file locally only (upload later)
  maplefile-cli files add "/path/to/file.txt" --collection COLLECTION_ID --local-only --password PASSWORD

For detailed help: maplefile-cli files COMMAND --help
`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Core file management commands (clean and simple)
	cmd.AddCommand(addFileCmd(logger, addService, fileUploadService))
	cmd.AddCommand(listFilesCmd(logger, listService))
	cmd.AddCommand(getFileCmd(logger, downloadService, onloadService))
	cmd.AddCommand(deleteFileCmd(logger, localOnlyDeleteService, cloudOnlyDeleteService))
	cmd.AddCommand(filesync.FileSyncCmd(offloadService, onloadService, cloudOnlyDeleteService, logger))
	cmd.AddCommand(misc.MiscFilesCmd(
		logger,
		localOnlyDeleteService,
		downloadService,
		lockService,
		unlockService,
		getFileUseCase,
		getUserByIsLoggedInUseCase,
		getCollectionUseCase,
	))

	return cmd
}
