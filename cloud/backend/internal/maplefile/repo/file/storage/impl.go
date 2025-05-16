// cloud/backend/internal/maplefile/repo/file/storage/impl.go
package storage

import (
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	s3storage "github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/object/s3"
)

type FileStorageRepository interface {
	StoreEncryptedData(ownerID string, fileID string, encryptedData []byte) (string, error)
	GetEncryptedData(storagePath string) ([]byte, error)
	DeleteEncryptedData(storagePath string) error
	GeneratePresignedURL(storagePath string, duration time.Duration) (string, error)
}

type fileStorageRepositoryImpl struct {
	Logger  *zap.Logger
	Storage s3storage.S3ObjectStorage
}

func NewRepository(cfg *config.Configuration, logger *zap.Logger, s3 s3storage.S3ObjectStorage) FileStorageRepository {
	return &fileStorageRepositoryImpl{
		Logger:  logger.With(zap.String("repository", "file_storage")),
		Storage: s3,
	}
}
