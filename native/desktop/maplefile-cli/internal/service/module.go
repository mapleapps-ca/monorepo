// internal/service/module.go
package service

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filesyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localcollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/register"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/remotecollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/remotefile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/tokenservice"
)

// ServiceModule provides the service-layer--related dependencies
func ServiceModule() fx.Option {
	return fx.Options(
		// Registration service
		fx.Provide(register.NewRegisterService),

		// Auth services
		fx.Provide(auth.NewUserVerificationDataTransformer),
		fx.Provide(auth.NewEmailVerificationService),
		fx.Provide(auth.NewLoginOTTService),
		fx.Provide(auth.NewLoginOTTVerificationService),
		fx.Provide(auth.NewCompleteLoginService),

		// Token refresh service
		fx.Provide(tokenservice.NewTokenRefreshService),

		// Local collection services
		fx.Provide(localcollection.NewCreateService),
		fx.Provide(localcollection.NewGetService),
		fx.Provide(localcollection.NewListService),
		fx.Provide(localcollection.NewUpdateService),
		fx.Provide(localcollection.NewDeleteService),
		fx.Provide(localcollection.NewMoveService),

		// Remote collection services
		fx.Provide(remotecollection.NewCreateService),
		fx.Provide(remotecollection.NewFetchService),
		fx.Provide(remotecollection.NewListService),

		// Collection synchronization services
		fx.Provide(collectionsyncer.NewFindByRemoteIDService),
		fx.Provide(collectionsyncer.NewDownloadService),
		fx.Provide(collectionsyncer.NewUploadService),

		// Local file services
		fx.Provide(localfile.NewCreateService),
		fx.Provide(localfile.NewGetService),
		fx.Provide(localfile.NewListService),
		fx.Provide(localfile.NewUpdateService),
		fx.Provide(localfile.NewDeleteService),
		fx.Provide(localfile.NewImportService),

		// Remote file services
		fx.Provide(remotefile.NewCreateService),
		fx.Provide(remotefile.NewFetchService),
		fx.Provide(remotefile.NewListService),
		fx.Provide(remotefile.NewDeleteService),
		fx.Provide(remotefile.NewUploadService),
		fx.Provide(remotefile.NewDownloadService),

		// File synchronization service
		fx.Provide(filesyncer.NewUploadToRemoteService),
		fx.Provide(filesyncer.NewDownloadToLocalService),
		fx.Provide(filesyncer.NewSyncFileService),
		fx.Provide(filesyncer.NewSyncCollectionService),
	)
}
