// internal/manifold/interface/http/module.go
package http

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/manifold/interface/http/middleware"
)

func Module() fx.Option {
	return fx.Options(
		// Include middleware module
		middleware.Module(),

		fx.Provide(
			NewUnifiedHTTPServer,
			fx.Annotate(
				NewServeMux,
				fx.ParamTags(`group:"routes"`),
			),
		),

		// Register all HTTP routes
		fx.Provide(
			// Core application routes
			AsRoute(NewEchoHandler),

			// Observability routes using pkg/observability components
			AsRoute(NewGetHealthCheckHTTPHandler),
			AsRoute(NewGetReadinessHTTPHandler),
			AsRoute(NewGetLivenessHTTPHandler),
			AsRoute(NewGetMetricsHTTPHandler),
		),
	)
}
