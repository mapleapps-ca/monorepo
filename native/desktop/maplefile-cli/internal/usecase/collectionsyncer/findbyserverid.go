// internal/usecase/collectionsyncer/findbyserverid.go
package collectionsyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

// ListLocalCollectionsWithServerIDUseCase is a helper interface for finding local collections by server ID
type ListLocalCollectionsWithServerIDUseCase interface {
	FindByServerID(ctx context.Context, serverID primitive.ObjectID) (*localcollection.LocalCollection, error)
}

// findByServerIDUseCase implements the ListLocalCollectionsWithServerIDUseCase interface
type findByServerIDUseCase struct {
	logger     *zap.Logger
	repository localcollection.LocalCollectionRepository
}

// NewFindByServerIDUseCase creates a new use case for finding local collections by server ID
func NewFindByServerIDUseCase(
	logger *zap.Logger,
	repository localcollection.LocalCollectionRepository,
) ListLocalCollectionsWithServerIDUseCase {
	return &findByServerIDUseCase{
		logger:     logger,
		repository: repository,
	}
}

// FindByServerID finds a local collection by its server ID
func (uc *findByServerIDUseCase) FindByServerID(
	ctx context.Context,
	serverID primitive.ObjectID,
) (*localcollection.LocalCollection, error) {
	// Validate inputs
	if serverID.IsZero() {
		return nil, errors.NewAppError("server ID is required", nil)
	}

	// Get all local collections
	collections, err := uc.repository.List(ctx, localcollection.LocalCollectionFilter{})
	if err != nil {
		return nil, errors.NewAppError("failed to list local collections", err)
	}

	// Find the one matching the server ID
	for _, collection := range collections {
		if collection.ID == serverID {
			return collection, nil
		}
	}

	// Not found
	return nil, nil
}
