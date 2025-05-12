package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

func (mid *middleware) RateLimitMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	// Special thanks: https://ubogdan.com/2021/09/ip-based-rate-limit-middleware-using-go.uber.org/ratelimit/
	var lmap sync.Map

	return func(w http.ResponseWriter, r *http.Request) {
		// Open our program's context based on the request and save the
		// slash-seperated array from our URL path.
		ctx := r.Context()

		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			mid.Logger.Error("invalid RemoteAddr", zap.Any("err", err), zap.Any("middleware", "RateLimitMiddleware"))
			http.Error(w, fmt.Sprintf("invalid RemoteAddr: %s", err), http.StatusInternalServerError)
			return
		}

		lif, ok := lmap.Load(host)
		if !ok {
			lif = ratelimit.New(50) // per second.
		}

		lm, ok := lif.(ratelimit.Limiter)
		if !ok {
			mid.Logger.Error("internal middleware error: typecast failed", zap.Any("middleware", "RateLimitMiddleware"))
			http.Error(w, "internal middleware error: typecast failed", http.StatusInternalServerError)
			return
		}

		lm.Take()
		lmap.Store(host, lm)

		// Flow to the next middleware.
		fn(w, r.WithContext(ctx))
	}
}
