package middleware

import (
	"context"
	"net/http"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
)

func (mid *middleware) IPAddressMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the IPAddress. Code taken from: https://stackoverflow.com/a/55738279
		IPAddress := r.Header.Get("X-Real-Ip")
		if IPAddress == "" {
			IPAddress = r.Header.Get("X-Forwarded-For")
		}
		if IPAddress == "" {
			IPAddress = r.RemoteAddr
		}

		// Save our IP address to the context.
		ctx := r.Context()
		ctx = context.WithValue(ctx, constants.SessionIPAddress, IPAddress)
		fn(w, r.WithContext(ctx)) // Flow to the next middleware.
	}
}
