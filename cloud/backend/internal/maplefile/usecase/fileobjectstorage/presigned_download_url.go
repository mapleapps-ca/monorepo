// cloud/backend/internal/maplefile/usecase/fileobjectstorage/presigned_download_url.go
package fileobjectstorage

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GeneratePresignedDownloadURLUseCase interface {
	Execute(ctx context.Context, storagePath string, duration time.Duration) (string, error)
}

type generatePresignedDownloadURLUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileObjectStorageRepository
}

func NewGeneratePresignedDownloadURLUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileObjectStorageRepository,
) GeneratePresignedDownloadURLUseCase {
	logger = logger.Named("GeneratePresignedDownloadURLUseCase")
	return &generatePresignedDownloadURLUseCaseImpl{config, logger, repo}
}

func (uc *generatePresignedDownloadURLUseCaseImpl) Execute(ctx context.Context, storagePath string, duration time.Duration) (string, error) {
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
		uc.logger.Warn("Failed validating generate presigned download URL",
			zap.Any("error", e))
		return "", httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Generate and get presigned download URL.
	//

	url, err := uc.repo.GeneratePresignedDownloadURL(storagePath, duration)
	if err != nil {
		uc.logger.Error("Failed to generate presigned download URL",
			zap.String("storage_path", storagePath),
			zap.Duration("duration", duration),
			zap.Error(err))
		return "", err
	}

	return url, nil
}
