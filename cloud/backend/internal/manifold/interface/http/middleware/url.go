package middleware

import (
	"context"
	"net/http"
	"strings"
)

// URLProcessorMiddleware Middleware will split the full URL path into slash-sperated parts and save to
// the context to flow downstream in the app for this particular request.
func (mid *middleware) URLProcessorMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Split path into slash-separated parts, for example, path "/foo/bar"
		// gives p==["foo", "bar"] and path "/" gives p==[""]. Our API starts with
		// "/api", as a result we will start the array slice at "1".
		p := strings.Split(r.URL.Path, "/")[1:]

		// log.Println(p) // For debugging purposes only.

		// Open our program's context based on the request and save the
		// slash-seperated array from our URL path.
		ctx := r.Context()
		ctx = context.WithValue(ctx, "url_split", p)

		// Flow to the next middleware.
		fn(w, r.WithContext(ctx))
	}
}
