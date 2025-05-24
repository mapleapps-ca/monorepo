// cloud/backend/internal/maplefile/service/module.go
package service

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/service/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/service/me"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			// Me services
			me.NewDeleteMeService,
			me.NewGetMeService,
			me.NewUpdateMeService,
			me.NewVerifyProfileService,

			// Collection services - Basic CRUD
			collection.NewCreateCollectionService,
			collection.NewGetCollectionService,
			collection.NewUpdateCollectionService,
			collection.NewDeleteCollectionService,

			// Collection services - Hierarchical operations
			collection.NewListUserCollectionsService,
			collection.NewFindCollectionsByParentService,
			collection.NewFindRootCollectionsService,
			collection.NewGetCollectionHierarchyService,
			collection.NewMoveCollectionService,

			// Collection services - Sharing
			collection.NewShareCollectionService,
			collection.NewRemoveMemberService,
			collection.NewListSharedCollectionsService,

			// File services - Original workflow
			file.NewUploadFileService,
			file.NewDeleteFileService,
			file.NewDeleteMultipleFilesService,
			file.NewDownloadFileService,
			file.NewGetFileService,
			file.NewListFilesByCollectionService,
			file.NewUpdateFileService,

			// File services - New three-step workflow
			file.NewCreatePendingFileService,
			file.NewCompleteFileUploadService,
		),
	)
}
