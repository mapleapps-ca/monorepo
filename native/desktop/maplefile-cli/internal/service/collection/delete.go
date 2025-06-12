// monorepo/native/desktop/maplefile-cli/internal/service/collection/delete.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_tx "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// DeleteFromCloudOutput represents the result of deleting a collection from cloud
type DeleteFromCloudOutput struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	DeletedFromCloud bool   `json:"deleted_from_cloud"`
	DeletedFromLocal bool   `json:"deleted_from_local"`
	ChildrenDeleted  int    `json:"children_deleted"`
}

// DeleteService defines the interface for deleting collections from cloud
type DeleteService interface {
	Delete(ctx context.Context, id gocql.UUID) error
}

// deleteService implements the DeleteService interface
type deleteService struct {
	logger                                   *zap.Logger
	transactionManager                       dom_tx.Manager
	getUserByIsLoggedInUseCase               uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase                     uc_collection.GetCollectionUseCase
	softSoftDeleteCollectionFromCloudUseCase uc_collectiondto.SoftDeleteCollectionFromCloudUseCase
	deleteCollectionUseCase                  uc_collection.DeleteCollectionUseCase
	listCollectionsUseCase                   uc_collection.ListCollectionsUseCase
}

// NewDeleteService creates a new service for deleting collections from cloud
func NewDeleteService(
	logger *zap.Logger,
	transactionManager dom_tx.Manager,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	softSoftDeleteCollectionFromCloudUseCase uc_collectiondto.SoftDeleteCollectionFromCloudUseCase,
	deleteCollectionUseCase uc_collection.DeleteCollectionUseCase,
	listCollectionsUseCase uc_collection.ListCollectionsUseCase,
) DeleteService {
	logger = logger.Named("DeleteService")
	return &deleteService{
		logger:                                   logger,
		transactionManager:                       transactionManager,
		getUserByIsLoggedInUseCase:               getUserByIsLoggedInUseCase,
		getCollectionUseCase:                     getCollectionUseCase,
		softSoftDeleteCollectionFromCloudUseCase: softSoftDeleteCollectionFromCloudUseCase,
		deleteCollectionUseCase:                  deleteCollectionUseCase,
		listCollectionsUseCase:                   listCollectionsUseCase,
	}
}

