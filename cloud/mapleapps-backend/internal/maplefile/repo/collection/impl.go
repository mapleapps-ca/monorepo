// cloud/backend/internal/maplefile/repo/collection/impl.go
package collection

import (
	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

type collectionRepositoryImpl struct {
	Logger *zap.Logger
	//TODO: Impl.
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, client *mongo.Client) dom_collection.CollectionRepository {
	loggerp = loggerp.Named("CollectionRepository")

	//TODO: Impl.

	return &collectionRepositoryImpl{
		Logger: loggerp,
		//TODO: Impl.
	}
}
