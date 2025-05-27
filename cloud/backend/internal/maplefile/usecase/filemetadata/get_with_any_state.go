// cloud/backend/internal/maplefile/usecase/filemetadata/get_with_any_state.go
package filemetadata

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetFileMetadataWithAnyStateUseCase interface {
	Execute(id primitive.ObjectID) (*dom_file.File, error)
}

type getFileMetadataWithAnyStateUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewGetFileMetadataWithAnyStateUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) GetFileMetadataWithAnyStateUseCase {
	return &getFileMetadataWithAnyStateUseCaseImpl{config, logger, repo}
}

func (uc *getFileMetadataWithAnyStateUseCaseImpl) Execute(id primitive.ObjectID) (*dom_file.File, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.IsZero() {
		e["id"] = "File ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata retrieval (any state)",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database regardless of state.
	//

	file, err := uc.repo.GetWithAnyState(id)
	if err != nil {
		return nil, err
	}

	if file == nil {
		uc.logger.Debug("File metadata not found (any state)",
			zap.Any("id", id))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	return file, nil
}

// Also need to extend the existing GetFileMetadataUseCase interface to include this method
// Update to cloud/backend/internal/maplefile/usecase/filemetadata/get.go

// Add this method to the existing getFileMetadataUseCaseImpl:
func (uc *getFileMetadataUseCaseImpl) ExecuteWithAnyState(id primitive.ObjectID) (*dom_file.File, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.IsZero() {
		e["id"] = "File ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata retrieval (any state)",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database regardless of state.
	//

	file, err := uc.repo.GetWithAnyState(id)
	if err != nil {
		return nil, err
	}

	if file == nil {
		uc.logger.Debug("File metadata not found (any state)",
			zap.Any("id", id))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	return file, nil
}
