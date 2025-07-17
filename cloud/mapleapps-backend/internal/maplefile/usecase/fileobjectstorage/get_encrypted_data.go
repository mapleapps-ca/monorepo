// monorepo/cloud/backend/internal/maplefile/usecase/fileobjectstorage/get_encrypted_data.go
package fileobjectstorage

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetEncryptedDataUseCase interface {
	Execute(storagePath string) ([]byte, error)
}

type getEncryptedDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileObjectStorageRepository
}

func NewGetEncryptedDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileObjectStorageRepository,
) GetEncryptedDataUseCase {
	logger = logger.Named("GetEncryptedDataUseCase")
	return &getEncryptedDataUseCaseImpl{config, logger, repo}
}

func (uc *getEncryptedDataUseCaseImpl) Execute(storagePath string) ([]byte, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if storagePath == "" {
		e["storage_path"] = "Storage path is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get encrypted data",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get encrypted data.
	//

	data, err := uc.repo.GetEncryptedData(storagePath)
	if err != nil {
		uc.logger.Error("Failed to get encrypted data",
			zap.String("storage_path", storagePath),
			zap.Error(err))
		return nil, err
	}

	uc.logger.Debug("Successfully retrieved encrypted data",
		zap.String("storage_path", storagePath),
		zap.Int("data_size", len(data)))

	return data, nil
}
