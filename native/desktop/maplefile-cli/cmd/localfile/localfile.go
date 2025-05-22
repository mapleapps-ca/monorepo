// monorepo/native/desktop/maplefile-cli/cmd/localfile/localfile.go
package localfile

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// LocalFileCmd creates a command for local file operations
func LocalFileCmd(
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "localfile",
		Short: "Manage local files",
		Long:  `Import and manage files on the local filesystem without synchronization.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add file management subcommands
	// cmd.AddCommand(createLocalFileCmd(importService, logger))
	// cmd.AddCommand(listLocalFileCmd(listService, logger))

	return cmd
}
