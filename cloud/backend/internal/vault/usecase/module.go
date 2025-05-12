// cloud/backend/internal/vault/usecase/module.go
package usecase

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/usecase/encryptedfile"
)

// Module registers all encrypted file use cases
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			encryptedfile.NewCreateEncryptedFileUseCase,
			encryptedfile.NewGetEncryptedFileByIDUseCase,
			encryptedfile.NewGetEncryptedFileByFileIDUseCase,
			encryptedfile.NewUpdateEncryptedFileUseCase,
			encryptedfile.NewDeleteEncryptedFileUseCase,
			encryptedfile.NewListEncryptedFilesUseCase,
			encryptedfile.NewDownloadEncryptedFileUseCase,
			encryptedfile.NewGetEncryptedFileDownloadURLUseCase,
		),
	)
}
