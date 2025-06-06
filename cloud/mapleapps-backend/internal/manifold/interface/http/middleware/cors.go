// Example of adding CORS middleware to your Go backend
package middleware

import (
	"net/http"
)

// CORSMiddleware adds Cross-Origin Resource Sharing headers
func (mid *middleware) CORSMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		origin := r.Header.Get("Origin")
		// To allow any origin while supporting credentials, the server must reflect
		// the specific origin from the request. The wildcard "*" is not permitted by
		// browsers in Access-Control-Allow-Origin if Access-Control-Allow-Credentials
		// is true.
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			// Add Vary: Origin header to inform caches that the response differs
			// based on the Origin header. This is important for dynamic ACAO.
			w.Header().Add("Vary", "Origin")
		}
		// If no Origin header is present (e.g., same-origin request or non-CORS request),
		// Access-Control-Allow-Origin is not set. This is fine, as CORS headers are
		// not needed for such requests. The subsequent Access-Control-Allow-Credentials
		// header will be set but will have no effect without ACAO.

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		fn(w, r)
	}
}

// Then in your main.go or where you set up middleware:
// router.Use(middleware.CORSMiddleware)
