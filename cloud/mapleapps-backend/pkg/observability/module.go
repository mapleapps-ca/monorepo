// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/observability/module.go
package observability

import (
	"go.uber.org/fx"
)

// Module provides observability components for FX
func Module() fx.Option {
	return fx.Options(
		// Provide core observability components
		fx.Provide(
			NewHealthChecker,
			NewMetricsServer,
		),

		// Register health checks for infrastructure components
		fx.Invoke(registerRealHealthChecks),

		// Start observability server on separate port (:8080)
		fx.Invoke(startObservabilityServer),
	)
}
