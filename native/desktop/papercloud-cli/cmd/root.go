// Package cmd provides the CLI commands
// Location: monorepo/native/desktop/papercloud-cli/cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/completelogin"
	config_cmd "github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/remote"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/requestloginott"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/uploadfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/verifyemail"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/verifyloginott"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd/version"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
)

// NewRootCmd creates a new root command with all dependencies injected
func NewRootCmd(configService config.ConfigService, userRepo user.Repository) *cobra.Command {
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
	rootCmd.AddCommand(config_cmd.ConfigCmd(configService))
	rootCmd.AddCommand(remote.RemoteCmd(configService))
	rootCmd.AddCommand(register.RegisterCmd(configService, userRepo))
	rootCmd.AddCommand(verifyemail.VerifyEmailCmd(configService, userRepo))
	rootCmd.AddCommand(requestloginott.RequestLoginOneTimeTokenUserCmd(configService))
	rootCmd.AddCommand(verifyloginott.VerifyLoginOneTimeTokenUserCmd())
	rootCmd.AddCommand(completelogin.CompleteLoginCmd())
	rootCmd.AddCommand(uploadfile.UploadFileCmd())

	return rootCmd
}
