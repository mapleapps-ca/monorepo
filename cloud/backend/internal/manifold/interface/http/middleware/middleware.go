// monorepo/cloud/backend/internal/manifold/interface/http/middleware/middleware.go
package middleware

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/blacklist"
	ipcb "github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/ipcountryblocker"
)

type Middleware interface {
	Attach(fn http.HandlerFunc) http.HandlerFunc
	Shutdown()
}

type middleware struct {
	Logger           *zap.Logger
	Blacklist        blacklist.Provider
	IPCountryBlocker ipcb.Provider
}

func NewMiddleware(
	loggerp *zap.Logger,
	blp blacklist.Provider,
	ipcountryblocker ipcb.Provider,
) Middleware {
	loggerp = loggerp.With(zap.String("module", "manifold"))
	return &middleware{
		Logger:           loggerp,
		Blacklist:        blp,
		IPCountryBlocker: ipcountryblocker,
	}
}

// Attach function attaches to HTTP router to apply for every API call.
func (mid *middleware) Attach(fn http.HandlerFunc) http.HandlerFunc {
	// Attach our middleware handlers here. Please note that all our middleware
	// will start from the bottom and proceed upwards.
	// Ex: `RateLimitMiddleware` will be executed first and
	//     `ProtectedURLsMiddleware` will be executed last.
	fn = mid.EnforceRestrictCountryIPsMiddleware(fn)
	fn = mid.EnforceBlacklistMiddleware(fn)
	fn = mid.IPAddressMiddleware(fn)
	fn = mid.URLProcessorMiddleware(fn)
	fn = mid.RateLimitMiddleware(fn)
	fn = mid.CORSMiddleware(fn)

	return func(w http.ResponseWriter, r *http.Request) {
		// Flow to the next middleware.
		fn(w, r)
	}
}

// Shutdown shuts down the middleware.
func (mid *middleware) Shutdown() {
	// Log a message to indicate that the HTTP server is shutting down.
	mid.Logger.Info("Gracefully shutting down HTTP middleware")
	mid.IPCountryBlocker.Close()
}
