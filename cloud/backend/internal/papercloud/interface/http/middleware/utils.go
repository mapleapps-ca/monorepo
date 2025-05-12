// github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/interface/http/middleware/utils.go
package middleware

import (
	"regexp"

	"go.uber.org/zap"
)

type protectedRoute struct {
	pattern string
	regex   *regexp.Regexp
}

var (
	exactPaths    = make(map[string]bool)
	patternRoutes []protectedRoute
)

func init() {
	// Exact matches
	exactPaths = map[string]bool{
		"/papercloud/api/v1/me":          true,
		"/papercloud/api/v1/me/delete":   true,
		"/papercloud/api/v1/dashboard":   true,
		"/papercloud/api/v1/collections": true,
		"/papercloud/api/v1/files":       true,
	}

	// Pattern matches
	patterns := []string{
		// "^/papercloud/api/v1/user/[0-9]+$",                      // Regex designed for non-zero integers.
		// "^/papercloud/api/v1/wallet/[0-9a-f]+$",                 // Regex designed for mongodb ids.
		// "^/papercloud/api/v1/public-wallets/0x[0-9a-fA-F]{40}$", // Regex designed for ethereum addresses.
		// "^/papercloud/api/v1/users/[0-9a-f]+$",                  // Regex designed for mongodb ids.
		"^/papercloud/api/v1/collections/[a-zA-Z0-9-]+$",       // Regex designed for collection IDs
		"^/papercloud/api/v1/collections/[a-zA-Z0-9-]+/files$", // Regex designed for collection IDs
		"^/papercloud/api/v1/files/[a-zA-Z0-9-]+$",             // Regex designed for collection IDs
		"^/papercloud/api/v1/files/[a-zA-Z0-9-]+/data$",        // Regex designed for collection IDs
	}

	// Precompile patterns
	patternRoutes = make([]protectedRoute, len(patterns))
	for i, pattern := range patterns {
		patternRoutes[i] = protectedRoute{
			pattern: pattern,
			regex:   regexp.MustCompile(pattern),
		}
	}
}

func isProtectedPath(logger *zap.Logger, path string) bool {
	// fmt.Println("isProtectedPath - path:", path) // For debugging purposes only.

	// Check exact matches first (O(1) lookup)
	if exactPaths[path] {
		logger.Debug("✅ found via map - url is protected",
			zap.String("path", path))
		return true
	}

	// Check patterns
	for _, route := range patternRoutes {
		if route.regex.MatchString(path) {
			logger.Debug("✅ found via regex - url is protected",
				zap.String("path", path))
			return true
		}
	}

	logger.Debug("❌ not found",
		zap.String("path", path))

	return false
}
