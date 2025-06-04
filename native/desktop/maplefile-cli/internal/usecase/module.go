// internal/usecase/module.go
package usecase

import (
	"go.uber.org/fx"

	uc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectionsharingdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/filedto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/fileupload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/publiclookupdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/refreshtoken"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/syncstate"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// UseCaseModule provides the service-layer--related dependencies
func UseCaseModule() fx.Option {
	return fx.Options(
		// Auth use cases
		fx.Provide(uc_authdto.NewEmailVerificationUseCase),
		fx.Provide(uc_authdto.NewLoginOTTUseCase),
		fx.Provide(uc_authdto.NewLoginOTTVerificationUseCase),
		fx.Provide(uc_authdto.NewCompleteLoginUseCase),
		fx.Provide(uc_authdto.NewLogoutUseCase),
		fx.Provide(uc_authdto.NewRecoveryUseCase),

		// User repository use cases
		fx.Provide(user.NewGetByEmailUseCase),
		fx.Provide(user.NewUpsertByEmailUseCase),
		fx.Provide(user.NewDeleteByEmailUseCase),
		fx.Provide(user.NewListAllUseCase),
		fx.Provide(user.NewGetByIsLoggedInUseCase),

		// Cloud-based collection use cases
		fx.Provide(collectiondto.NewCreateCollectionInCloudUseCase),
		fx.Provide(collectiondto.NewGetFilteredCollectionsFromCloudUseCase),
		fx.Provide(collectiondto.NewGetCollectionFromCloudUseCase),
		// Local-based collection use cases
		fx.Provide(collection.NewCreateCollectionUseCase),
		fx.Provide(collection.NewGetCollectionUseCase),
		fx.Provide(collection.NewListCollectionsUseCase),
		fx.Provide(collection.NewUpdateCollectionUseCase),
		fx.Provide(collection.NewDeleteCollectionUseCase),
		fx.Provide(collection.NewMoveCollectionUseCase),
		fx.Provide(collection.NewGetCollectionPathUseCase),
		fx.Provide(collection.NewSoftDeleteService),

		// Cloud-based collection sharing use cases
		fx.Provide(collectionsharingdto.NewShareCollectionUseCase),
		fx.Provide(collectionsharingdto.NewRemoveMemberUseCase),
		fx.Provide(collectionsharingdto.NewListSharedCollectionsUseCase),

		// File database use cases (for managing file records)
		fx.Provide(file.NewCreateFileUseCase),
		fx.Provide(file.NewCreateFilesUseCase),
		fx.Provide(file.NewGetFileUseCase),
		fx.Provide(file.NewGetFilesByIDsUseCase),
		fx.Provide(file.NewListFilesByCollectionUseCase),
		fx.Provide(file.NewUpdateFileUseCase),
		fx.Provide(file.NewDeleteFileUseCase),
		fx.Provide(file.NewDeleteFilesUseCase),
		fx.Provide(file.NewCheckFileExistsUseCase),
		fx.Provide(file.NewCheckFileAccessUseCase),
		fx.Provide(file.NewSwapIDsUseCase),

		// Local file system use cases (actual file operations)
		fx.Provide(localfile.NewReadFileUseCase),
		fx.Provide(localfile.NewWriteFileUseCase),
		fx.Provide(localfile.NewDeleteFileUseCase),
		fx.Provide(localfile.NewCopyFileUseCase),
		fx.Provide(localfile.NewMoveFileUseCase),
		fx.Provide(localfile.NewCheckFileExistsUseCase),
		fx.Provide(localfile.NewGetFileInfoUseCase),
		fx.Provide(localfile.NewCreateDirectoryUseCase),
		fx.Provide(localfile.NewListDirectoryUseCase),
		fx.Provide(localfile.NewComputeFileHashUseCase),
		fx.Provide(localfile.NewPathUtilsUseCase),

		// File upload use cases
		fx.Provide(fileupload.NewEncryptFileUseCase),
		fx.Provide(fileupload.NewPrepareFileUploadUseCase),

		// File DTO use cases
		fx.Provide(filedto.NewGetPresignedDownloadURLUseCase),
		fx.Provide(filedto.NewDownloadFileUseCase),

		// Registration use cases
		fx.Provide(register.NewGenerateCredentialsUseCase),
		fx.Provide(register.NewCreateLocalUserUseCase),
		fx.Provide(register.NewSendRegistrationToServerUseCase),

		// Token refresh usecase
		fx.Provide(refreshtoken.NewRefreshTokenUseCase),

		// Sync state use cases
		fx.Provide(syncstate.NewGetSyncStateUseCase),
		fx.Provide(syncstate.NewSaveSyncStateUseCase),
		fx.Provide(syncstate.NewResetSyncStateUseCase),
		fx.Provide(syncstate.NewUpdateCollectionSyncUseCase),
		fx.Provide(syncstate.NewUpdateFileSyncUseCase),

		// Sync DTO use cases
		fx.Provide(syncdto.NewGetCollectionSyncDataUseCase),
		fx.Provide(syncdto.NewGetFileSyncDataUseCase),
		fx.Provide(syncdto.NewBuildSyncCursorUseCase),
		fx.Provide(syncdto.NewProcessSyncResponseUseCase),

		// Get public lookup DTO
		fx.Provide(publiclookupdto.NewGetPublicLookupFromCloudUseCase),
	)
}
