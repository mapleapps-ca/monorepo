// internal/repo/module.go - Complete updated RepoModule function

package repo

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/collectiondsharingdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/filedto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/medto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/publiclookupdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/recovery"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/recoverydto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/syncstate"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/transaction"
	svc_recovery "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/recovery"
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
		fx.Provide(
			fx.Annotate(
				config.NewLevelDBConfigurationProviderForRecovery,
				fx.ResultTags(`name:"recovery_db_config_provider"`),
			),
		),
		// NEW: Recovery State Storage Configuration Provider
		fx.Provide(
			fx.Annotate(
				config.NewLevelDBConfigurationProviderForRecoveryState,
				fx.ResultTags(`name:"recovery_state_db_config_provider"`),
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
		fx.Provide(
			fx.Annotate(
				leveldb.NewDiskStorage,
				fx.ParamTags(`name:"recovery_db_config_provider"`),
				fx.ResultTags(`name:"recovery_db"`),
			),
		),
		// NEW: Recovery State Storage Instance
		fx.Provide(
			fx.Annotate(
				leveldb.NewDiskStorage,
				fx.ParamTags(`name:"recovery_state_db_config_provider"`),
				fx.ResultTags(`name:"recovery_state_storage"`),
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
		fx.Provide(
			fx.Annotate(
				authdto.NewTokenDTORepository,
				// Note: The token repository now requires these additional dependencies
				// fx.ParamTags(``, ``, ``, ``), // logger, configService, userRepo, tokenDecryptionService
			),
		),
		fx.Provide(authdto.NewEmailVerificationDTORepository),
		fx.Provide(authdto.NewLoginOTTDTORepository),
		fx.Provide(authdto.NewLoginOTTVerificationDTORepository),
		fx.Provide(authdto.NewCompleteLoginDTORepository),
		fx.Provide(authdto.NewRecoveryRepository),

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
		// Cloud Public Lookup DTO repository
		//----------------------------------------------
		fx.Provide(publiclookupdto.NewPublicLookupDTORepository),

		//----------------------------------------------
		// Cloud Me DTO repository
		//----------------------------------------------
		fx.Provide(medto.NewMeDTORepository),

		//----------------------------------------------
		// Recovery repositories
		//----------------------------------------------
		fx.Provide(
			fx.Annotate(
				recovery.NewRecoveryRepository,
				fx.ParamTags(``, `name:"recovery_db"`),
			),
		),

		//----------------------------------------------
		// Recovery DTO repository (for cloud API calls)
		//----------------------------------------------
		fx.Provide(recoverydto.NewRecoveryDTORepository),

		//----------------------------------------------
		// Recovery State Manager (NEW: with dedicated storage)
		//----------------------------------------------
		fx.Provide(
			fx.Annotate(
				svc_recovery.NewRecoveryStateManager,
				fx.ParamTags(``, `name:"recovery_state_storage"`, ``), // logger, storage, recoveryRepo
			),
		),

		//----------------------------------------------
		// Transaction manager
		//----------------------------------------------
		fx.Provide(transaction.NewTransactionManager),
	)
}
