package s3

import (
	"github.com/mapleapps-ca/monorepo/cloud/backend/config"

	"go.uber.org/zap"
)

func NewProvider(cfg *config.Configuration, logger *zap.Logger) S3ObjectStorage {
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
