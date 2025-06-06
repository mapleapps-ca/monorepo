// cloud/backend/internal/maplefile/repo/fileobjectstorage/impl.go
package fileobjectstorage

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	s3storage "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/object/s3"
)

type fileObjectStorageRepositoryImpl struct {
	Logger  *zap.Logger
	Storage s3storage.S3ObjectStorage
}

func NewRepository(cfg *config.Configuration, logger *zap.Logger, s3 s3storage.S3ObjectStorage) dom_file.FileObjectStorageRepository {
	logger = logger.Named("FileObjectStorageRepository")
	return &fileObjectStorageRepositoryImpl{
		Logger:  logger.With(zap.String("repository", "file_storage")),
		Storage: s3,
	}
}
