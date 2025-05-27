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
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/filesync"
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
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filedownload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/fileupload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	svc_register "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/register"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
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
	collectionSoftDeleteService collection.SoftDeleteService,
	addFileService localfile.AddService,
	listFileService localfile.ListService,
	localOnlyDeleteService localfile.LocalOnlyDeleteService,
	uploadFileService fileupload.UploadService,
	downloadService filedownload.DownloadService,
	lockService localfile.LockService,
	unlockService localfile.UnlockService,
	offloadService filesyncer.OffloadService,
	onloadService filesyncer.OnloadService,
	cloudOnlyDeleteService filesyncer.CloudOnlyDeleteService,
	getFileUseCase uc_file.GetFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "maplefile-cli",
		Short: "MapleFile CLI",
		Long:  `MapleFile Command Line Interface with End-to-End Encryption`,
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
		collectionSoftDeleteService,
		logger,
	))
	rootCmd.AddCommand(files.FilesCmd(
		logger,
		addFileService,
		uploadFileService,
		listFileService,
		localOnlyDeleteService,
		downloadService,
		lockService,
		unlockService,
		getFileUseCase,
		getUserByIsLoggedInUseCase,
		getCollectionUseCase,
	))
	// Add the filesync command
	rootCmd.AddCommand(filesync.FileSyncCmd(
		offloadService,
		onloadService,
		cloudOnlyDeleteService,
		logger,
	))
	return rootCmd
}
