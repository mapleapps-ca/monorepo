// internal/usecase/remotecollection/fetch.go
package remotecollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
)

// FetchRemoteCollectionUseCase defines the interface for fetching a cloud collection
type FetchRemoteCollectionUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) (*remotecollection.RemoteCollection, error)
}

// fetchRemoteCollectionUseCase implements the FetchRemoteCollectionUseCase interface
type fetchRemoteCollectionUseCase struct {
	logger     *zap.Logger
	repository remotecollection.RemoteCollectionRepository
}

// NewFetchRemoteCollectionUseCase creates a new use case for fetching cloud collections
func NewFetchRemoteCollectionUseCase(
	logger *zap.Logger,
	repository remotecollection.RemoteCollectionRepository,
) FetchRemoteCollectionUseCase {
	return &fetchRemoteCollectionUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute fetches a cloud collection by ID
func (uc *fetchRemoteCollectionUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) (*remotecollection.RemoteCollection, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Fetch the collection from the repository
	collection, err := uc.repository.Fetch(ctx, id)
	if err != nil {
		return nil, errors.NewAppError("failed to fetch cloud collection", err)
	}

	if collection == nil {
		return nil, errors.NewAppError("cloud collection not found", nil)
	}

	return collection, nil
}
