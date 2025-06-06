// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware/middleware.go
package middleware

import (
	"context"
	"net/http"

	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/user"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"go.uber.org/zap"
)

type Middleware interface {
	Attach(fn http.HandlerFunc) http.HandlerFunc
	Shutdown(ctx context.Context)
}

type middleware struct {
	logger                    *zap.Logger
	jwt                       jwt.Provider
	userGetBySessionIDUseCase uc_user.UserGetBySessionIDUseCase
}

func NewMiddleware(
	logger *zap.Logger,
	jwtp jwt.Provider,
	uc1 uc_user.UserGetBySessionIDUseCase,
) Middleware {
	logger = logger.With(zap.String("module", "maplefile"))
	logger = logger.Named("MapleFile Middleware")
	return &middleware{
		logger:                    logger,
		jwt:                       jwtp,
		userGetBySessionIDUseCase: uc1,
	}
}

// Attach function attaches to HTTP router to apply for every API call.
func (mid *middleware) Attach(fn http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Apply base middleware to all requests
		handler := mid.applyBaseMiddleware(fn)

		// Check if the path requires authentication
		if isProtectedPath(mid.logger, r.URL.Path) {

			// Apply auth middleware for protected paths
			handler = mid.PostJWTProcessorMiddleware(handler)
			handler = mid.JWTProcessorMiddleware(handler)
			// handler = mid.EnforceBlacklistMiddleware(handler)
		}

		handler(w, r)
	}
}

// Attach function attaches to HTTP router to apply for every API call.
func (mid *middleware) applyBaseMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	// Apply middleware in reverse order (bottom up)
	handler := fn
	handler = mid.URLProcessorMiddleware(handler)

	return handler
}

// Shutdown shuts down the middleware.
func (mid *middleware) Shutdown(ctx context.Context) {
	// Log a message to indicate that the HTTP server is shutting down.
}
