// cloud/backend/internal/maplefile/repo/module.go
package repo

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/fileobjectstorage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/templatedemailer"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			filemetadata.NewRepository,
			fileobjectstorage.NewRepository,
			user.NewRepository,
			templatedemailer.NewTemplatedEmailer,
			collection.NewRepository,
		),
	)
}
