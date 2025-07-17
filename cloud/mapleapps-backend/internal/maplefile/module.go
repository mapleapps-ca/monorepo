// monorepo/cloud/backend/internal/me/module.go
package maplefile

import (
	"go.uber.org/fx"

	iface "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase"
)

func Module() fx.Option {
	return fx.Options(
		repo.Module(),
		usecase.Module(),
		service.Module(),
		iface.Module(),
	)
}
