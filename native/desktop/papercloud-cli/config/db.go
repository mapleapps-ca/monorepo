// monorepo/native/desktop/papercloud-cli/config/dbconfig.go
package config

import (
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/pkg/storage/leveldb"
)

func NewLevelDBConfigurationProviderForUser() leveldb.LevelDBConfigurationProvider {
	return leveldb.NewLevelDBConfigurationProvider("./", "users")
}

func NewLevelDBConfigurationProviderForCollection() leveldb.LevelDBConfigurationProvider {
	return leveldb.NewLevelDBConfigurationProvider("./", "collections")
}
