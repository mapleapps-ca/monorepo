// cloud/backend/internal/vault/module.go
package vault

import (
	"go.uber.org/fx"

	iface "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/interface/http"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/repo"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/service"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/usecase"
)

func Module() fx.Option {
	return fx.Options(
		repo.Module(),
		usecase.Module(),
		service.Module(),
		iface.Module(),
	)
}
