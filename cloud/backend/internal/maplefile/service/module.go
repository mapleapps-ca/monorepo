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

			// Collection services
			collection.NewCreateCollectionService,
			collection.NewGetCollectionService,
			collection.NewListUserCollectionsService,
			collection.NewUpdateCollectionService,
			collection.NewDeleteCollectionService,

			// File services
			file.NewCreateFileService,
			file.NewGetFileService,
			file.NewUpdateFileService,
			file.NewDeleteFileService,
			file.NewListFilesByCollectionService,
			file.NewStoreFileDataService,
			file.NewGetFileDataService,
		),
	)
}
