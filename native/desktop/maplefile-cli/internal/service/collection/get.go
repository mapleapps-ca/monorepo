// internal/service/collection/get.go
package collection

import (
	"context"

	"github.com/gocql/gocql"
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
	Get(ctx context.Context, id gocql.UUID) (*GetOutput, error)
	GetPath(ctx context.Context, id gocql.UUID) ([]*collection.Collection, error)
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
	logger = logger.Named("CollectionGetService")
	return &getService{
		logger:      logger,
		useCase:     useCase,
		pathUseCase: pathUseCase,
	}
}

// Get retrieves a local collection by ID
func (s *getService) Get(ctx context.Context, id gocql.UUID) (*GetOutput, error) {
	// Validate input
	if id.String() == "" {
		s.logger.Error("❌ collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Call the use case to get the collection
	collection, err := s.useCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("❌ failed to get local collection",
			zap.String("id", id.String()),
			zap.Error(err))
		return nil, err
	}

	return &GetOutput{
		Collection: collection,
	}, nil
}

// GetPath retrieves the full path (ancestors) of a collection
func (s *getService) GetPath(ctx context.Context, id gocql.UUID) ([]*collection.Collection, error) {
	// Validate input
	if id.String() == "" {
		s.logger.Error("❌ collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Call the use case to get the path
	path, err := s.pathUseCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("❌ failed to get collection path",
			zap.String("id", id.String()),
			zap.Error(err))
		return nil, err
	}

	return path, nil
}
