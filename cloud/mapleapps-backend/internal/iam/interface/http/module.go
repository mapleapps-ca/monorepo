// cloud/mapleapps-backend/internal/iam/interface/http/module.go
package http

import (
	"go.uber.org/fx"

	commonhttp "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/common"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/gateway"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/middleware"
	unifiedhttp "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/manifold/interface/http"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			middleware.NewMiddleware,
		),
		fx.Provide(
			unifiedhttp.AsRoute(commonhttp.NewGetMapleSendVersionHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewGatewayFederatedUserRegisterHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewGatewayVerifyEmailHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewGatewayRequestLoginOTTHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewGatewayVerifyLoginOTTHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewGatewayCompleteLoginHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewGatewayLogoutHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewGatewayRefreshTokenHTTPHandler),
			// unifiedhttp.AsRoute(gateway.NewGatewayResetPasswordHTTPHandler),
			// unifiedhttp.AsRoute(gateway.NewGatewayForgotPasswordHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewGatewayFederatedUserPublicLookupHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewInitiateRecoveryHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewVerifyRecoveryHTTPHandler),
			unifiedhttp.AsRoute(gateway.NewCompleteRecoveryHTTPHandler),
		),
	)
}
