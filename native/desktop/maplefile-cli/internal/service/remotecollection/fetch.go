// internal/service/remotecollection/fetch.go
package remotecollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotecollection"
)

// FetchOutput represents the result of fetching a remote collection
type FetchOutput struct {
	Collection *remotecollection.RemoteCollection `json:"collection"`
}

// FetchService defines the interface for fetching remote collections
type FetchService interface {
	Fetch(ctx context.Context, id string) (*FetchOutput, error)
}

// fetchService implements the FetchService interface
type fetchService struct {
	logger       *zap.Logger
	fetchUseCase uc.FetchRemoteCollectionUseCase
}

// NewFetchService creates a new service for fetching remote collections
func NewFetchService(
	logger *zap.Logger,
	fetchUseCase uc.FetchRemoteCollectionUseCase,
) FetchService {
	return &fetchService{
		logger:       logger,
		fetchUseCase: fetchUseCase,
	}
}

// Fetch retrieves a remote collection by ID
func (s *fetchService) Fetch(ctx context.Context, id string) (*FetchOutput, error) {
	// Validate input
	if id == "" {
		s.logger.Error("collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Convert ID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		s.logger.Error("invalid collection ID format", zap.String("id", id), zap.Error(err))
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Call the use case to fetch the collection
	collection, err := s.fetchUseCase.Execute(ctx, objectID)
	if err != nil {
		s.logger.Error("failed to fetch remote collection", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return &FetchOutput{
		Collection: collection,
	}, nil
}
