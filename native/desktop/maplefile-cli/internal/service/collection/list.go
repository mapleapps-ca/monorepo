// internal/service/collection/list.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
)

// ListOutput represents the result of listing local collections
type ListOutput struct {
	Collections []*collection.Collection `json:"collections"`
	Count       int                      `json:"count"`
}

// ListService defines the interface for listing local collections
type ListService interface {
	ListRoots(ctx context.Context) (*ListOutput, error)
	ListByParent(ctx context.Context, parentID gocql.UUID) (*ListOutput, error)
	ListModifiedLocally(ctx context.Context) (*ListOutput, error)
}

// listService implements the ListService interface
type listService struct {
	logger      *zap.Logger
	listUseCase uc.ListCollectionsUseCase
}

// NewListService creates a new service for listing local collections
func NewListService(
	logger *zap.Logger,
	listUseCase uc.ListCollectionsUseCase,
) ListService {
	logger = logger.Named("CollectionListService")
	return &listService{
		logger:      logger,
		listUseCase: listUseCase,
	}
}

// ListRoots lists root-level local collections
func (s *listService) ListRoots(ctx context.Context) (*ListOutput, error) {
	// Call the use case to list root collections
	collections, err := s.listUseCase.ListRoots(ctx)
	if err != nil {
		s.logger.Error("❌ failed to list root collections", zap.Error(err))
		return nil, err
	}

	return &ListOutput{
		Collections: collections,
		Count:       len(collections),
	}, nil
}

// ListByParent lists local collections under a specific parent
func (s *listService) ListByParent(ctx context.Context, parentID gocql.UUID) (*ListOutput, error) {
	// Validate input
	if parentID.String() == "" {
		s.logger.Error("❌ parent ID is required")
		return nil, errors.NewAppError("parent ID is required", nil)
	}

	// Call the use case to list collections by parent
	collections, err := s.listUseCase.ListByParent(ctx, parentID)
	if err != nil {
		s.logger.Error("❌ failed to list collections by parent",
			zap.String("parentID", parentID.String()),
			zap.Error(err))
		return nil, err
	}

	return &ListOutput{
		Collections: collections,
		Count:       len(collections),
	}, nil
}

// ListModifiedLocally lists locally modified collections
func (s *listService) ListModifiedLocally(ctx context.Context) (*ListOutput, error) {
	// Call the use case to list modified collections
	collections, err := s.listUseCase.ListModifiedLocally(ctx)
	if err != nil {
		s.logger.Error("❌ failed to list modified collections", zap.Error(err))
		return nil, err
	}

	return &ListOutput{
		Collections: collections,
		Count:       len(collections),
	}, nil
}
