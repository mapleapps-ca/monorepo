// internal/repo/module.go
package repo

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/collectiondsharingdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/filedto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/syncstate"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/transaction"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage/leveldb"
)

// RepoModule provides the repository-layer-related dependencies
func RepoModule() fx.Option {
	return fx.Options(
		//----------------------------------------------
		// Provide named LevelDB configuration providers
		//----------------------------------------------
		fx.Provide(
			fx.Annotate(
				config.NewLevelDBConfigurationProviderForUser,
				fx.ResultTags(`name:"user_db_config_provider"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				config.NewLevelDBConfigurationProviderForCollection,
				fx.ResultTags(`name:"collection_db_config_provider"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				config.NewLevelDBConfigurationProviderForFile,
				fx.ResultTags(`name:"file_db_config_provider"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				config.NewLevelDBConfigurationProviderForSyncState,
				fx.ResultTags(`name:"sync_state_db_config_provider"`),
			),
		),

		//----------------------------------------------
		// Provide specific disk storage for our app
		//----------------------------------------------
		fx.Provide(
			fx.Annotate(
				leveldb.NewDiskStorage,
				fx.ParamTags(`name:"user_db_config_provider"`),
				fx.ResultTags(`name:"user_db"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				leveldb.NewDiskStorage,
				fx.ParamTags(`name:"collection_db_config_provider"`),
				fx.ResultTags(`name:"collection_db"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				leveldb.NewDiskStorage,
				fx.ParamTags(`name:"file_db_config_provider"`),
				fx.ResultTags(`name:"file_db"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				leveldb.NewDiskStorage,
				fx.ParamTags(`name:"sync_state_db_config_provider"`),
				fx.ResultTags(`name:"sync_state_db"`),
			),
		),

		//----------------------------------------------
		// Provide user repository
		//----------------------------------------------
		fx.Provide(
			fx.Annotate(
				NewUserRepo,
				fx.ParamTags(``, `name:"user_db"`),
			),
		),

		//----------------------------------------------
		// Auth repositories
		//----------------------------------------------
		fx.Provide(auth.NewTokenRepository),
		fx.Provide(auth.NewEmailVerificationRepository),
		fx.Provide(auth.NewLoginOTTRepository),
		fx.Provide(auth.NewLoginOTTVerificationRepository),
		fx.Provide(auth.NewCompleteLoginRepository),

		//----------------------------------------------
		// Local collection repository
		//----------------------------------------------
		fx.Provide(
			fx.Annotate(
				collection.NewCollectionRepository,
				fx.ParamTags(``, ``, `name:"collection_db"`),
			),
		),

		//----------------------------------------------
		// Cloud collection DTO repository
		//----------------------------------------------
		fx.Provide(collectiondto.NewCollectionDTORepository),

		//----------------------------------------------
		// File repository
		//----------------------------------------------
		fx.Provide(
			fx.Annotate(
				file.NewFileRepository,
				fx.ParamTags(``, ``, `name:"file_db"`),
			),
		),

		//----------------------------------------------
		// Cloud file DTO repository
		//----------------------------------------------
		fx.Provide(filedto.NewFileDTORepository),

		//----------------------------------------------
		// Sync state repository
		//----------------------------------------------
		fx.Provide(
			fx.Annotate(
				syncstate.NewSyncStateRepository,
				fx.ParamTags(``, `name:"sync_state_db"`),
			),
		),

		//----------------------------------------------
		// Cloud Sync DTO repository
		//----------------------------------------------
		fx.Provide(syncdto.NewSyncDTORepository),

		//----------------------------------------------
		// Cloud Collection Sharing DTO repository
		//----------------------------------------------
		fx.Provide(collectiondsharingdto.NewCollectionSharingDTORepository),

		//----------------------------------------------
		// Transaction manager
		//----------------------------------------------
		fx.Provide(transaction.NewTransactionManager),
	)
}
