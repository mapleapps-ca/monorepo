// internal/app/register_module.go
package service

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/service/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/service/register"
)

// ServiceModule provides the service-layer--related dependencies
func ServiceModule() fx.Option {
	return fx.Options(
		// Registration service
		fx.Provide(register.NewRegisterService),

		// Auth service
		fx.Provide(auth.NewLoginOTTService),
		fx.Provide(auth.NewEmailVerificationService),
	)
}
