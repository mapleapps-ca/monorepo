// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware/utils.go
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
		"/maplefile/api/v1/me":                   true,
		"/maplefile/api/v1/me/delete":            true,
		"/maplefile/api/v1/dashboard":            true,
		"/maplefile/api/v1/collections":          true,
		"/maplefile/api/v1/collections/filtered": true,
		"/maplefile/api/v1/collections/root":     true,
		"/maplefile/api/v1/collections/shared":   true,
		"/maplefile/api/v1/files":                true,
		"/maplefile/api/v1/files/pending":        true, // Three-step workflow file-create endpoint: Start
		"/maplefile/api/v1/files/recent":         true,
		"/maplefile/api/v1/sync/collections":     true,
		"/maplefile/api/v1/sync/files":           true,
	}

	// Pattern matches
	patterns := []string{
		// Collection patterns
		"^/maplefile/api/v1/collections/[a-zA-Z0-9-]+$",           // Individual collection operations
		"^/maplefile/api/v1/collections/[a-zA-Z0-9-]+/files$",     // Collection files
		"^/maplefile/api/v1/collections/[a-zA-Z0-9-]+/move$",      // Move collection
		"^/maplefile/api/v1/collections/[a-zA-Z0-9-]+/share$",     // Share collection
		"^/maplefile/api/v1/collections/[a-zA-Z0-9-]+/members$",   // Collection members
		"^/maplefile/api/v1/collections/[a-zA-Z0-9-]+/archive$",   // Archive collection
		"^/maplefile/api/v1/collections/[a-zA-Z0-9-]+/restore$",   // Restore collection
		"^/maplefile/api/v1/collections-by-parent/[a-zA-Z0-9-]+$", // Collections by parent

		// File patterns
		"^/maplefile/api/v1/files/[a-zA-Z0-9-]+$",              // Individual file operations
		"^/maplefile/api/v1/files/[a-zA-Z0-9-]+/data$",         // File data
		"^/maplefile/api/v1/files/[a-zA-Z0-9-]+/upload-url$",   // File upload URL
		"^/maplefile/api/v1/files/[a-zA-Z0-9-]+/download-url$", // File download URL
		"^/maplefile/api/v1/files/[a-zA-Z0-9-]+/complete$",     // Complete file upload
		"^/maplefile/api/v1/files/[a-zA-Z0-9-]+/archive$",      // Archive file
		"^/maplefile/api/v1/files/[a-zA-Z0-9-]+/restore$",      // Restore file
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
