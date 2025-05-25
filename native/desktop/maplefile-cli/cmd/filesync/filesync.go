// native/desktop/maplefile-cli/cmd/filesync/filesync.go
package filesync

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
)

// FileSyncCmd creates a command for file synchronization operations
func FileSyncCmd(
	offloadService filesyncer.OffloadService,
	onloadService filesyncer.OnloadService,
	cloudDeleteService filesyncer.CloudDeleteService,
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
	cmd.AddCommand(cloudDeleteCmd(cloudDeleteService, logger))

	return cmd
}
