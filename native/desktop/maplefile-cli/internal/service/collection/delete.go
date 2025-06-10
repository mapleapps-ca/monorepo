// internal/service/collection/delete.go
package collection

import (
	"context"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
)

// DeleteService defines the interface for deleting local collections
type DeleteService interface {
	Delete(ctx context.Context, id gocql.UUID) error
	DeleteWithChildren(ctx context.Context, id gocql.UUID) error
}

// deleteService implements the DeleteService interface
type deleteService struct {
	logger        *zap.Logger
	deleteUseCase collection.DeleteCollectionUseCase
}

// NewDeleteService creates a new service for deleting local collections
func NewDeleteService(
	logger *zap.Logger,
	deleteUseCase collection.DeleteCollectionUseCase,
) DeleteService {
	logger = logger.Named("DeleteService")
	return &deleteService{
		logger:        logger,
		deleteUseCase: deleteUseCase,
	}
}

// Delete deletes a local collection by ID
func (s *deleteService) Delete(ctx context.Context, id gocql.UUID) error {
	// Validate input
	if id.String() == "" {
		s.logger.Error("❌ collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Call the use case to delete the collection
	err := s.deleteUseCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("❌ failed to delete local collection",
			zap.String("id", id.String()),
			zap.Error(err))
		return err
	}

	s.logger.Info("✅ local collection deleted successfully",
		zap.String("id", id.String()))
	return nil
}

// DeleteWithChildren deletes a local collection and all its children
func (s *deleteService) DeleteWithChildren(ctx context.Context, id gocql.UUID) error {
	// Validate input
	if id.String() == "" {
		s.logger.Error("❌ collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Call the use case to delete the collection with children

	if err := s.deleteUseCase.DeleteWithChildren(ctx, id); err != nil {
		s.logger.Error("❌ failed to delete local collection with children",
			zap.String("id", id.String()), zap.Error(err))
		return err
	}

	s.logger.Info("✅ local collection and its children deleted successfully",
		zap.String("id", id.String()))
	return nil
}
