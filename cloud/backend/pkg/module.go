package pkg

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/distributedmutex"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/emailer/mailgun"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/blacklist"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/ipcountryblocker"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/password"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/database/mongodb"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/database/mongodbcache"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/object/s3"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				mailgun.NewMapleFileModuleEmailer,
				fx.ResultTags(`name:"maplefile-module-emailer"`), // Create name for better dependency management handling.
			),
			fx.Annotate(
				mailgun.NewPaperCloudModuleEmailer,
				fx.ResultTags(`name:"papercloud-module-emailer"`), // Create name for better dependency management handling.
			),
		),
		fx.Provide(
			blacklist.NewProvider,
			distributedmutex.NewAdapter,
			ipcountryblocker.NewProvider,
			jwt.NewProvider,
			password.NewProvider,
			mongodb.NewProvider,
			mongodbcache.NewProvider,
			s3.NewProvider,
		),
	)
}
