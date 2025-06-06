// internal/manifold/interface/http/http.go
package http

import (
	"context"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/manifold/interface/http/middleware"
)

func NewUnifiedHTTPServer(
	lc fx.Lifecycle,
	log *zap.Logger,
	config *config.Configuration,
	mux *http.ServeMux,
	mw middleware.Middleware, // Add middleware dependency
) *http.Server {
	srv := &http.Server{Addr: ":8000", Handler: mux}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info("Starting HTTP server", zap.String("addr", srv.Addr))
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Properly shutdown middleware
			mw.Shutdown()
			return srv.Shutdown(ctx)
		},
	})
	return srv
}