// DeleteFromCloud deletes a collection from cloud and from local storage
func (s *deleteService) Delete(ctx context.Context, id gocql.UUID) error {
	//
	// STEP 1: Validate inputs
	//

	if id.String() == "" {
		s.logger.Error("Collection ID is required")
		return errors.NewAppError("Collection ID is required", nil)
	}

	//
	// STEP 2: Get user data for validation
	//

	userData, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("❌ failed to get authenticated user", zap.Error(err))
		return errors.NewAppError("failed to get user data", err)
	}

	if userData == nil {
		s.logger.Error("❌ authenticated user not found")
		return errors.NewAppError("authenticated user not found; please login first", nil)
	}

	//
	// STEP 3: Get collection to validate ownership/permissions
	//

	collection, err := s.getCollectionUseCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("❌ failed to get collection for deletion",
			zap.String("collectionID", id.String()),
			zap.Error(err))
		return err
	}

	if collection == nil {
		s.logger.Warn("⚠️ collection not found for deletion",
			zap.String("collectionID", id.String()))
		return errors.NewAppError("collection not found", nil)
	}

	// Validate user has permission to delete (owner or admin)
	canDelete := collection.OwnerID == userData.ID
	if !canDelete {
		// Check if user has admin permissions through membership
		for _, member := range collection.Members {
			if member.RecipientID == userData.ID && member.PermissionLevel == "admin" {
				canDelete = true
				break
			}
		}
	}

	if !canDelete {
		s.logger.Warn("⚠️ user does not have permission to delete collection",
			zap.String("userID", userData.ID.String()),
			zap.String("collectionID", id.String()))
		return errors.NewAppError("you don't have permission to delete this collection", nil)
	}

	//
	// STEP 4: Begin transaction for coordinated deletion
	//

	if err := s.transactionManager.Begin(); err != nil {
		s.logger.Error("❌ failed to begin transaction", zap.Error(err))
		return errors.NewAppError("failed to begin transaction", err)
	}

	output := &DeleteFromCloudOutput{
		Success:          false,
		DeletedFromCloud: false,
		DeletedFromLocal: false,
		ChildrenDeleted:  0,
	}

	//
	// STEP 5: Handle child collections
	//

	childCollections, err := s.listCollectionsUseCase.ListByParent(ctx, id)
	if err != nil {
		s.transactionManager.Rollback()
		s.logger.Error("❌ failed to list child collections",
			zap.String("collectionID", id.String()),
			zap.Error(err))
		return errors.NewAppError("failed to list child collections", err)
	}

	// Recursively delete child collections
	for _, child := range childCollections {
		err := s.deleteFromCloudRecursive(ctx, child.ID)
		if err != nil {
			s.transactionManager.Rollback()
			s.logger.Error("❌ failed to delete child collection",
				zap.String("parentID", id.String()),
				zap.String("childID", child.ID.String()),
				zap.Error(err))
			return err
		}
	}

	//
	// STEP 6: Delete from cloud
	//

	s.logger.Debug("🗑️ Deleting collection from cloud",
		zap.String("collectionID", id.String()))

	err = s.softSoftDeleteCollectionFromCloudUseCase.Execute(ctx, id)
	if err != nil {
		s.transactionManager.Rollback()
		s.logger.Error("❌ failed to delete collection from cloud",
			zap.String("collectionID", id.String()),
			zap.Error(err))
		return errors.NewAppError("failed to delete collection from cloud", err)
	}

	output.DeletedFromCloud = true
	s.logger.Info("✅ successfully deleted collection from cloud",
		zap.String("collectionID", id.String()))

	//
	// STEP 7: Delete from local storage if requested
	//

	s.logger.Debug("🗑️ Deleting collection from local storage",
		zap.String("collectionID", id.String()))

	err = s.deleteCollectionUseCase.Execute(ctx, id)
	if err != nil {
		// Don't rollback cloud deletion for local storage failure
		s.logger.Warn("⚠️ failed to delete collection from local storage (cloud deletion succeeded)",
			zap.String("collectionID", id.String()),
			zap.Error(err))
	} else {
		output.DeletedFromLocal = true
		s.logger.Info("✅ successfully deleted collection from local storage",
			zap.String("collectionID", id.String()))
	}

	//
	// STEP 8: Commit transaction and return result
	//

	if err := s.transactionManager.Commit(); err != nil {
		s.logger.Error("❌ failed to commit transaction", zap.Error(err))
		s.transactionManager.Rollback()
		return errors.NewAppError("failed to commit transaction", err)
	}

	output.Success = true
	output.Message = "Collection and children deleted successfully"

	s.logger.Info("✅ successfully completed collection deletion",
		zap.String("collectionID", id.String()),
		zap.Bool("deletedFromCloud", output.DeletedFromCloud),
		zap.Bool("deletedFromLocal", output.DeletedFromLocal),
		zap.Int("childrenDeleted", output.ChildrenDeleted))

	return nil
}

// deleteFromCloudRecursive is a helper method for recursive deletion without transaction management
func (s *deleteService) deleteFromCloudRecursive(ctx context.Context, id gocql.UUID) error {
	output := &DeleteFromCloudOutput{
		Success:          false,
		DeletedFromCloud: false,
		DeletedFromLocal: false,
		ChildrenDeleted:  0,
	}

	// Handle child collections if requested
	childCollections, err := s.listCollectionsUseCase.ListByParent(ctx, id)
	if err != nil {
		return err
	}

	// Recursively delete child collections
	for _, child := range childCollections {
		err := s.deleteFromCloudRecursive(ctx, child.ID)
		if err != nil {
			return err
		}
	}

	// Delete from cloud
	err = s.softSoftDeleteCollectionFromCloudUseCase.Execute(ctx, id)
	if err != nil {
		return err
	}

	// Delete from local storage if requested
	err = s.deleteCollectionUseCase.Execute(ctx, id)
	if err != nil {
		// Log warning but don't fail for local deletion issues
		s.logger.Warn("⚠️ failed to delete collection from local storage during recursive deletion",
			zap.String("collectionID", id.String()),
			zap.Error(err))
	} else {
		output.DeletedFromLocal = true
	}
	return nil
}
