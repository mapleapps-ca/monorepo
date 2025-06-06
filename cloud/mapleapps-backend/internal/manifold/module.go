package manifold

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam"
	commonhttp "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/manifold/interface/http"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg"

	// "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/papercloud"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile"
)

func Module() fx.Option {
	return fx.Options(
		pkg.Module(), // Shared utilities, types, and helpers used across layers
		commonhttp.Module(),
		iam.Module(),
		maplefile.Module(),
		// papercloud.Module(),
		fx.Invoke(func(*http.Server) {}),
	)
}
