// internal/usecase/localcollection/path.go
package localcollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

// GetLocalCollectionPathUseCase defines the interface for getting a collection's path
type GetLocalCollectionPathUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) ([]*localcollection.LocalCollection, error)
}

// getLocalCollectionPathUseCase implements the GetLocalCollectionPathUseCase interface
type getLocalCollectionPathUseCase struct {
	logger     *zap.Logger
	getUseCase GetLocalCollectionUseCase
}

// NewGetLocalCollectionPathUseCase creates a new use case for getting collection paths
func NewGetLocalCollectionPathUseCase(
	logger *zap.Logger,
	getUseCase GetLocalCollectionUseCase,
) GetLocalCollectionPathUseCase {
	return &getLocalCollectionPathUseCase{
		logger:     logger,
		getUseCase: getUseCase,
	}
}

// Execute retrieves the full path (ancestors) of a collection
func (uc *getLocalCollectionPathUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) ([]*localcollection.LocalCollection, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	path := make([]*localcollection.LocalCollection, 0)

	// Get the initial collection
	current, err := uc.getUseCase.Execute(ctx, id)
	if err != nil {
		return nil, err
	}

	// Follow parent links up to the root
	for current != nil {
		// Add the current collection to the path
		path = append([]*localcollection.LocalCollection{current}, path...)

		// If this is a root collection, we're done
		if current.ParentID.IsZero() {
			break
		}

		// Get the parent
		current, err = uc.getUseCase.Execute(ctx, current.ParentID)
		if err != nil {
			return nil, errors.NewAppError("failed to retrieve parent collection", err)
		}
	}

	return path, nil
}
