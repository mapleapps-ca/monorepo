// cloud/backend/internal/maplefile/repo/file/impl.go
package file

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/file/metadata"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/file/storage"
	s3storage "github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/object/s3"
)

// Composite repository implementing the domain FileRepository interface
type fileRepositoryImpl struct {
	logger   *zap.Logger
	metadata metadata.FileMetadataRepository
	storage  storage.FileStorageRepository
}

// Constructor function for FileRepository
func NewFileRepository(
	logger *zap.Logger,
	metadataRepo metadata.FileMetadataRepository,
	storageRepo storage.FileStorageRepository,
) dom_file.FileRepository {
	return &fileRepositoryImpl{
		logger:   logger.With(zap.String("repository", "file")),
		metadata: metadataRepo,
		storage:  storageRepo,
	}
}

// Constructor for metadata repository
func NewFileMetadataRepository(
	cfg *config.Configuration,
	logger *zap.Logger,
	client *mongo.Client,
) metadata.FileMetadataRepository {
	return metadata.NewRepository(cfg, logger, client)
}

// Constructor for storage repository
func NewFileStorageRepository(
	cfg *config.Configuration,
	logger *zap.Logger,
	s3 s3storage.S3ObjectStorage,
) storage.FileStorageRepository {
	return storage.NewRepository(cfg, logger, s3)
}
