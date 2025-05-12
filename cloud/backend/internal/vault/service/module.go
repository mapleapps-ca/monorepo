// cloud/backend/internal/vault/service/module.go
package service

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/service/encryptedfile"
)

// Module registers all services for encrypted files
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			encryptedfile.NewCreateEncryptedFileService,
			encryptedfile.NewGetEncryptedFileByIDService,
			encryptedfile.NewGetEncryptedFileByFileIDService,
			encryptedfile.NewUpdateEncryptedFileService,
			encryptedfile.NewDeleteEncryptedFileService,
			encryptedfile.NewListEncryptedFilesService,
			encryptedfile.NewDownloadEncryptedFileService,
			encryptedfile.NewGetEncryptedFileDownloadURLService,
		),
	)
}
