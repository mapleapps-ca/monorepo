// internal/app/register_module.go
package app

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/repo/transaction"
	registerService "github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/service/register"
	registerUseCase "github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/usecase/register"
	userUseCase "github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/usecase/user"
)

// RegisterModule provides the registration-related dependencies
func RegisterModule() fx.Option {
	return fx.Options(
		// Transaction manager
		fx.Provide(transaction.NewTransactionManager),

		// User repository use cases
		fx.Provide(userUseCase.NewGetByEmailUseCase),
		fx.Provide(userUseCase.NewUpsertByEmailUseCase),
		fx.Provide(userUseCase.NewDeleteByEmailUseCase),
		fx.Provide(userUseCase.NewListAllUseCase),

		// Registration use cases
		fx.Provide(registerUseCase.NewGenerateCredentialsUseCase),
		fx.Provide(registerUseCase.NewCreateLocalUserUseCase),
		fx.Provide(registerUseCase.NewSendRegistrationToServerUseCase),

		// Registration service
		fx.Provide(registerService.NewRegisterService),
	)
}
