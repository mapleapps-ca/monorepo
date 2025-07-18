// monorepo/cloud/mapleapps-backend/internal/maplefile/service/module.go
package service

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/dashboard"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/me"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/storagedailyusage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/storageusageevent"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			// File services
			file.NewSoftDeleteFileService,
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
			file.NewListFileSyncDataService,
			file.NewListRecentFilesService,

			// Dashboard services (now depends on recent files service)
			dashboard.NewGetDashboardService,

			// Me services
			me.NewDeleteMeService,
			me.NewGetMeService,
			me.NewUpdateMeService,
			me.NewVerifyProfileService,

			// Collection services - Basic CRUD
			collection.NewCreateCollectionService,
			collection.NewGetCollectionService,
			collection.NewUpdateCollectionService,
			collection.NewSoftDeleteCollectionService,
			collection.NewArchiveCollectionService,
			collection.NewRestoreCollectionService,

			// Collection services - Hierarchical operations
			collection.NewListUserCollectionsService,
			collection.NewFindCollectionsByParentService,
			collection.NewFindRootCollectionsService,
			collection.NewMoveCollectionService,

			// Collection services - Sharing
			collection.NewShareCollectionService,
			collection.NewRemoveMemberService,
			collection.NewListSharedCollectionsService,

			// Collection services - Filtered operations
			collection.NewGetFilteredCollectionsService,

			// Collection services - Sync Data
			collection.NewGetCollectionSyncDataService,

			// Storage Usage Event services
			storageusageevent.NewGetStorageUsageEventsService,
			storageusageevent.NewGetStorageUsageEventsTrendAnalysisService,
			storageusageevent.NewCreateStorageUsageEventService,

			// Storage Daily Usage services
			storagedailyusage.NewGetStorageDailyUsageTrendService,
			storagedailyusage.NewGetStorageUsageSummaryService,
			storagedailyusage.NewUpdateStorageUsageService,
			storagedailyusage.NewGetStorageUsageByDateRangeService,
		),
	)
}
