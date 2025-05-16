// cloud/backend/internal/maplefile/interface/http/module.go
package http

import (
	"go.uber.org/fx"

	unifiedhttp "github.com/mapleapps-ca/monorepo/cloud/backend/internal/manifold/interface/http"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/interface/http/collection"
	commonhttp "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/interface/http/common"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/interface/http/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/interface/http/me"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/interface/http/middleware"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			middleware.NewMiddleware,
		),
		fx.Provide(
			// Common handlers
			unifiedhttp.AsRoute(commonhttp.NewMapleFileVersionHTTPHandler),

			// Me handlers
			unifiedhttp.AsRoute(me.NewGetMeHTTPHandler),
			unifiedhttp.AsRoute(me.NewPutUpdateMeHTTPHandler),
			unifiedhttp.AsRoute(me.NewDeleteMeHTTPHandler),

			// Collection handlers
			unifiedhttp.AsRoute(collection.NewCreateCollectionHTTPHandler),
			unifiedhttp.AsRoute(collection.NewGetCollectionHTTPHandler),
			unifiedhttp.AsRoute(collection.NewListUserCollectionsHTTPHandler),
			unifiedhttp.AsRoute(collection.NewUpdateCollectionHTTPHandler),
			unifiedhttp.AsRoute(collection.NewDeleteCollectionHTTPHandler),

			// File handlers
			unifiedhttp.AsRoute(file.NewCreateFileHTTPHandler),
			unifiedhttp.AsRoute(file.NewGetFileHTTPHandler),
			unifiedhttp.AsRoute(file.NewUpdateFileHTTPHandler),
			unifiedhttp.AsRoute(file.NewDeleteFileHTTPHandler),
			unifiedhttp.AsRoute(file.NewListFilesByCollectionHTTPHandler),
			unifiedhttp.AsRoute(file.NewStoreFileDataHTTPHandler),
			unifiedhttp.AsRoute(file.NewGetFileDataHTTPHandler),
		),
	)
}
