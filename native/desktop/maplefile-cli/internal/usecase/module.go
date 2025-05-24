// internal/usecase/module.go
package usecase

import (
	"go.uber.org/fx"

	authUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/crypto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/refreshtoken"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// UseCaseModule provides the service-layer--related dependencies
func UseCaseModule() fx.Option {
	return fx.Options(
		// Crypto use cases
		fx.Provide(crypto.NewCryptoUseCase),
		fx.Provide(crypto.NewDecryptCollectionNameUseCase),

		// Auth use cases
		fx.Provide(authUseCase.NewEmailVerificationUseCase),
		fx.Provide(authUseCase.NewLoginOTTUseCase),
		fx.Provide(authUseCase.NewLoginOTTVerificationUseCase),
		fx.Provide(authUseCase.NewCompleteLoginUseCase),

		// User repository use cases
		fx.Provide(user.NewGetByEmailUseCase),
		fx.Provide(user.NewUpsertByEmailUseCase),
		fx.Provide(user.NewDeleteByEmailUseCase),
		fx.Provide(user.NewListAllUseCase),
		fx.Provide(user.NewGetByIsLoggedInUseCase),

		// Cloud-based collection use cases
		fx.Provide(collectiondto.NewCreateCollectionInCloudUseCase),

		// Local-based collection use cases
		fx.Provide(collection.NewCreateCollectionUseCase),
		fx.Provide(collection.NewGetCollectionUseCase),
		fx.Provide(collection.NewListCollectionsUseCase),
		fx.Provide(collection.NewUpdateCollectionUseCase),
		fx.Provide(collection.NewDeleteCollectionUseCase),
		fx.Provide(collection.NewMoveCollectionUseCase),
		fx.Provide(collection.NewGetCollectionPathUseCase),

		// Core file use cases (repository level)
		fx.Provide(file.NewCreateFileUseCase),
		fx.Provide(file.NewCreateFilesUseCase),
		fx.Provide(file.NewGetFileUseCase),
		fx.Provide(file.NewGetFilesUseCase),
		fx.Provide(file.NewListFilesUseCase),
		fx.Provide(file.NewUpdateFileUseCase),
		fx.Provide(file.NewDeleteFileUseCase),
		fx.Provide(file.NewDeleteFilesUseCase),
		fx.Provide(file.NewCheckFileExistsUseCase),
		fx.Provide(file.NewCheckFileAccessUseCase),

		// Local file use cases (business logic level)
		// Read operations
		fx.Provide(localfile.NewGetFileUseCase),
		fx.Provide(localfile.NewCheckFileExistsUseCase),
		fx.Provide(localfile.NewCheckUserAccessUseCase),
		fx.Provide(localfile.NewGetFilesByIDsUseCase),
		fx.Provide(localfile.NewGetFilesByCollectionUseCase),
		fx.Provide(localfile.NewValidateFilePathsUseCase),

		// Write operations
		fx.Provide(localfile.NewCreateFileUseCase),
		fx.Provide(localfile.NewUpdateFileUseCase),
		fx.Provide(localfile.NewDeleteFileUseCase),
		fx.Provide(localfile.NewCreateFilesUseCase),
		fx.Provide(localfile.NewDeleteFilesUseCase),
		fx.Provide(localfile.NewDeleteFilesByCollectionUseCase),
		fx.Provide(localfile.NewChangeStorageModeUseCase),
		fx.Provide(localfile.NewCleanupOrphanedFilesUseCase),

		// // Cloud collection use cases
		// fx.Provide(remotecollection.NewCreateRemoteCollectionUseCase),
		// fx.Provide(remotecollection.NewFetchRemoteCollectionUseCase),
		// fx.Provide(remotecollection.NewListRemoteCollectionsUseCase),

		// Registration use cases
		fx.Provide(register.NewGenerateCredentialsUseCase),
		fx.Provide(register.NewCreateLocalUserUseCase),
		fx.Provide(register.NewSendRegistrationToServerUseCase),

		// Token refresh usecase
		fx.Provide(refreshtoken.NewRefreshTokenUseCase),
	)
}
