// cloud/backend/internal/maplefile/usecase/file/get.go
package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) (*dom_file.File, error)
}

type getFileUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileRepository
}

func NewGetFileUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileRepository,
) GetFileUseCase {
	return &getFileUseCaseImpl{config, logger, repo}
}

func (uc *getFileUseCaseImpl) Execute(ctx context.Context, id primitive.ObjectID) (*dom_file.File, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.IsZero() {
		e["id"] = "File ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file retrieval",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	file, err := uc.repo.Get(id)
	if err != nil {
		return nil, err
	}

	if file == nil {
		uc.logger.Debug("File not found",
			zap.Any("id", id))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	return file, nil
}
