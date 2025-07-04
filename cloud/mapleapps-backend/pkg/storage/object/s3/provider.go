package s3

import (
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"

	"go.uber.org/zap"
)

func NewS3ObjectStorageProvider(cfg *config.Configuration, logger *zap.Logger) S3ObjectStorage {
	configProvider := NewS3ObjectStorageConfigurationProvider(
		cfg.AWS.AccessKey,
		cfg.AWS.SecretKey,
		cfg.AWS.Endpoint,
		cfg.AWS.Region,
		cfg.AWS.BucketName,
		false,
	)

	return NewObjectStorage(configProvider, logger)
}
