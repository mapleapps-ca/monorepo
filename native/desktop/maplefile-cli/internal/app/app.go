// Package app provides application initialization and dependency injection
package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/cmd"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/repo"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage/mongodb"
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

		fx.Provide(
			// Provide setup for `pkg` items.
			mongodb.NewSecureObjectIDGenerator,
		),

		// Include app modules
		repo.RepoModule(),
		service.ServiceModule(),
		usecase.UseCaseModule(),

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
