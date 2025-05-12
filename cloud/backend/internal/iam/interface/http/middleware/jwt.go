// monorepo/cloud/backend/internal/iam/interface/http/middleware/jwt.go
package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
)

func (mid *middleware) JWTProcessorMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract our auth header array.
		reqToken := r.Header.Get("Authorization")

		// For debugging purposes only.
		log.Printf("monorepo/cloud/backend/internal/iam/interface/http/middleware/jwt.go --> reqToken: %v\n", reqToken)

		// Before running our JWT middleware we need to confirm there is an
		// an `Authorization` header to run our middleware. This is an important
		// step!
		if reqToken != "" && strings.Contains(reqToken, "undefined") == false {

			// Special thanks to "poise" via https://stackoverflow.com/a/44700761
			splitToken := strings.Split(reqToken, "JWT ")
			if len(splitToken) < 2 {
				http.Error(w, "not properly formatted authorization header", http.StatusBadRequest)
				return
			}

			reqToken = splitToken[1]
			// log.Println("JWTProcessorMiddleware | reqToken:", reqToken) // For debugging purposes only.

			sessionID, err := mid.jwt.ProcessJWTToken(reqToken)
			// log.Println("JWTProcessorMiddleware | sessionUUID:", sessionUUID) // For debugging purposes only.

			if err == nil {
				// Update our context to save our JWT token content information.
				ctx = context.WithValue(ctx, constants.SessionIsAuthorized, true)
				ctx = context.WithValue(ctx, constants.SessionID, sessionID)

				// Flow to the next middleware with our JWT token saved.
				fn(w, r.WithContext(ctx))
				return
			}

			http.Error(w, fmt.Sprintf("attempting to access a protected endpoint and has session error: %v", err), http.StatusUnauthorized)
			return
		} else {
			http.Error(w, "attempting to access a protected endpoint and authorization not set", http.StatusUnauthorized)
			return
		}
	}
}
