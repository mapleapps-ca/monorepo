// internal/manifold/interface/http/middleware/module.go
package middleware

import (
	"go.uber.org/fx"
)

// Module provides middleware components
func Module() fx.Option {
	return fx.Provide(
		NewMiddleware,
	)
}
