// cloud/backend/internal/maplefile/usecase/fileobjectstorage/delete_multiple_encrypted_data.go
package fileobjectstorage

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type DeleteMultipleEncryptedDataUseCase interface {
	Execute(storagePaths []string) error
}

type deleteMultipleEncryptedDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileObjectStorageRepository
}

func NewDeleteMultipleEncryptedDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileObjectStorageRepository,
) DeleteMultipleEncryptedDataUseCase {
	logger = logger.Named("DeleteMultipleEncryptedDataUseCase")
	return &deleteMultipleEncryptedDataUseCaseImpl{config, logger, repo}
}

func (uc *deleteMultipleEncryptedDataUseCaseImpl) Execute(storagePaths []string) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if storagePaths == nil || len(storagePaths) == 0 {
		e["storage_paths"] = "Storage paths are required"
	} else {
		for i, path := range storagePaths {
			if path == "" {
				e[fmt.Sprintf("storage_paths[%d]", i)] = "Storage path is required"
			}
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating delete multiple encrypted data",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Delete encrypted data files.
	//

	var errors []error
	successCount := 0

	for _, storagePath := range storagePaths {
		err := uc.repo.DeleteEncryptedData(storagePath)
		if err != nil {
			uc.logger.Error("Failed to delete encrypted data",
				zap.String("storage_path", storagePath),
				zap.Error(err))
			errors = append(errors, fmt.Errorf("failed to delete %s: %w", storagePath, err))
		} else {
			successCount++
			uc.logger.Debug("Successfully deleted encrypted data",
				zap.String("storage_path", storagePath))
		}
	}

	// Log summary
	uc.logger.Info("Completed bulk delete operation",
		zap.Int("total_requested", len(storagePaths)),
		zap.Int("successful_deletions", successCount),
		zap.Int("failed_deletions", len(errors)))

	// If all operations failed, return the first error
	if len(errors) == len(storagePaths) {
		return errors[0]
	}

	// If some operations failed, log but don't return error (partial success)
	if len(errors) > 0 {
		uc.logger.Warn("Some delete operations failed",
			zap.Int("failed_count", len(errors)))
	}

	return nil
}
