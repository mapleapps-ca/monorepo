package middleware

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
)

// Note: This middleware must have `IPAddressMiddleware` executed first before running.
func (mid *middleware) EnforceBlacklistMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Open our program's context based on the request and save the
		// slash-seperated array from our URL path.
		ctx := r.Context()

		ipAddress, _ := ctx.Value(constants.SessionIPAddress).(string)
		proxies, _ := ctx.Value(constants.SessionProxies).(string)

		// Case 1 of 2: Check banned IP addresses.
		if mid.Blacklist.IsBannedIPAddress(ipAddress) {

			// If the client IP address is banned, check to see if the client
			// is making a call to a URL which is not banned, and if the URL
			// is not banned (has not been banned before) then print it to
			// the console logs for future analysis. Else if the URL is banned
			// then don't bother printing to console. The purpose of this code
			// is to not clog the console log with warnings.
			if !mid.Blacklist.IsBannedURL(r.URL.Path) {
				mid.Logger.Warn("rejected request by ip",
					zap.Any("url", r.URL.Path),
					zap.String("ip_address", ipAddress),
					zap.String("proxies", proxies),
					zap.Any("middleware", "EnforceBlacklistMiddleware"))
			}
			http.Error(w, "forbidden at this time", http.StatusForbidden)
			return
		}

		// Case 2 of 2: Check banned URL.
		if mid.Blacklist.IsBannedURL(r.URL.Path) {

			// If the URL is banned, check to see if the client IP address is
			// banned, and if the client has not been banned before then print
			// to console the new offending client ip. Else do not print if
			// the offending client IP address has been banned before. The
			// purpose of this code is to not clog the console log with warnings.
			if !mid.Blacklist.IsBannedIPAddress(ipAddress) {
				mid.Logger.Warn("rejected request by url",
					zap.Any("url", r.URL.Path),
					zap.String("ip_address", ipAddress),
					zap.String("proxies", proxies),
					zap.Any("middleware", "EnforceBlacklistMiddleware"))
			}

			// DEVELOPERS NOTE:
			// Simply return a 404, but in our console log we can see the IP
			// address whom made this call.
			http.Error(w, "does not exist at this time", http.StatusNotFound)
			return
		}

		next(w, r.WithContext(ctx))
	}
}
