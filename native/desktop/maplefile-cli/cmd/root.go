// Package cmd provides the CLI commands
// Location: monorepo/native/desktop/maplefile-cli/cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/collections"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/completelogin"
	config_cmd "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/filesyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/refreshtoken"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/remote"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/requestloginott"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/verifyemail"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/verifyloginott"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/version"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsyncer"
	srv_filesyncer "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localcollection"
	localfileService "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	registerService "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/remotecollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/remotefile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/tokenservice"
)

// NewRootCmd creates a new root command with all dependencies injected
func NewRootCmd(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	regService registerService.RegisterService,
	emailVerificationService auth.EmailVerificationService,
	loginOTTService auth.LoginOTTService,
	loginOTTVerificationService auth.LoginOTTVerificationService,
	completeLoginService auth.CompleteLoginService,
	tokenRefreshSvc tokenservice.TokenRefreshService,
	remoteCollectionService remotecollection.CreateService,
	remoteListService remotecollection.ListService,
	downloadService collectionsyncer.DownloadService,
	collectionListService localcollection.ListService,
	fileImportService localfileService.ImportService,
	fileListService localfileService.ListService,
	fileDeleteService localfileService.DeleteService,
	fileGetService localfileService.GetService,
	remoteFetchService remotefile.FetchService,
	// File syncer services with correct interfaces
	uploadToRemoteService srv_filesyncer.UploadToRemoteService,
	downloadToLocalService srv_filesyncer.DownloadToLocalService,
	syncFileService srv_filesyncer.SyncFileService,
	syncCollectionService srv_filesyncer.SyncCollectionService,
	// other services...
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
	rootCmd.AddCommand(remote.RemoteCmd(configService, remoteListService, logger))
	rootCmd.AddCommand(register.RegisterCmd(regService))
	rootCmd.AddCommand(verifyemail.VerifyEmailCmd(emailVerificationService, logger))
	rootCmd.AddCommand(requestloginott.RequestLoginOneTimeTokenUserCmd(loginOTTService, logger))
	rootCmd.AddCommand(verifyloginott.VerifyLoginOneTimeTokenUserCmd(loginOTTVerificationService, logger))
	rootCmd.AddCommand(completelogin.CompleteLoginCmd(completeLoginService, logger))
	rootCmd.AddCommand(refreshtoken.RefreshTokenCmd(logger, configService, userRepo, tokenRefreshSvc))
	rootCmd.AddCommand(collections.CollectionsCmd(
		remoteCollectionService,
		downloadService,
		collectionListService,
		logger))
	rootCmd.AddCommand(localfile.LocalFileCmd(
		fileImportService,
		fileDeleteService,
		fileGetService,
		fileListService,
		remoteFetchService,
		logger))
	rootCmd.AddCommand(filesyncer.FileSyncerCmd(
		fileImportService,
		fileDeleteService,
		fileGetService,
		uploadToRemoteService,
		downloadToLocalService,
		syncFileService,
		syncCollectionService,
		remoteFetchService,
		logger))

	return rootCmd
}
