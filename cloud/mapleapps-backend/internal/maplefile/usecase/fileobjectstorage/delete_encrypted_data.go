// monorepo/cloud/backend/internal/maplefile/usecase/fileobjectstorage/delete_encrypted_data.go
package fileobjectstorage

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type DeleteEncryptedDataUseCase interface {
	Execute(storagePath string) error
}

type deleteEncryptedDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileObjectStorageRepository
}

func NewDeleteEncryptedDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileObjectStorageRepository,
) DeleteEncryptedDataUseCase {
	logger = logger.Named("DeleteEncryptedDataUseCase")
	return &deleteEncryptedDataUseCaseImpl{config, logger, repo}
}

func (uc *deleteEncryptedDataUseCaseImpl) Execute(storagePath string) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if storagePath == "" {
		e["storage_path"] = "Storage path is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating delete encrypted data",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Delete encrypted data.
	//

	err := uc.repo.DeleteEncryptedData(storagePath)
	if err != nil {
		uc.logger.Error("Failed to delete encrypted data",
			zap.String("storage_path", storagePath),
			zap.Error(err))
		return err
	}

	uc.logger.Info("Successfully deleted encrypted data",
		zap.String("storage_path", storagePath))

	return nil
}
