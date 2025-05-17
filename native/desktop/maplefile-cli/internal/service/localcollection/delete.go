// internal/service/localcollection/delete.go
package localcollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localcollection"
)

// DeleteService defines the interface for deleting local collections
type DeleteService interface {
	Delete(ctx context.Context, id string) error
	DeleteWithChildren(ctx context.Context, id string) error
}

// deleteService implements the DeleteService interface
type deleteService struct {
	logger        *zap.Logger
	deleteUseCase localcollection.DeleteLocalCollectionUseCase
}

// NewDeleteService creates a new service for deleting local collections
func NewDeleteService(
	logger *zap.Logger,
	deleteUseCase localcollection.DeleteLocalCollectionUseCase,
) DeleteService {
	return &deleteService{
		logger:        logger,
		deleteUseCase: deleteUseCase,
	}
}

// Delete deletes a local collection by ID
func (s *deleteService) Delete(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		s.logger.Error("collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Convert ID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		s.logger.Error("invalid collection ID format", zap.String("id", id), zap.Error(err))
		return errors.NewAppError("invalid collection ID format", err)
	}

	// Call the use case to delete the collection
	err = s.deleteUseCase.Execute(ctx, objectID)
	if err != nil {
		s.logger.Error("failed to delete local collection", zap.String("id", id), zap.Error(err))
		return err
	}

	s.logger.Info("local collection deleted successfully", zap.String("id", id))
	return nil
}

// DeleteWithChildren deletes a local collection and all its children
func (s *deleteService) DeleteWithChildren(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		s.logger.Error("collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Convert ID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		s.logger.Error("invalid collection ID format", zap.String("id", id), zap.Error(err))
		return errors.NewAppError("invalid collection ID format", err)
	}

	// Call the use case to delete the collection with children
	err = s.deleteUseCase.DeleteWithChildren(ctx, objectID)
	if err != nil {
		s.logger.Error("failed to delete local collection with children", zap.String("id", id), zap.Error(err))
		return err
	}

	s.logger.Info("local collection and its children deleted successfully", zap.String("id", id))
	return nil
}
