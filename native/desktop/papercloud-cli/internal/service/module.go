// internal/app/register_module.go
package service

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/service/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/service/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/service/tokenservice"
)

// ServiceModule provides the service-layer--related dependencies
func ServiceModule() fx.Option {
	return fx.Options(
		// Registration service
		fx.Provide(register.NewRegisterService),

		// Auth service
		fx.Provide(auth.NewUserVerificationDataTransformer),
		fx.Provide(auth.NewEmailVerificationService),
		fx.Provide(auth.NewLoginOTTService),
		fx.Provide(auth.NewLoginOTTVerificationService),
		fx.Provide(auth.NewCompleteLoginService),

		// Token refresh service
		fx.Provide(tokenservice.NewTokenRefreshService),
	)
}
