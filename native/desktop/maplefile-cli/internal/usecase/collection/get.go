// internal/usecase/collection/get.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

// GetCollectionUseCase defines the interface for getting a local collection
type GetCollectionUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) (*collection.Collection, error)
}

// getCollectionUseCase implements the GetCollectionUseCase interface
type getCollectionUseCase struct {
	logger     *zap.Logger
	repository collection.CollectionRepository
}

// NewGetCollectionUseCase creates a new use case for getting local collections
func NewGetCollectionUseCase(
	logger *zap.Logger,
	repository collection.CollectionRepository,
) GetCollectionUseCase {
	logger = logger.Named("GetCollectionUseCase")
	return &getCollectionUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute retrieves a local collection by ID
func (uc *getCollectionUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) (*collection.Collection, error) {
	uc.logger.Debug("🔎 Attempting to get collection by ID", zap.String("collection_id", id.Hex()))

	// Validate inputs
	if id.IsZero() {
		uc.logger.Error("🚫 collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Get the collection from the repository
	collection, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("💾🔥 failed to get local collection from repository", zap.Error(err), zap.String("collection_id", id.Hex()))
		return nil, errors.NewAppError("failed to get local collection", err)
	}

	if collection == nil {
		uc.logger.Debug("🔍🚫 local collection not found", zap.String("collection_id", id.Hex()))
		return nil, nil
	}

	uc.logger.Info("✅ Successfully retrieved collection", zap.String("collection_id", id.Hex()))
	return collection, nil
}
