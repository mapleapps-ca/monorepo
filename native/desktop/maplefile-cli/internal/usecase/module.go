// internal/usecase/module.go
package usecase

import (
	"go.uber.org/fx"

	authUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/crypto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localcollection"
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

		// Local collection use cases
		fx.Provide(localcollection.NewCreateLocalCollectionUseCase),
		fx.Provide(localcollection.NewGetLocalCollectionUseCase),
		fx.Provide(localcollection.NewListLocalCollectionsUseCase),
		fx.Provide(localcollection.NewUpdateLocalCollectionUseCase),
		fx.Provide(localcollection.NewDeleteLocalCollectionUseCase),
		fx.Provide(localcollection.NewMoveLocalCollectionUseCase),
		fx.Provide(localcollection.NewGetLocalCollectionPathUseCase),

		// // Remote collection use cases
		// fx.Provide(remotecollection.NewCreateRemoteCollectionUseCase),
		// fx.Provide(remotecollection.NewFetchRemoteCollectionUseCase),
		// fx.Provide(remotecollection.NewListRemoteCollectionsUseCase),

		// Registration use cases
		fx.Provide(registerUseCase.NewGenerateCredentialsUseCase),
		fx.Provide(registerUseCase.NewCreateLocalUserUseCase),
		fx.Provide(registerUseCase.NewSendRegistrationToServerUseCase),

		// Token refresh usecase
		fx.Provide(refreshtoken.NewRefreshTokenUseCase),
	)
}
