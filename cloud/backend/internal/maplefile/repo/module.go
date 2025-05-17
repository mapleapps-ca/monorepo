// cloud/backend/internal/maplefile/repo/module.go
package repo

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/bannedipaddress"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/templatedemailer"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/user"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			bannedipaddress.NewRepository,
			file.NewFileMetadataRepository,
			file.NewFileStorageRepository,
			file.NewFileRepository,
			user.NewRepository,
			templatedemailer.NewTemplatedEmailer,
			collection.NewRepository,
		),
	)
}
