// internal/usecase/localfile/get_files_by_collection.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// GetFilesByCollectionUseCase defines the interface for getting files by collection
type GetFilesByCollectionUseCase interface {
	Execute(ctx context.Context, collectionID primitive.ObjectID) ([]*file.File, error)
}

// getFilesByCollectionUseCase implements the GetFilesByCollectionUseCase interface
type getFilesByCollectionUseCase struct {
	logger                       *zap.Logger
	listFilesByCollectionUseCase fileUseCase.ListFilesByCollectionUseCase
}

// NewGetFilesByCollectionUseCase creates a new use case for getting files by collection
func NewGetFilesByCollectionUseCase(
	logger *zap.Logger,
	listFilesByCollectionUseCase fileUseCase.ListFilesByCollectionUseCase,
) GetFilesByCollectionUseCase {
	return &getFilesByCollectionUseCase{
		logger:                       logger,
		listFilesByCollectionUseCase: listFilesByCollectionUseCase,
	}
}

// Execute retrieves all files in a collection
func (uc *getFilesByCollectionUseCase) Execute(
	ctx context.Context,
	collectionID primitive.ObjectID,
) ([]*file.File, error) {
	uc.logger.Debug("Getting local files by collection", zap.String("collectionID", collectionID.Hex()))

	files, err := uc.listFilesByCollectionUseCase.Execute(ctx, collectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get local files by collection", err)
	}

	return files, nil
}
