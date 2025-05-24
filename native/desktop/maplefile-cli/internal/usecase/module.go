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
	registerUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/register"
	userUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
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
		fx.Provide(userUseCase.NewGetByEmailUseCase),
		fx.Provide(userUseCase.NewUpsertByEmailUseCase),
		fx.Provide(userUseCase.NewDeleteByEmailUseCase),
		fx.Provide(userUseCase.NewListAllUseCase),
		fx.Provide(userUseCase.NewGetByIsLoggedInUseCase),

		// Cloud-based collection use cases
		fx.Provide(collectiondto.NewCreateCollectionInCloudUseCase),
		// fx.Provide(remotecollection.NewCreateRemoteCollectionUseCase),
		// fx.Provide(remotecollection.NewFetchRemoteCollectionUseCase),
		// fx.Provide(remotecollection.NewListRemoteCollectionsUseCase),

		// Local-based collection use cases
		fx.Provide(collection.NewCreateCollectionUseCase),
		fx.Provide(collection.NewGetCollectionUseCase),
		fx.Provide(collection.NewListCollectionsUseCase),
		fx.Provide(collection.NewUpdateCollectionUseCase),
		fx.Provide(collection.NewDeleteCollectionUseCase),
		fx.Provide(collection.NewMoveCollectionUseCase),
		fx.Provide(collection.NewGetCollectionPathUseCase),

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
		fx.Provide(localfile.NewPathUtilsUseCase),

		// Registration use cases
		fx.Provide(registerUseCase.NewGenerateCredentialsUseCase),
		fx.Provide(registerUseCase.NewCreateLocalUserUseCase),
		fx.Provide(registerUseCase.NewSendRegistrationToServerUseCase),

		// Token refresh usecase
		fx.Provide(refreshtoken.NewRefreshTokenUseCase),
	)
}
