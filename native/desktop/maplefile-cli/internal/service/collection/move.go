// internal/service/collection/move.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
)

// MoveInput represents the input for moving a local collection
type MoveInput struct {
	ID          string `json:"id"`
	NewParentID string `json:"new_parent_id"`
}

// MoveOutput represents the result of moving a local collection
type MoveOutput struct {
	Collection *collection.Collection `json:"collection"`
}

// MoveService defines the interface for moving local collections
type MoveService interface {
	Move(ctx context.Context, input MoveInput) (*MoveOutput, error)
}

// moveService implements the MoveService interface
type moveService struct {
	logger      *zap.Logger
	moveUseCase uc.MoveCollectionUseCase
}

// NewMoveService creates a new service for moving local collections
func NewMoveService(
	logger *zap.Logger,
	moveUseCase uc.MoveCollectionUseCase,
) MoveService {
	logger = logger.Named("CollectionMoveService")
	return &moveService{
		logger:      logger,
		moveUseCase: moveUseCase,
	}
}

// Move moves a local collection to a new parent
func (s *moveService) Move(ctx context.Context, input MoveInput) (*MoveOutput, error) {
	// Validate inputs
	if input.ID == "" {
		s.logger.Error("⚠️ collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	if input.NewParentID == "" {
		s.logger.Error("⚠️ new parent ID is required")
		return nil, errors.NewAppError("new parent ID is required", nil)
	}

	// Convert ID strings
	objectID, err := gocql.ParseUUID(input.ID)
	if err != nil {
		s.logger.Error("❌ invalid collection ID format", zap.String("id", input.ID), zap.Error(err))
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	newParentObjectID, err := gocql.ParseUUID(input.NewParentID)
	if err != nil {
		s.logger.Error("❌ invalid new parent ID format", zap.String("newParentID", input.NewParentID), zap.Error(err))
		return nil, errors.NewAppError("invalid new parent ID format", err)
	}

	// Prepare the use case input
	useCaseInput := uc.MoveCollectionInput{
		ID:          objectID,
		NewParentID: newParentObjectID,
	}

	// Call the use case to move the collection
	collection, err := s.moveUseCase.Execute(ctx, useCaseInput)
	if err != nil {
		s.logger.Error("💥 failed to move local collection",
			zap.String("id", input.ID),
			zap.String("newParentID", input.NewParentID),
			zap.Error(err))
		return nil, err
	}

	return &MoveOutput{
		Collection: collection,
	}, nil
}
