package manifold

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam"
	commonhttp "github.com/mapleapps-ca/monorepo/cloud/backend/internal/manifold/interface/http"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg"
	// "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud"
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
