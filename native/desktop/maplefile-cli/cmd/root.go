// Package cmd provides the CLI commands
// Location: monorepo/native/desktop/maplefile-cli/cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/cloud"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/collections"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/completelogin"
	config_cmd "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/files"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/refreshtoken"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/requestloginott"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/verifyemail"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/verifyloginott"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/version"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_auth "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
	svc_register "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// NewRootCmd creates a new root command with all dependencies injected
func NewRootCmd(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository auth.TokenRepository,
	userRepo user.Repository,
	regService svc_register.RegisterService,
	emailVerificationService svc_auth.EmailVerificationService,
	loginOTTService svc_auth.LoginOTTService,
	loginOTTVerificationService svc_auth.LoginOTTVerificationService,
	completeLoginService svc_auth.CompleteLoginService,
	createCollectionService collection.CreateService,
	collectionListService collection.ListService,
	// Local file use cases
	readFileUseCase localfile.ReadFileUseCase,
	checkFileExistsUseCase localfile.CheckFileExistsUseCase,
	getFileInfoUseCase localfile.GetFileInfoUseCase,
	pathUtilsUseCase localfile.PathUtilsUseCase,
) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "maplefile-cli",
		Short: "MapleFile CLI",
		Long:  `MapleFile Command Line Interface`,
		Run: func(cmd *cobra.Command, args []string) {
			// Root command does nothing by default
			cmd.Help()
		},
	}

	// Attach sub-commands to our main root
	rootCmd.AddCommand(version.VersionCmd())
	rootCmd.AddCommand(config_cmd.ConfigCmd(configService))
	rootCmd.AddCommand(cloud.RemoteCmd(configService, logger))
	rootCmd.AddCommand(register.RegisterCmd(regService))
	rootCmd.AddCommand(verifyemail.VerifyEmailCmd(emailVerificationService, logger))
	rootCmd.AddCommand(requestloginott.RequestLoginOneTimeTokenUserCmd(loginOTTService, logger))
	rootCmd.AddCommand(verifyloginott.VerifyLoginOneTimeTokenUserCmd(loginOTTVerificationService, logger))
	rootCmd.AddCommand(completelogin.CompleteLoginCmd(completeLoginService, logger))
	rootCmd.AddCommand(refreshtoken.RefreshTokenCmd(logger, configService, tokenRepository))
	rootCmd.AddCommand(collections.CollectionsCmd(
		createCollectionService,
		collectionListService,
		logger,
	))
	rootCmd.AddCommand(files.FilesCmd(
		logger,
		readFileUseCase,
		checkFileExistsUseCase,
		getFileInfoUseCase,
		pathUtilsUseCase,
	))
	return rootCmd
}
