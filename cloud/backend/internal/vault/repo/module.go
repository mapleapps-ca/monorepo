// cloud/backend/internal/vault/repo/module.go
package repo

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/repo/encryptedfile"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			encryptedfile.NewRepository,
		),
	)
}
