// monorepo/cloud/mapleapps-backend/internal/iam/service/module_updated.go
package service

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service/gateway"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service/me"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service/token"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			// Token services
			token.NewTokenVerifyService,
			token.NewTokenGetSessionService,

			// Token encryption service
			gateway.NewTokenEncryptionService,

			// Gateway services
			gateway.NewGatewayFederatedUserRegisterService,
			gateway.NewGatewayVerifyEmailService,

			// E2EE login services
			gateway.NewGatewayRequestLoginOTTService,
			gateway.NewGatewayVerifyLoginOTTService,
			gateway.NewGatewayCompleteLoginService,

			// Other services
			gateway.NewGatewayLogoutService,
			gateway.NewGatewayRefreshTokenService,
			gateway.NewGatewayFederatedUserPublicLookupService,
			gateway.NewInitiateRecoveryService,
			gateway.NewVerifyRecoveryService,
			gateway.NewCompleteRecoveryService,

			// Me services
			me.NewGetMeService,
			me.NewUpdateMeService,
			me.NewDeleteMeService,
			me.NewVerifyProfileService,

			// Storage services
			me.NewGetStorageUsageService,
			me.NewUpgradePlanService,
		),
	)
}
