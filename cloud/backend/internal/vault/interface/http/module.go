// cloud/backend/internal/vault/interface/http/module.go
package http

import (
	"go.uber.org/fx"

	unifiedhttp "github.com/mapleapps-ca/monorepo/cloud/backend/internal/manifold/interface/http"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/interface/http/encryptedfile"
)

// Module registers all HTTP handlers for encrypted files
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			unifiedhttp.AsRoute(encryptedfile.NewCreateEncryptedFileHandler),
			unifiedhttp.AsRoute(encryptedfile.NewGetEncryptedFileByIDHandler),
			unifiedhttp.AsRoute(encryptedfile.NewGetEncryptedFileByFileIDHandler),
			unifiedhttp.AsRoute(encryptedfile.NewUpdateEncryptedFileHandler),
			unifiedhttp.AsRoute(encryptedfile.NewDeleteEncryptedFileHandler),
			unifiedhttp.AsRoute(encryptedfile.NewListEncryptedFilesHandler),
			unifiedhttp.AsRoute(encryptedfile.NewDownloadEncryptedFileHandler),
			unifiedhttp.AsRoute(encryptedfile.NewGetEncryptedFileDownloadURLHandler),
		),
	)
}
