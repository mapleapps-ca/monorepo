// cloud/backend/internal/maplefile/interface/http/module.go
package http

import (
	"go.uber.org/fx"

	unifiedhttp "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/manifold/interface/http"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/collection"
	commonhttp "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/common"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/me"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	// "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/file"
	// "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/me"
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

			// Collection handlers - Basic CRUD
			unifiedhttp.AsRoute(collection.NewCreateCollectionHTTPHandler),
			unifiedhttp.AsRoute(collection.NewGetCollectionHTTPHandler),
			unifiedhttp.AsRoute(collection.NewListUserCollectionsHTTPHandler),
			unifiedhttp.AsRoute(collection.NewUpdateCollectionHTTPHandler),
			unifiedhttp.AsRoute(collection.NewSoftDeleteCollectionHTTPHandler),
			unifiedhttp.AsRoute(collection.NewArchiveCollectionHTTPHandler),
			unifiedhttp.AsRoute(collection.NewRestoreCollectionHTTPHandler),

			// Collection handlers - Hierarchical operations
			unifiedhttp.AsRoute(collection.NewFindCollectionsByParentHTTPHandler),
			unifiedhttp.AsRoute(collection.NewFindRootCollectionsHTTPHandler),
			unifiedhttp.AsRoute(collection.NewMoveCollectionHTTPHandler),

			// Collection handlers - Sharing
			unifiedhttp.AsRoute(collection.NewShareCollectionHTTPHandler),
			unifiedhttp.AsRoute(collection.NewRemoveMemberHTTPHandler),
			unifiedhttp.AsRoute(collection.NewListSharedCollectionsHTTPHandler),

			// Collection handlers - Filtered operations
			unifiedhttp.AsRoute(collection.NewGetFilteredCollectionsHTTPHandler),

			// Sync handlers
			unifiedhttp.AsRoute(collection.NewCollectionSyncHTTPHandler),

			// // // File handlers
			// // unifiedhttp.AsRoute(file.NewSoftDeleteFileHTTPHandler),
			// // unifiedhttp.AsRoute(file.NewDeleteMultipleFilesHTTPHandler),
			// // unifiedhttp.AsRoute(file.NewGetFileHTTPHandler),
			// // unifiedhttp.AsRoute(file.NewListFilesByCollectionHTTPHandler),
			// // unifiedhttp.AsRoute(file.NewUpdateFileHTTPHandler),
			// // unifiedhttp.AsRoute(file.NewCreatePendingFileHTTPHandler),
			// // unifiedhttp.AsRoute(file.NewCompleteFileUploadHTTPHandler),
			// // unifiedhttp.AsRoute(file.NewGetPresignedUploadURLHTTPHandler),
			// // unifiedhttp.AsRoute(file.NewGetPresignedDownloadURLHTTPHandler),
			// // unifiedhttp.AsRoute(file.NewArchiveFileHTTPHandler),
			// // unifiedhttp.AsRoute(file.NewRestoreFileHTTPHandler),

			// // // Sync handlers
			// // unifiedhttp.AsRoute(file.NewFileSyncHTTPHandler),
		),
	)
}
