// Package config provides a unified API for managing application configuration
// Location: monorepo/native/desktop/papercloud-cli/internal/config/config.go
package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/pkg/storage/leveldb"
)

// LevelDB support functions - updated to use app directory path

// NewLevelDBConfigurationProviderForUser returns a LevelDB configuration provider for users
func NewLevelDBConfigurationProviderForUser() leveldb.LevelDBConfigurationProvider {
	// The proper way to do this would be to use the ConfigService's GetAppDirPath,
	// but since this is a static function, we'll use the default path directly
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed getting user config directory with error: %v\n", err)
	}

	// Use the app directory for storing the LevelDB database
	appDir := filepath.Join(configDir, AppName)

	return leveldb.NewLevelDBConfigurationProvider(appDir, "users")
}

// NewLevelDBConfigurationProviderForCollection returns a LevelDB configuration provider for collections
func NewLevelDBConfigurationProviderForCollection() leveldb.LevelDBConfigurationProvider {
	// The proper way to do this would be to use the ConfigService's GetAppDirPath,
	// but since this is a static function, we'll use the default path directly
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed getting user config directory with error: %v\n", err)
	}

	// Use the app directory for storing the LevelDB database
	appDir := filepath.Join(configDir, AppName)

	return leveldb.NewLevelDBConfigurationProvider(appDir, "collections")
}
