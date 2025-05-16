// cloud/backend/internal/maplefile/usecase/file/store_encrypted_data.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type StoreEncryptedDataUseCase interface {
	Execute(ctx context.Context, fileID string, encryptedData []byte) error
}

type storeEncryptedDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileRepository
}

func NewStoreEncryptedDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileRepository,
) StoreEncryptedDataUseCase {
	return &storeEncryptedDataUseCaseImpl{config, logger, repo}
}

func (uc *storeEncryptedDataUseCaseImpl) Execute(ctx context.Context, fileID string, encryptedData []byte) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if fileID == "" {
		e["file_id"] = "File ID is required"
	}
	if encryptedData == nil || len(encryptedData) == 0 {
		e["encrypted_data"] = "Encrypted data is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating store encrypted data",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Store encrypted data.
	//

	return uc.repo.StoreEncryptedData(fileID, encryptedData)
}
