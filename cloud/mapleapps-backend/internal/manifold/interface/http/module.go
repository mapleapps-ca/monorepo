// internal/manifold/interface/http/module.go
package http

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/manifold/interface/http/middleware"
)

func Module() fx.Option {
	return fx.Options(
		middleware.Module(), // Include middleware module
		fx.Provide(
			NewUnifiedHTTPServer,
			fx.Annotate(
				NewServeMux,
				fx.ParamTags(`group:"routes"`),
			),
		),
		fx.Provide(
			AsRoute(NewEchoHandler),
			AsRoute(NewGetHealthCheckHTTPHandler),
			// Add other routes here
		),
	)
}
