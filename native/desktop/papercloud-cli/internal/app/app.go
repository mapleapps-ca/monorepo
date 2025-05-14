// Package app provides application initialization and dependency injection
package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/repo"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/pkg/storage/leveldb"
)

// App represents the CLI application
type App struct {
	rootCmd *cobra.Command
}

// NewApp creates a new CLI application with dependency injection
func NewApp() *App {
	var app App

	logger, _ := zap.NewDevelopment()

	fxApp := fx.New(
		// Provide logger
		fx.Provide(
			func() *zap.Logger {
				return logger
			},
		),

		// Provide the configuration service
		config.Module(),

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
				repo.NewUserRepo,
				fx.ParamTags(``, `name:"user_db"`), // Use the user_db storage for the user repository
			),
		),

		// Provide root command
		fx.Provide(cmd.NewRootCmd),

		// Populate the root command for later use
		fx.Populate(&app.rootCmd),
	)

	// Start the application to initialize dependencies
	ctx := context.Background()
	if err := fxApp.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", err)
		os.Exit(1)
	}

	return &app
}

// Execute runs the CLI application
func (a *App) Execute() {
	if err := a.rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
