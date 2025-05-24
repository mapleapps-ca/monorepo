// monorepo/native/desktop/maplefile-cli/cmd/files/files.go
package files

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// FilesCmd creates a command for local file operations
func FilesCmd(
	logger *zap.Logger,
	readFileUseCase localfile.ReadFileUseCase,
	checkFileExistsUseCase localfile.CheckFileExistsUseCase,
	getFileInfoUseCase localfile.GetFileInfoUseCase,
	pathUtilsUseCase localfile.PathUtilsUseCase,
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
		readFileUseCase,
		checkFileExistsUseCase,
		getFileInfoUseCase,
		pathUtilsUseCase,
	))

	return cmd
}
