// internal/config/leveldb.go
package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage/leveldb"
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

// NewLevelDBConfigurationProviderForFile returns a LevelDB configuration provider for files
func NewLevelDBConfigurationProviderForFile() leveldb.LevelDBConfigurationProvider {
	// The proper way to do this would be to use the ConfigService's GetAppDirPath,
	// but since this is a static function, we'll use the default path directly
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed getting user config directory with error: %v\n", err)
	}

	// Use the app directory for storing the LevelDB database
	appDir := filepath.Join(configDir, AppName)

	return leveldb.NewLevelDBConfigurationProvider(appDir, "files")
}

// NewLevelDBConfigurationProviderForSyncState returns a LevelDB configuration provider for sync state
func NewLevelDBConfigurationProviderForSyncState() leveldb.LevelDBConfigurationProvider {
	// The proper way to do this would be to use the ConfigService's GetAppDirPath,
	// but since this is a static function, we'll use the default path directly
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed getting user config directory with error: %v\n", err)
	}

	// Use the app directory for storing the LevelDB database
	appDir := filepath.Join(configDir, AppName)

	return leveldb.NewLevelDBConfigurationProvider(appDir, "syncstate")
}

// NewLevelDBConfigurationProviderForRecovery returns a LevelDB configuration provider for recovery data
func NewLevelDBConfigurationProviderForRecovery() leveldb.LevelDBConfigurationProvider {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed getting user config directory with error: %v\n", err)
	}

	// Use the app directory for storing the LevelDB database
	appDir := filepath.Join(configDir, AppName)

	return leveldb.NewLevelDBConfigurationProvider(appDir, "recovery")
}

// NewLevelDBConfigurationProviderForRecoveryState creates a LevelDB configuration provider for recovery state storage
func NewLevelDBConfigurationProviderForRecoveryState(configService ConfigService) leveldb.LevelDBConfigurationProvider {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed getting user config directory with error: %v\n", err)
	}

	// Use the app directory for storing the LevelDB database
	appDir := filepath.Join(configDir, AppName)

	return leveldb.NewLevelDBConfigurationProvider(appDir, "recovery_state")
}
