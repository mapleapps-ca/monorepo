package repo

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/repo/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/repo/templatedemailer"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			federateduser.NewRepository,

			// Annotate the constructor to specify which parameter should receive the named dependency
			fx.Annotate(
				templatedemailer.NewTemplatedEmailer,
				fx.ParamTags(`name:"maplefile-module-emailer"`, `name:"papercloud-module-emailer"`),
			),
		),
	)
}
