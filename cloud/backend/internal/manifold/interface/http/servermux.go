// internal/manifold/interface/http/servermux.go
package http

import (
	"net/http"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/manifold/interface/http/middleware"
)

func NewServeMux(routes []Route, mw middleware.Middleware) *http.ServeMux {
	mux := http.NewServeMux()
	for _, route := range routes {
		// Apply middleware to each route
		wrappedHandler := http.HandlerFunc(mw.Attach(route.ServeHTTP))
		mux.Handle(route.Pattern(), wrappedHandler)
	}
	return mux
}
