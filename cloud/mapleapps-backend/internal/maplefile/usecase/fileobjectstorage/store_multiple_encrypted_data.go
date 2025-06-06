// cloud/backend/internal/maplefile/usecase/fileobjectstorage/store_multiple_encrypted_data.go
package fileobjectstorage

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// EncryptedDataItem represents a single item to be stored
type EncryptedDataItem struct {
	OwnerID       string `json:"owner_id"`
	FileID        string `json:"file_id"`
	EncryptedData []byte `json:"encrypted_data"`
}

// StorageResult represents the result of storing a single item
type StorageResult struct {
	FileID      string `json:"file_id"`
	StoragePath string `json:"storage_path,omitempty"`
	Error       error  `json:"error,omitempty"`
}

type StoreMultipleEncryptedDataUseCase interface {
	Execute(items []EncryptedDataItem) ([]StorageResult, error)
}

type storeMultipleEncryptedDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileObjectStorageRepository
}

func NewStoreMultipleEncryptedDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileObjectStorageRepository,
) StoreMultipleEncryptedDataUseCase {
	logger = logger.Named("StoreMultipleEncryptedDataUseCase")
	return &storeMultipleEncryptedDataUseCaseImpl{config, logger, repo}
}

func (uc *storeMultipleEncryptedDataUseCaseImpl) Execute(items []EncryptedDataItem) ([]StorageResult, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if items == nil || len(items) == 0 {
		e["items"] = "Items are required"
	} else {
		for i, item := range items {
			if item.OwnerID == "" {
				e[fmt.Sprintf("items[%d].owner_id", i)] = "Owner ID is required"
			}
			if item.FileID == "" {
				e[fmt.Sprintf("items[%d].file_id", i)] = "File ID is required"
			}
			if item.EncryptedData == nil || len(item.EncryptedData) == 0 {
				e[fmt.Sprintf("items[%d].encrypted_data", i)] = "Encrypted data is required"
			}
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating store multiple encrypted data",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Store encrypted data files.
	//

	results := make([]StorageResult, len(items))
	successCount := 0

	for i, item := range items {
		storagePath, err := uc.repo.StoreEncryptedData(item.OwnerID, item.FileID, item.EncryptedData)

		results[i] = StorageResult{
			FileID:      item.FileID,
			StoragePath: storagePath,
			Error:       err,
		}

		if err != nil {
			uc.logger.Error("Failed to store encrypted data",
				zap.String("owner_id", item.OwnerID),
				zap.String("file_id", item.FileID),
				zap.Int("data_size", len(item.EncryptedData)),
				zap.Error(err))
		} else {
			successCount++
			uc.logger.Debug("Successfully stored encrypted data",
				zap.String("owner_id", item.OwnerID),
				zap.String("file_id", item.FileID),
				zap.String("storage_path", storagePath),
				zap.Int("data_size", len(item.EncryptedData)))
		}
	}

	// Log summary
	uc.logger.Info("Completed bulk store operation",
		zap.Int("total_requested", len(items)),
		zap.Int("successful_stores", successCount),
		zap.Int("failed_stores", len(items)-successCount))

	return results, nil
}
