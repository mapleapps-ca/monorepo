// cloud/backend/internal/maplefile/repo/filemetadata/impl.go
package filemetadata

import (
	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

type fileMetadataRepositoryImpl struct {
	Logger *zap.Logger
	//TODO
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, client *mongo.Client) dom_file.FileMetadataRepository {
	loggerp = loggerp.Named("FileMetadataRepository")

	//TODO

	return &fileMetadataRepositoryImpl{
		Logger: loggerp,
		//TODO

	}
}
