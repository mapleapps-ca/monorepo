// internal/usecase/collection/path.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

// GetCollectionPathUseCase defines the interface for getting a collection's path
type GetCollectionPathUseCase interface {
	Execute(ctx context.Context, id gocql.UUID) ([]*collection.Collection, error)
}

// getCollectionPathUseCase implements the GetCollectionPathUseCase interface
type getCollectionPathUseCase struct {
	logger     *zap.Logger
	getUseCase GetCollectionUseCase
}

// NewGetCollectionPathUseCase creates a new use case for getting collection paths
func NewGetCollectionPathUseCase(
	logger *zap.Logger,
	getUseCase GetCollectionUseCase,
) GetCollectionPathUseCase {
	logger = logger.Named("GetCollectionPathUseCase")
	return &getCollectionPathUseCase{
		logger:     logger,
		getUseCase: getUseCase,
	}
}

// Execute retrieves the full path (ancestors) of a collection
func (uc *getCollectionPathUseCase) Execute(
	ctx context.Context,
	id gocql.UUID,
) ([]*collection.Collection, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	path := make([]*collection.Collection, 0)

	// Get the initial collection
	current, err := uc.getUseCase.Execute(ctx, id)
	if err != nil {
		return nil, err
	}

	// Follow parent links up to the root
	for current != nil {
		// Add the current collection to the path
		path = append([]*collection.Collection{current}, path...)

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
