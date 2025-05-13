// monorepo/native/desktop/papercloud-cli/internal/app/app.go
package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/pkg/storage/leveldb"
)

// Application name constant
const AppName = "papercloud-cli"

// App represents the CLI application
type App struct {
	rootCmd *cobra.Command
}

// NewApp creates a new CLI application with dependency injection
func NewApp() *App {
	var app App

	fxApp := fx.New(
		// Provide configuration repository
		fx.Provide(func() (config.ConfigRepository, error) {
			return config.NewFileConfigRepository(AppName)
		}),

		// Provide configuration use case
		fx.Provide(config.NewConfigUseCase),

		// Provide named LevelDB configuration providers
		fx.Provide(
			fx.Annotate(
				config.NewLevelDBConfigurationProviderForUser,
				// Assuming the return type is config.LevelDBConfigurationProvider or similar interface
				fx.ResultTags(`name:"user_db_provider"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				config.NewLevelDBConfigurationProviderForCollection,
				// Assuming the return type is config.LevelDBConfigurationProvider or similar interface
				fx.ResultTags(`name:"collection_db_provider"`),
			),
		),

		// Provide specific disk storage for our app.
		fx.Provide(
			fx.Annotate(
				leveldb.NewDiskStorage,
				fx.ParamTags(`name:"user_db_provider"`), // Map `NewLevelDBConfigurationProviderForUser` to this `NewDiskStorage`.
				fx.ResultTags(`name:"user_db"`),         // To access this `NewDiskStorage` using key `user_db`.
			),
		),
		fx.Provide(
			fx.Annotate(
				leveldb.NewDiskStorage,
				fx.ParamTags(`name:"collection_db_provider"`), // Map `NewLevelDBConfigurationProviderForCollection` to this `NewDiskStorage`.
				fx.ResultTags(`name:"collection_db"`),         // To access this `NewDiskStorage` using key `collection_db`.
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

	// Stop the application gracefully on exit (optional but good practice)
	// Consider adding fx.StopTimeout(5*time.Second) to fx.New options
	// and calling fxApp.Stop(context.Background()) in a deferred function
	// or signal handler if resources need cleanup.

	return &app
}

// Execute runs the CLI application
func (a *App) Execute() {
	if err := a.rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
