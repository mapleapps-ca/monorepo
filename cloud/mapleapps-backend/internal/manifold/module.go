// monorepo/cloud/mapleapps-backend/internal/manifold/module.go
package manifold

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam"
	commonhttp "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/manifold/interface/http"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg"
	// "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/papercloud"
)

func Module() fx.Option {
	return fx.Options(
		// Step 1: Core infrastructure components (databases, caches, storage)
		pkg.Module(),

		// Step 2: Domain modules
		iam.Module(),
		maplefile.Module(),
		// papercloud.Module(), // Coming soon

		// Step 3: HTTP interface layer (depends on observability for handlers)
		commonhttp.Module(),

		// Step 4: Ensure HTTP server is instantiated
		fx.Invoke(func(*http.Server) {}),
	)
}
