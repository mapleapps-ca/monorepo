package main

import (
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/app"
)

func main() {
	// Create and initialize the application with dependency injection
	application := app.NewApp()

	// Execute the CLI application
	application.Execute()
}
