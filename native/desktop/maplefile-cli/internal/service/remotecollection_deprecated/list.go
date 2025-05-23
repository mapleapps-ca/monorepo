// internal/service/remotecollection/list.go
package remotecollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotecollection"
)

// ListOutput represents the result of listing cloud collections
type ListOutput struct {
	Collections []*remotecollection.RemoteCollection `json:"collections"`
	Count       int                                  `json:"count"`
}

// ListService defines the interface for listing cloud collections
type ListService interface {
	ListRoots(ctx context.Context) (*ListOutput, error)
	ListByParent(ctx context.Context, parentID string) (*ListOutput, error)
}

// listService implements the ListService interface
type listService struct {
	logger      *zap.Logger
	listUseCase uc.ListRemoteCollectionsUseCase
}

// NewListService creates a new service for listing cloud collections
func NewListService(
	logger *zap.Logger,
	listUseCase uc.ListRemoteCollectionsUseCase,
) ListService {
	return &listService{
		logger:      logger,
		listUseCase: listUseCase,
	}
}

// ListRoots lists root-level cloud collections
func (s *listService) ListRoots(ctx context.Context) (*ListOutput, error) {
	// Call the use case to list root collections
	collections, err := s.listUseCase.ListRoots(ctx)
	if err != nil {
		s.logger.Error("failed to list root cloud collections", zap.Error(err))
		return nil, err
	}

	return &ListOutput{
		Collections: collections,
		Count:       len(collections),
	}, nil
}

// ListByParent lists cloud collections under a specific parent
func (s *listService) ListByParent(ctx context.Context, parentID string) (*ListOutput, error) {
	// Validate input
	if parentID == "" {
		s.logger.Error("parent ID is required")
		return nil, errors.NewAppError("parent ID is required", nil)
	}

	// Convert parent ID string to ObjectID
	parentObjectID, err := primitive.ObjectIDFromHex(parentID)
	if err != nil {
		s.logger.Error("invalid parent ID format", zap.String("parentID", parentID), zap.Error(err))
		return nil, errors.NewAppError("invalid parent ID format", err)
	}

	// Call the use case to list collections by parent
	collections, err := s.listUseCase.ListByParent(ctx, parentObjectID)
	if err != nil {
		s.logger.Error("failed to list cloud collections by parent", zap.String("parentID", parentID), zap.Error(err))
		return nil, err
	}

	return &ListOutput{
		Collections: collections,
		Count:       len(collections),
	}, nil
}
