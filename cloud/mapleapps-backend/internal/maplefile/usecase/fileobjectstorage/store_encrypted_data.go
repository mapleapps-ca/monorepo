// monorepo/cloud/backend/internal/maplefile/usecase/fileobjectstorage/store_encrypted_data.go
package fileobjectstorage

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type StoreEncryptedDataUseCase interface {
	Execute(ownerID string, fileID string, encryptedData []byte) (string, error)
}

type storeEncryptedDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileObjectStorageRepository
}

func NewStoreEncryptedDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileObjectStorageRepository,
) StoreEncryptedDataUseCase {
	logger = logger.Named("StoreEncryptedDataUseCase")
	return &storeEncryptedDataUseCaseImpl{config, logger, repo}
}

func (uc *storeEncryptedDataUseCaseImpl) Execute(ownerID string, fileID string, encryptedData []byte) (string, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if ownerID == "" {
		e["owner_id"] = "Owner ID is required"
	}
	if fileID == "" {
		e["file_id"] = "File ID is required"
	}
	if encryptedData == nil || len(encryptedData) == 0 {
		e["encrypted_data"] = "Encrypted data is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating store encrypted data",
			zap.Any("error", e))
		return "", httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Store encrypted data.
	//

	storagePath, err := uc.repo.StoreEncryptedData(ownerID, fileID, encryptedData)
	if err != nil {
		uc.logger.Error("Failed to store encrypted data",
			zap.String("owner_id", ownerID),
			zap.String("file_id", fileID),
			zap.Int("data_size", len(encryptedData)),
			zap.Error(err))
		return "", err
	}

	uc.logger.Info("Successfully stored encrypted data",
		zap.String("owner_id", ownerID),
		zap.String("file_id", fileID),
		zap.String("storage_path", storagePath),
		zap.Int("data_size", len(encryptedData)))

	return storagePath, nil
}
