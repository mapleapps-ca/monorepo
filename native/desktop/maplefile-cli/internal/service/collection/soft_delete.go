// internal/service/collection/soft_delete.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_tx "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
)

// SoftDeleteService defines the interface for soft deleting local collections
type SoftDeleteService interface {
	SoftDelete(ctx context.Context, id gocql.UUID) error
	SoftDeleteWithChildren(ctx context.Context, id gocql.UUID) error
	Archive(ctx context.Context, id gocql.UUID) error
	Restore(ctx context.Context, id gocql.UUID) error
}

// softDeleteService implements the SoftDeleteService interface
type softDeleteService struct {
	logger                                   *zap.Logger
	transactionManager                       dom_tx.Manager
	getUseCase                               uc.GetCollectionUseCase
	updateUseCase                            uc.UpdateCollectionUseCase
	listUseCase                              uc.ListCollectionsUseCase
	softSoftDeleteCollectionFromCloudUseCase uc_collectiondto.SoftDeleteCollectionFromCloudUseCase
}

// NewSoftDeleteService creates a new service for soft deleting local collections
func NewSoftDeleteService(
	logger *zap.Logger,
	transactionManager dom_tx.Manager,
	getUseCase uc.GetCollectionUseCase,
	updateUseCase uc.UpdateCollectionUseCase,
	listUseCase uc.ListCollectionsUseCase,
	softSoftDeleteCollectionFromCloudUseCase uc_collectiondto.SoftDeleteCollectionFromCloudUseCase,
) SoftDeleteService {
	logger = logger.Named("CollectionSoftDeleteService")
	return &softDeleteService{
		logger:                                   logger,
		transactionManager:                       transactionManager,
		getUseCase:                               getUseCase,
		updateUseCase:                            updateUseCase,
		listUseCase:                              listUseCase,
		softSoftDeleteCollectionFromCloudUseCase: softSoftDeleteCollectionFromCloudUseCase,
	}
}

// SoftDelete marks a collection as deleted by updating its state
func (s *softDeleteService) SoftDelete(ctx context.Context, id gocql.UUID) error {
	// Validate input
	if id.String() == "" {
		s.logger.Error("❌ collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Get the collection to validate it exists and check current state
	existingCollection, err := s.getUseCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("❌ failed to get collection for soft delete",
			zap.String("id", id.String()),
			zap.Error(err))
		return err
	}

	// Check if state transition is valid
	if err := collection.IsValidStateTransition(existingCollection.State, collection.CollectionStateDeleted); err != nil {
		s.logger.Error("⚠️ invalid state transition for soft delete",
			zap.String("id", id.String()),
			zap.String("currentState", existingCollection.State),
			zap.Error(err))
		return errors.NewAppError("cannot delete collection in current state", err)
	}

	// Begin transaction for coordinated deletion
	if err := s.transactionManager.Begin(); err != nil {
		s.logger.Error("❌ failed to begin transaction", zap.Error(err))
		return errors.NewAppError("failed to begin transaction", err)
	}

	// Update collection state to deleted
	newState := collection.CollectionStateDeleted
	updateInput := uc.UpdateCollectionInput{
		ID:    id,
		State: &newState,
	}

	_, err = s.updateUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ failed to soft delete collection",
			zap.String("id", id.String()),
			zap.Error(err))
		s.transactionManager.Rollback()
		return err
	}

	s.logger.Info("✅ collection (soft)deleted successfully",
		zap.String("id", id.String()),
		zap.String("previousState", existingCollection.State),
		zap.String("newState", collection.CollectionStateDeleted))

	// Delete from cloud
	err = s.softSoftDeleteCollectionFromCloudUseCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("❌ failed to (soft)delete collection from cloud",
			zap.String("collectionID", id.String()),
			zap.Error(err))
		s.transactionManager.Rollback()
		return errors.NewAppError("failed to delete collection from cloud", err)
	}

	// Commit transaction and return result
	if err := s.transactionManager.Commit(); err != nil {
		s.logger.Error("❌ failed to commit transaction", zap.Error(err))
		s.transactionManager.Rollback()
		return errors.NewAppError("failed to commit transaction", err)
	}

	return nil
}

// SoftDeleteWithChildren soft deletes a collection and all its children
func (s *softDeleteService) SoftDeleteWithChildren(ctx context.Context, id gocql.UUID) error {
	// Validate input
	if id.String() == "" {
		s.logger.Error("❌ collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Get all children of this collection
	children, err := s.listUseCase.ListByParent(ctx, id)
	if err != nil {
		return errors.NewAppError("failed to list child collections", err)
	}

	// Soft delete each child recursively
	for _, child := range children {
		err = s.SoftDeleteWithChildren(ctx, child.ID)
		if err != nil {
			s.logger.Error("❌ failed to soft delete child collection",
				zap.String("parentID", id.String()),
				zap.String("childID", child.ID.String()),
				zap.Error(err))
			return err
		}
	}

	// Soft delete the collection itself
	return s.SoftDelete(ctx, id)
}

// Archive marks a collection as archived
func (s *softDeleteService) Archive(ctx context.Context, id gocql.UUID) error {
	// Validate input
	if id.String() == "" {
		s.logger.Error("❌ collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Get the collection to validate it exists and check current state
	existingCollection, err := s.getUseCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("❌ failed to get collection for archive",
			zap.String("id", id.String()),
			zap.Error(err))
		return err
	}

	// Check if state transition is valid
	if err := collection.IsValidStateTransition(existingCollection.State, collection.CollectionStateArchived); err != nil {
		s.logger.Error("⚠️ invalid state transition for archive",
			zap.String("id", id.String()),
			zap.String("currentState", existingCollection.State),
			zap.Error(err))
		return errors.NewAppError("cannot archive collection in current state", err)
	}

	// Update collection state to archived
	newState := collection.CollectionStateArchived
	updateInput := uc.UpdateCollectionInput{
		ID:    id,
		State: &newState,
	}

	_, err = s.updateUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ failed to archive collection",
			zap.String("id", id.String()),
			zap.Error(err))
		return err
	}

	s.logger.Info("✅ collection archived successfully",
		zap.String("id", id.String()),
		zap.String("previousState", existingCollection.State),
		zap.String("newState", collection.CollectionStateArchived))

	return nil
}

// Restore marks a deleted or archived collection as active
func (s *softDeleteService) Restore(ctx context.Context, id gocql.UUID) error {
	// Validate input
	if id.String() == "" {
		s.logger.Error("❌ collection ID is required")
		return errors.NewAppError("collection ID is required", nil)
	}

	// Get the collection to validate it exists and check current state
	existingCollection, err := s.getUseCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("❌ failed to get collection for restore",
			zap.String("id", id.String()),
			zap.Error(err))
		return err
	}

	// Check if state transition is valid
	if err := collection.IsValidStateTransition(existingCollection.State, collection.CollectionStateActive); err != nil {
		s.logger.Error("⚠️ invalid state transition for restore",
			zap.String("id", id.String()),
			zap.String("currentState", existingCollection.State),
			zap.Error(err))
		return errors.NewAppError("cannot restore collection from current state", err)
	}

	// Update collection state to active
	newState := collection.CollectionStateActive
	updateInput := uc.UpdateCollectionInput{
		ID:    id,
		State: &newState,
	}

	_, err = s.updateUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ failed to restore collection",
			zap.String("id", id.String()),
			zap.Error(err))
		return err
	}

	s.logger.Info("✅ collection restored successfully",
		zap.String("id", id.String()),
		zap.String("previousState", existingCollection.State),
		zap.String("newState", collection.CollectionStateActive))

	return nil
}
