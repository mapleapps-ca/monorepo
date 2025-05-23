// internal/repo/module.go
package repo

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/auth"
	localcollection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo/localcollection"
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
		fx.Provide(auth.NewEmailVerificationRepository),
		fx.Provide(auth.NewLoginOTTRepository),
		fx.Provide(auth.NewLoginOTTVerificationRepository),
		fx.Provide(auth.NewCompleteLoginRepository),
		fx.Provide(auth.NewTokenRefresherRepo),

		//----------------------------------------------
		// Local collection repository
		//----------------------------------------------
		fx.Provide(
			fx.Annotate(
				localcollection.NewLocalCollectionRepository,
				fx.ParamTags(``, ``, ``, ``, `name:"collection_db"`),
			),
		),

		// //----------------------------------------------
		// // Remote collection repository
		// //----------------------------------------------
		// fx.Provide(remotecollection.NewRemoteCollectionRepository),

		// //----------------------------------------------
		// // Local file repository
		// //----------------------------------------------
		// fx.Provide(
		// 	fx.Annotate(
		// 		localfile.NewLocalFileRepository,
		// 		fx.ParamTags(``, ``, `name:"file_db"`),
		// 	),
		// ),

		// //----------------------------------------------
		// // Remote file repository
		// //----------------------------------------------
		// fx.Provide(remotefile.NewRemoteFileRepository),

		//----------------------------------------------
		// Transaction manager
		//----------------------------------------------
		fx.Provide(transaction.NewTransactionManager),
	)
}
