// cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"

	config_cmd "github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/remote"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/version"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/config"
)

// NewRootCmd creates a new root command with all dependencies injected
func NewRootCmd(configUseCase config.ConfigUseCase) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "papercloud-cli",
		Short: "PaperCloud CLI",
		Long:  `PaperCloud Command Line Interface`,
		Run: func(cmd *cobra.Command, args []string) {
			// Root command does nothing by default
			cmd.Help()
		},
	}

	// Attach sub-commands to our main root
	rootCmd.AddCommand(version.VersionCmd())
	rootCmd.AddCommand(config_cmd.ConfigCmd(configUseCase))
	rootCmd.AddCommand(remote.RemoteCmd(configUseCase))

	return rootCmd
}
