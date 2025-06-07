// cloud/backend/internal/me/module.go
package maplefile

import (
	"go.uber.org/fx"

	// iface "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http"// TOOD: Uncomment when transitioned to Cassandra DB
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo"
	// "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service"// TOOD: Uncomment when transitioned to Cassandra DB
	// "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase"// TOOD: Uncomment when transitioned to Cassandra DB
)

func Module() fx.Option {
	return fx.Options(
		repo.Module(),
		// usecase.Module(), // TOOD: Uncomment when transitioned to Cassandra DB
		// service.Module(),// TOOD: Uncomment when transitioned to Cassandra DB
		// iface.Module(),// TOOD: Uncomment when transitioned to Cassandra DB
	)
}
