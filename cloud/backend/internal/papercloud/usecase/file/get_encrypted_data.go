// cloud/backend/internal/papercloud/usecase/file/get_encrypted_data.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetEncryptedDataUseCase interface {
	Execute(ctx context.Context, fileID string) ([]byte, error)
}

type getEncryptedDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileRepository
}

func NewGetEncryptedDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileRepository,
) GetEncryptedDataUseCase {
	return &getEncryptedDataUseCaseImpl{config, logger, repo}
}

func (uc *getEncryptedDataUseCaseImpl) Execute(ctx context.Context, fileID string) ([]byte, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if fileID == "" {
		e["file_id"] = "File ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get encrypted data",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get encrypted data.
	//

	data, err := uc.repo.GetEncryptedData(fileID)
	if err != nil {
		return nil, err
	}

	if data == nil || len(data) == 0 {
		uc.logger.Debug("Encrypted data not found",
			zap.String("file_id", fileID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Encrypted data not found")
	}

	return data, nil
}
