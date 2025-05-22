// cloud/backend/internal/maplefile/usecase/fileobjectstorage/generate_presigned_url.go
package fileobjectstorage

import (
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GeneratePresignedURLUseCase interface {
	Execute(storagePath string, duration time.Duration) (string, error)
}

type generatePresignedURLUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileObjectStorageRepository
}

func NewGeneratePresignedURLUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileObjectStorageRepository,
) GeneratePresignedURLUseCase {
	return &generatePresignedURLUseCaseImpl{config, logger, repo}
}

func (uc *generatePresignedURLUseCaseImpl) Execute(storagePath string, duration time.Duration) (string, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if storagePath == "" {
		e["storage_path"] = "Storage path is required"
	}
	if duration <= 0 {
		e["duration"] = "Duration must be greater than 0"
	}
	// Set reasonable limits for presigned URL duration
	maxDuration := 24 * time.Hour // 24 hours max
	if duration > maxDuration {
		e["duration"] = "Duration cannot exceed 24 hours"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating generate presigned URL",
			zap.Any("error", e))
		return "", httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Generate presigned URL.
	//

	url, err := uc.repo.GeneratePresignedURL(storagePath, duration)
	if err != nil {
		uc.logger.Error("Failed to generate presigned URL",
			zap.String("storage_path", storagePath),
			zap.Duration("duration", duration),
			zap.Error(err))
		return "", err
	}

	uc.logger.Debug("Successfully generated presigned URL",
		zap.String("storage_path", storagePath),
		zap.Duration("duration", duration))

	return url, nil
}
