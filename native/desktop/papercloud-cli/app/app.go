// app/app.go
package app

import (
	"context"
	"fmt"
	"os"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/cmd"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
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
