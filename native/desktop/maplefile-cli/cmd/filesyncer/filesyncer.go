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
	fileSyncerUploadService filesyncer.UploadService,
	remoteFetchService remotefile.FetchService,
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "filesyncer",
		Short: "Manage remote/local files",
		Long:  `Handles synchronization between local and remote.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add file management subcommands
	cmd.AddCommand(uploadLocalFileCmd(fileSyncerUploadService, getService, logger))

	return cmd
}

// confirmAction asks for user confirmation and returns true if the user confirms
func confirmAction(message string) bool {
	var response string
	fmt.Print(message)
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}
