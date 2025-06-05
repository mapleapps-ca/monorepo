// native/desktop/maplefile-cli/cmd/root.go
// Package cmd provides the CLI commands
// Location: monorepo/native/desktop/maplefile-cli/cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/cloud"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/collections"
	config_cmd "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/files"
	healthcheck "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/healthcheck"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/login"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/logout"
	cmd_md "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/me"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/recovery"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/refreshtoken"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/sync"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/verifyemail"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd/version"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsharing"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filedownload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/fileupload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	svc_me "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/me"
	svc_register "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/register"
	svc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/sync"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	uc_publiclookupdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/publiclookupdto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// NewRootCmd creates a new root command with all dependencies injected
func NewRootCmd(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository authdto.TokenDTORepository,
	userRepo user.Repository,
	regService svc_register.RegisterService,
	emailVerificationService svc_authdto.EmailVerificationService,
	loginOTTService svc_authdto.LoginOTTService,
	loginOTTVerificationService svc_authdto.LoginOTTVerificationService,
	completeLoginService svc_authdto.CompleteLoginService,
	logoutService svc_authdto.LogoutService,
	recoveryService svc_authdto.RecoveryService,
	recoveryKeyService svc_authdto.RecoveryKeyService,
	createCollectionService collection.CreateService,
	collectionListService collection.ListService,
	collectionSoftDeleteService collection.SoftDeleteService,
	collectionSharingService collectionsharing.CollectionSharingService,
	collectionGetMembersService collectionsharing.CollectionSharingGetMembersService,
	collectionListSharedService collectionsharing.ListSharedCollectionsService,
	collectionRemoveMemberService collectionsharing.CollectionSharingRemoveMembersService,
	addFileService localfile.LocalFileAddService,
	listFileService localfile.ListService,
	localOnlyDeleteService localfile.LocalOnlyDeleteService,
	uploadFileService fileupload.FileUploadService,
	downloadService filedownload.DownloadService,
	lockService localfile.LockService,
	unlockService localfile.UnlockService,
	offloadService filesyncer.OffloadService,
	onloadService filesyncer.OnloadService,
	cloudOnlyDeleteService filesyncer.CloudOnlyDeleteService,
	getFileUseCase uc_file.GetFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	getPublicLookupFromCloudUseCase uc_publiclookupdto.GetPublicLookupFromCloudUseCase,
	syncCollectionService svc_sync.SyncCollectionService,
	syncFileService svc_sync.SyncFileService,
	syncFullService svc_sync.SyncFullService,
	syncDebugService svc_sync.SyncDebugService,
	synchronizedSharingService collectionsharing.SynchronizedCollectionSharingService,
	originalSharingService collectionsharing.CollectionSharingService,
	getMeService svc_me.GetMeService,
	updateMeService svc_me.UpdateMeService,
) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "maplefile-cli",
		Short: "MapleFile CLI",
		Long: `MapleFile Command Line Interface with End-to-End Encryption

A secure, end-to-end encrypted file storage and collaboration platform.

Quick start:
  1. Register: maplefile-cli register --email you@example.com [options]
  2. Login:    maplefile-cli login --email you@example.com
  3. Create:   maplefile-cli collections create "My Files" --password PASSWORD
  4. Add:      maplefile-cli files add FILE_PATH --collection COLLECTION_ID --password PASSWORD
  5. Sync:     maplefile-cli sync --password PASSWORD

Core commands:
  login         Log in to your account
  collections   Manage collections (create, list, delete, restore, share)
  files         Manage files (add, list, get, delete)
  sync          Synchronize with cloud (unified sync + debug)
  me            View and update your profile

Advanced:
  config        Configure CLI settings
  health        Check server connectivity
  recovery      Account recovery options

For detailed help: maplefile-cli COMMAND --help`,
		Run: func(cmd *cobra.Command, args []string) {
			// Root command does nothing by default
			cmd.Help()
		},
	}

	// ========================================
	// AUTHENTICATION & USER MANAGEMENT
	// ========================================

	rootCmd.AddCommand(cmd_md.MeCmd(
		getMeService,
		updateMeService,
		logger,
	))

	rootCmd.AddCommand(register.RegisterCmd(regService))
	rootCmd.AddCommand(verifyemail.VerifyEmailCmd(emailVerificationService, logger))

	rootCmd.AddCommand(login.LoginCmd(
		loginOTTService,
		loginOTTVerificationService,
		completeLoginService,
		logger,
	))
	rootCmd.AddCommand(login.RequestLoginTokenCmd(loginOTTService, logger))
	rootCmd.AddCommand(login.VerifyLoginTokenCmd(loginOTTVerificationService, logger))
	rootCmd.AddCommand(login.CompleteLoginCmd(completeLoginService, logger))

	rootCmd.AddCommand(logout.LogoutCmd(logoutService, logger))
	rootCmd.AddCommand(refreshtoken.RefreshTokenCmd(logger, configService, tokenRepository))

	rootCmd.AddCommand(recovery.RecoveryCmd(recoveryService, logger))
	rootCmd.AddCommand(recovery.ShowRecoveryKeyCmd(recoveryKeyService, logger))

	// ========================================
	// COLLECTIONS
	// ========================================

	rootCmd.AddCommand(collections.CollectionsCmd(
		createCollectionService,
		collectionListService,
		collectionSoftDeleteService,
		collectionSharingService,
		collectionGetMembersService,
		collectionListSharedService,
		collectionRemoveMemberService,
		synchronizedSharingService,
		originalSharingService,
		logger,
	))

	// ========================================
	// FILES
	// ========================================

	rootCmd.AddCommand(files.FilesCmd(
		logger,
		addFileService,
		uploadFileService,
		listFileService,
		localOnlyDeleteService,
		downloadService,
		offloadService,
		onloadService,
		cloudOnlyDeleteService,
		lockService,
		unlockService,
		getFileUseCase,
		getUserByIsLoggedInUseCase,
		getCollectionUseCase,
	))

	// ========================================
	// SYNC
	// ========================================

	// Clean sync command with unified sync + debug
	rootCmd.AddCommand(sync.SyncCmd(
		syncCollectionService,
		syncFileService,
		syncDebugService,
		logger,
	))

	// ========================================
	// CONFIGURATION & UTILITIES
	// ========================================
	rootCmd.AddCommand(healthcheck.HealthCheckCmd(configService))
	rootCmd.AddCommand(config_cmd.ConfigCmd(configService))
	rootCmd.AddCommand(version.VersionCmd())
	rootCmd.AddCommand(cloud.CloudCmd(
		configService,
		getPublicLookupFromCloudUseCase,
		logger))

	return rootCmd
}
