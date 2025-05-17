// internal/usecase/localcollection/get.go
package localcollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

// GetLocalCollectionUseCase defines the interface for getting a local collection
type GetLocalCollectionUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) (*localcollection.LocalCollection, error)
}

// getLocalCollectionUseCase implements the GetLocalCollectionUseCase interface
type getLocalCollectionUseCase struct {
	logger     *zap.Logger
	repository localcollection.LocalCollectionRepository
}

// NewGetLocalCollectionUseCase creates a new use case for getting local collections
func NewGetLocalCollectionUseCase(
	logger *zap.Logger,
	repository localcollection.LocalCollectionRepository,
) GetLocalCollectionUseCase {
	return &getLocalCollectionUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute retrieves a local collection by ID
func (uc *getLocalCollectionUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) (*localcollection.LocalCollection, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Get the collection from the repository
	collection, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewAppError("failed to get local collection", err)
	}

	if collection == nil {
		return nil, errors.NewAppError("local collection not found", nil)
	}

	return collection, nil
}
