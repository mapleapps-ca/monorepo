// monorepo/native/desktop/maplefile-cli/cmd/file/file.go
package files

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// FileCmd creates a command for local file operations
func FilesCmd(
	logger *zap.Logger,
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
	// cmd.AddCommand(createFileCmd(importService, logger))
	// cmd.AddCommand(listFileCmd(listService, logger))

	return cmd
}
