// internal/service/module.go
package service

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/crypto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filedownload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/fileupload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncstate"
)

// ServiceModule provides the service-layer--related dependencies
func ServiceModule() fx.Option {
	return fx.Options(
		// Crypto service
		fx.Provide(crypto.NewCryptoService),

		// Registration service
		fx.Provide(register.NewRegisterService),

		// Auth services
		fx.Provide(auth.NewUserVerificationDataTransformer),
		fx.Provide(auth.NewEmailVerificationService),
		fx.Provide(auth.NewLoginOTTService),
		fx.Provide(auth.NewLoginOTTVerificationService),
		fx.Provide(auth.NewCompleteLoginService),

		// Collection services
		fx.Provide(collection.NewCreateService),
		fx.Provide(collection.NewGetService),
		fx.Provide(collection.NewListService),
		fx.Provide(collection.NewGetFilteredService),
		fx.Provide(collection.NewUpdateService),
		fx.Provide(collection.NewDeleteService),
		fx.Provide(collection.NewSoftDeleteService),
		fx.Provide(collection.NewMoveService),

		// Local file services
		fx.Provide(localfile.NewAddService),
		fx.Provide(localfile.NewListService),
		fx.Provide(localfile.NewLocalOnlyDeleteService),
		fx.Provide(localfile.NewLockService),
		fx.Provide(localfile.NewUnlockService),

		// File syncer services
		fx.Provide(filesyncer.NewOffloadService),
		fx.Provide(filesyncer.NewOnloadService),
		fx.Provide(filesyncer.NewCloudOnlyDeleteService),

		// Upload file services
		fx.Provide(fileupload.NewUploadService),

		// Download file services
		fx.Provide(filedownload.NewDownloadService),

		// Sync state services
		fx.Provide(syncstate.NewGetService),
		fx.Provide(syncstate.NewSaveService),
		fx.Provide(syncstate.NewResetService),

		// Sync DTO services
		fx.Provide(syncdto.NewGetCollectionsService),
		fx.Provide(syncdto.NewGetFilesService),
		fx.Provide(syncdto.NewGetFullSyncService),
		fx.Provide(syncdto.NewSyncProgressService),
	)
}
