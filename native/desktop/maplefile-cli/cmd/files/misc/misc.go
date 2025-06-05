// monorepo/native/desktop/maplefile-cli/cmd/files/misc/files.go
package misc

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filedownload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// MiscFilesCmd creates a command for local file operations
func MiscFilesCmd(
	logger *zap.Logger,
	localOnlyDeleteService localfile.LocalOnlyDeleteService,
	downloadService filedownload.DownloadService,
	lockService localfile.LockService,
	unlockService localfile.UnlockService,
	getFileUseCase uc_file.GetFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "misc",
		Short: "Misc file operations",
		Long:  `Extra misc file operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add file management subcommands
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
