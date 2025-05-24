// internal/usecase/localfile/change_storage_mode.go
package localfile

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// ChangeStorageModeUseCase defines the interface for changing file storage mode
type ChangeStorageModeUseCase interface {
	Execute(ctx context.Context, fileID primitive.ObjectID, newStorageMode string) (*file.File, error)
}

// changeStorageModeUseCase implements the ChangeStorageModeUseCase interface
type changeStorageModeUseCase struct {
	logger            *zap.Logger
	fileUpdateUseCase fileUseCase.UpdateFileUseCase
}

// NewChangeStorageModeUseCase creates a new use case for changing storage mode
func NewChangeStorageModeUseCase(
	logger *zap.Logger,
	fileUpdateUseCase fileUseCase.UpdateFileUseCase,
) ChangeStorageModeUseCase {
	return &changeStorageModeUseCase{
		logger:            logger,
		fileUpdateUseCase: fileUpdateUseCase,
	}
}

// Execute changes the storage mode of a file
func (uc *changeStorageModeUseCase) Execute(
	ctx context.Context,
	fileID primitive.ObjectID,
	newStorageMode string,
) (*file.File, error) {
	uc.logger.Debug("Changing file storage mode",
		zap.String("fileID", fileID.Hex()),
		zap.String("newStorageMode", newStorageMode))

	// Validate storage mode
	if newStorageMode != file.StorageModeEncryptedOnly &&
		newStorageMode != file.StorageModeDecryptedOnly &&
		newStorageMode != file.StorageModeHybrid {
		return nil, errors.NewAppError(fmt.Sprintf("invalid storage mode: %s", newStorageMode), nil)
	}

	// Update the file
	input := fileUseCase.UpdateFileInput{
		ID:          fileID,
		StorageMode: &newStorageMode,
	}

	updatedFile, err := uc.fileUpdateUseCase.Execute(ctx, input)
	if err != nil {
		return nil, errors.NewAppError("failed to change file storage mode", err)
	}

	return updatedFile, nil
}
