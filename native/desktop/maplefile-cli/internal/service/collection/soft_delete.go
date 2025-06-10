// internal/service/collection/soft_delete.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
)

// SoftDeleteService defines the interface for soft deleting local collections
type SoftDeleteService interface {
	SoftDelete(ctx context.Context, id string) error
	SoftDeleteWithChildren(ctx context.Context, id string) error
	Archive(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
}

// softDeleteService implements the SoftDeleteService interface
type softDeleteService struct {
	logger        *zap.Logger
	getUseCase    uc.GetCollectionUseCase
	updateUseCase uc.UpdateCollectionUseCase
	listUseCase   uc.ListCollectionsUseCase
}

// NewSoftDeleteService creates a new service for soft deleting local collections
func NewSoftDeleteService(
	logger *zap.Logger,
	getUseCase uc.GetCollectionUseCase,
	updateUseCase uc.UpdateCollectionUseCase,
	listUseCase uc.ListCollectionsUseCase,
) SoftDeleteService {
	logger = logger.Named("SoftDeleteService")
	return &softDeleteService{
		logger:        logger,
		getUseCase:    getUseCase,
		updateUseCase: updateUseCase,
		listUseCase:   listUseCase,
	}
}

// SoftDelete marks a collection as deleted by updating its state
func (s *softDeleteService) SoftDelete(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		s.logger.Error("❌ collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Convert ID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		s.logger.Error("❌ invalid collection ID format", zap.String("id", id), zap.Error(err))
		return errors.NewAppError("invalid collection ID format", err)
	}

	// Get the collection to validate it exists and check current state
	existingCollection, err := s.getUseCase.Execute(ctx, objectID)
	if err != nil {
		s.logger.Error("❌ failed to get collection for soft delete", zap.String("id", id), zap.Error(err))
		return err
	}

	// Check if state transition is valid
	if err := collection.IsValidStateTransition(existingCollection.State, collection.CollectionStateDeleted); err != nil {
		s.logger.Error("⚠️ invalid state transition for soft delete",
			zap.String("id", id),
			zap.String("currentState", existingCollection.State),
			zap.Error(err))
		return errors.NewAppError("cannot delete collection in current state", err)
	}

	// Update collection state to deleted
	newState := collection.CollectionStateDeleted
	updateInput := uc.UpdateCollectionInput{
		ID:    objectID,
		State: &newState,
	}

	_, err = s.updateUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ failed to soft delete collection", zap.String("id", id), zap.Error(err))
		return err
	}

	s.logger.Info("✅ collection soft deleted successfully",
		zap.String("id", id),
		zap.String("previousState", existingCollection.State),
		zap.String("newState", collection.CollectionStateDeleted))

	return nil
}

// SoftDeleteWithChildren soft deletes a collection and all its children
func (s *softDeleteService) SoftDeleteWithChildren(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		s.logger.Error("❌ collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Convert ID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		s.logger.Error("❌ invalid collection ID format", zap.String("id", id), zap.Error(err))
		return errors.NewAppError("invalid collection ID format", err)
	}

	// Get all children of this collection
	children, err := s.listUseCase.ListByParent(ctx, objectID)
	if err != nil {
		return errors.NewAppError("failed to list child collections", err)
	}

	// Soft delete each child recursively
	for _, child := range children {
		err = s.SoftDeleteWithChildren(ctx, child.ID.String())
		if err != nil {
			s.logger.Error("❌ failed to soft delete child collection",
				zap.String("parentID", id),
				zap.String("childID", child.ID.String()),
				zap.Error(err))
			return err
		}
	}

	// Soft delete the collection itself
	return s.SoftDelete(ctx, id)
}

// Archive marks a collection as archived
func (s *softDeleteService) Archive(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		s.logger.Error("❌ collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Convert ID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		s.logger.Error("❌ invalid collection ID format", zap.String("id", id), zap.Error(err))
		return errors.NewAppError("invalid collection ID format", err)
	}

	// Get the collection to validate it exists and check current state
	existingCollection, err := s.getUseCase.Execute(ctx, objectID)
	if err != nil {
		s.logger.Error("❌ failed to get collection for archive", zap.String("id", id), zap.Error(err))
		return err
	}

	// Check if state transition is valid
	if err := collection.IsValidStateTransition(existingCollection.State, collection.CollectionStateArchived); err != nil {
		s.logger.Error("⚠️ invalid state transition for archive",
			zap.String("id", id),
			zap.String("currentState", existingCollection.State),
			zap.Error(err))
		return errors.NewAppError("cannot archive collection in current state", err)
	}

	// Update collection state to archived
	newState := collection.CollectionStateArchived
	updateInput := uc.UpdateCollectionInput{
		ID:    objectID,
		State: &newState,
	}

	_, err = s.updateUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ failed to archive collection", zap.String("id", id), zap.Error(err))
		return err
	}

	s.logger.Info("✅ collection archived successfully",
		zap.String("id", id),
		zap.String("previousState", existingCollection.State),
		zap.String("newState", collection.CollectionStateArchived))

	return nil
}

// Restore marks a deleted or archived collection as active
func (s *softDeleteService) Restore(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		s.logger.Error("❌ collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Convert ID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		s.logger.Error("❌ invalid collection ID format", zap.String("id", id), zap.Error(err))
		return errors.NewAppError("invalid collection ID format", err)
	}

	// Get the collection to validate it exists and check current state
	existingCollection, err := s.getUseCase.Execute(ctx, objectID)
	if err != nil {
		s.logger.Error("❌ failed to get collection for restore", zap.String("id", id), zap.Error(err))
		return err
	}

	// Check if state transition is valid
	if err := collection.IsValidStateTransition(existingCollection.State, collection.CollectionStateActive); err != nil {
		s.logger.Error("⚠️ invalid state transition for restore",
			zap.String("id", id),
			zap.String("currentState", existingCollection.State),
			zap.Error(err))
		return errors.NewAppError("cannot restore collection from current state", err)
	}

	// Update collection state to active
	newState := collection.CollectionStateActive
	updateInput := uc.UpdateCollectionInput{
		ID:    objectID,
		State: &newState,
	}

	_, err = s.updateUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ failed to restore collection", zap.String("id", id), zap.Error(err))
		return err
	}

	s.logger.Info("✅ collection restored successfully",
		zap.String("id", id),
		zap.String("previousState", existingCollection.State),
		zap.String("newState", collection.CollectionStateActive))

	return nil
}
