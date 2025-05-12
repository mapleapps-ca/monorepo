// cloud/backend/internal/me/module.go
package papercloud

import (
	"go.uber.org/fx"

	iface "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/interface/http"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/repo"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/service"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/usecase"
)

func Module() fx.Option {
	return fx.Options(
		repo.Module(),
		usecase.Module(),
		service.Module(),
		iface.Module(),
	)
}
