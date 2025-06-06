// cloud/backend/internal/maplefile/repo/collection/impl.go
package collection

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

type collectionRepositoryImpl struct {
	Logger *zap.Logger
	//TODO: Impl.
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger) dom_collection.CollectionRepository {
	loggerp = loggerp.Named("CollectionRepository")

	//TODO: Impl.

	return &collectionRepositoryImpl{
		Logger: loggerp,
		//TODO: Impl.
	}
}
