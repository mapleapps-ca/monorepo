// internal/repo/module.go
package repo

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/repo/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/repo/transaction"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/pkg/storage/leveldb"
)

// RepoModule provides the repository-layer--related dependencies
func RepoModule() fx.Option {
	return fx.Options(
		// Provide named LevelDB configuration providers
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

		// Provide specific disk storage for our app.
		fx.Provide(
			fx.Annotate(
				leveldb.NewDiskStorage,
				fx.ParamTags(`name:"user_db_config_provider"`), // Map `NewLevelDBConfigurationProviderForUser` to this `NewDiskStorage`.
				fx.ResultTags(`name:"user_db"`),                // To access this `NewDiskStorage` using key `user_db`.
			),
		),
		fx.Provide(
			fx.Annotate(
				leveldb.NewDiskStorage,
				fx.ParamTags(`name:"collection_db_config_provider"`), // Map `NewLevelDBConfigurationProviderForCollection` to this `NewDiskStorage`.
				fx.ResultTags(`name:"collection_db"`),                // To access this `NewDiskStorage` using key `collection_db`.
			),
		),

		// Provide user repository
		fx.Provide(
			fx.Annotate(
				NewUserRepo,
				fx.ParamTags(``, `name:"user_db"`), // Use the user_db storage for the user repository
			),
		),

		fx.Provide(auth.NewLoginOTTRepository),

		// Transaction manager
		fx.Provide(transaction.NewTransactionManager),
	)
}
