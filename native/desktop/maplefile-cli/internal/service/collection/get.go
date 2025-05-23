// internal/service/collection/get.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
)

// GetOutput represents the result of getting a local collection
type GetOutput struct {
	Collection *collection.Collection `json:"collection"`
}

// GetService defines the interface for getting local collections
type GetService interface {
	Get(ctx context.Context, id string) (*GetOutput, error)
	GetPath(ctx context.Context, id string) ([]*collection.Collection, error)
}

// getService implements the GetService interface
type getService struct {
	logger      *zap.Logger
	useCase     uc.GetCollectionUseCase
	pathUseCase uc.GetCollectionPathUseCase
}

// NewGetService creates a new service for getting local collections
func NewGetService(
	logger *zap.Logger,
	useCase uc.GetCollectionUseCase,
	pathUseCase uc.GetCollectionPathUseCase,
) GetService {
	return &getService{
		logger:      logger,
		useCase:     useCase,
		pathUseCase: pathUseCase,
	}
}

// Get retrieves a local collection by ID
func (s *getService) Get(ctx context.Context, id string) (*GetOutput, error) {
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

	// Call the use case to get the collection
	collection, err := s.useCase.Execute(ctx, objectID)
	if err != nil {
		s.logger.Error("failed to get local collection", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return &GetOutput{
		Collection: collection,
	}, nil
}

// GetPath retrieves the full path (ancestors) of a collection
func (s *getService) GetPath(ctx context.Context, id string) ([]*collection.Collection, error) {
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

	// Call the use case to get the path
	path, err := s.pathUseCase.Execute(ctx, objectID)
	if err != nil {
		s.logger.Error("failed to get collection path", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return path, nil
}
