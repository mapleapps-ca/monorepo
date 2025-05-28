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
			collection.NewArchiveCollectionService,
			collection.NewRestoreCollectionService,

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

			// Collection services - Filtered operations
			collection.NewGetFilteredCollectionsService,

			// Collection services - Sync Data
			collection.NewGetCollectionSyncDataService,

			// File services
			file.NewDeleteFileService,
			file.NewDeleteMultipleFilesService,
			file.NewGetFileService,
			file.NewListFilesByCollectionService,
			file.NewUpdateFileService,
			file.NewCreatePendingFileService,
			file.NewCompleteFileUploadService,
			file.NewGetPresignedUploadURLService,
			file.NewGetPresignedDownloadURLService,
			file.NewListFilesByCreatedByUserIDService,
			file.NewListFilesByOwnerIDService,
			file.NewArchiveFileService,
			file.NewRestoreFileService,
			file.NewGetFileSyncDataService,
		),
	)
}
