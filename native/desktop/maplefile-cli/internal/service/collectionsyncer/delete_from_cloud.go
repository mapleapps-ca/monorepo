// monorepo/native/desktop/maplefile-cli/internal/service/collectionsyncer/delete_from_cloud.go
package collectionsyncer

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

// DeleteFromCloudInput represents the input for deleting a collection from cloud
type DeleteFromCloudInput struct {
	ID                 gocql.UUID `json:"id"`
	DeleteLocal        bool       `json:"delete_local"`         // Whether to also delete the local copy
	DeleteWithChildren bool       `json:"delete_with_children"` // Whether to delete child collections
}

// DeleteFromCloudOutput represents the result of deleting a collection from cloud
type DeleteFromCloudOutput struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	DeletedFromCloud bool   `json:"deleted_from_cloud"`
	DeletedFromLocal bool   `json:"deleted_from_local"`
	ChildrenDeleted  int    `json:"children_deleted"`
}

// DeleteFromCloudService defines the interface for deleting collections from cloud
type DeleteFromCloudService interface {
	DeleteFromCloud(ctx context.Context, input *DeleteFromCloudInput) (*DeleteFromCloudOutput, error)
}

// deleteFromCloudService implements the DeleteFromCloudService interface
type deleteFromCloudService struct {
	logger                                   *zap.Logger
	transactionManager                       dom_tx.Manager
	getUserByIsLoggedInUseCase               uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase                     uc_collection.GetCollectionUseCase
	softSoftDeleteCollectionFromCloudUseCase uc_collectiondto.SoftDeleteCollectionFromCloudUseCase
	deleteCollectionUseCase                  uc_collection.DeleteCollectionUseCase
	listCollectionsUseCase                   uc_collection.ListCollectionsUseCase
}

// NewDeleteFromCloudService creates a new service for deleting collections from cloud
func NewDeleteFromCloudService(
	logger *zap.Logger,
	transactionManager dom_tx.Manager,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	softSoftDeleteCollectionFromCloudUseCase uc_collectiondto.SoftDeleteCollectionFromCloudUseCase,
	deleteCollectionUseCase uc_collection.DeleteCollectionUseCase,
	listCollectionsUseCase uc_collection.ListCollectionsUseCase,
) DeleteFromCloudService {
	logger = logger.Named("DeleteFromCloudService")
	return &deleteFromCloudService{
		logger:                                   logger,
		transactionManager:                       transactionManager,
		getUserByIsLoggedInUseCase:               getUserByIsLoggedInUseCase,
		getCollectionUseCase:                     getCollectionUseCase,
		softSoftDeleteCollectionFromCloudUseCase: softSoftDeleteCollectionFromCloudUseCase,
		deleteCollectionUseCase:                  deleteCollectionUseCase,
		listCollectionsUseCase:                   listCollectionsUseCase,
	}
}

// DeleteFromCloud deletes a collection from cloud and optionally from local storage
func (s *deleteFromCloudService) DeleteFromCloud(ctx context.Context, input *DeleteFromCloudInput) (*DeleteFromCloudOutput, error) {
	//
	// STEP 1: Validate inputs
	//

	if input == nil {
		s.logger.Error("❌ input is required")
		return nil, errors.NewAppError("input is required", nil)
	}

	if input.ID.String() == "" {
		s.logger.Error("❌ collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	//
	// STEP 2: Get user data for validation
	//

	userData, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("❌ failed to get authenticated user", zap.Error(err))
		return nil, errors.NewAppError("failed to get user data", err)
	}

	if userData == nil {
		s.logger.Error("❌ authenticated user not found")
		return nil, errors.NewAppError("authenticated user not found; please login first", nil)
	}

	//
	// STEP 3: Get collection to validate ownership/permissions
	//

	collection, err := s.getCollectionUseCase.Execute(ctx, input.ID)
	if err != nil {
		s.logger.Error("❌ failed to get collection for deletion",
			zap.String("collectionID", input.ID.String()),
			zap.Error(err))
		return nil, err
	}

	if collection == nil {
		s.logger.Warn("⚠️ collection not found for deletion",
			zap.String("collectionID", input.ID.String()))
		return nil, errors.NewAppError("collection not found", nil)
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
			zap.String("collectionID", input.ID.String()))
		return nil, errors.NewAppError("you don't have permission to delete this collection", nil)
	}

	//
	// STEP 4: Begin transaction for coordinated deletion
	//

	if err := s.transactionManager.Begin(); err != nil {
		s.logger.Error("❌ failed to begin transaction", zap.Error(err))
		return nil, errors.NewAppError("failed to begin transaction", err)
	}

	output := &DeleteFromCloudOutput{
		Success:          false,
		DeletedFromCloud: false,
		DeletedFromLocal: false,
		ChildrenDeleted:  0,
	}

	//
	// STEP 5: Handle child collections if requested
	//

	if input.DeleteWithChildren {
		childCollections, err := s.listCollectionsUseCase.ListByParent(ctx, input.ID)
		if err != nil {
			s.transactionManager.Rollback()
			s.logger.Error("❌ failed to list child collections",
				zap.String("collectionID", input.ID.String()),
				zap.Error(err))
			return nil, errors.NewAppError("failed to list child collections", err)
		}

		// Recursively delete child collections
		for _, child := range childCollections {
			childInput := &DeleteFromCloudInput{
				ID:                 child.ID,
				DeleteLocal:        input.DeleteLocal,
				DeleteWithChildren: true, // Recursive deletion
			}

			childOutput, err := s.deleteFromCloudRecursive(ctx, childInput)
			if err != nil {
				s.transactionManager.Rollback()
				s.logger.Error("❌ failed to delete child collection",
					zap.String("parentID", input.ID.String()),
					zap.String("childID", child.ID.String()),
					zap.Error(err))
				return nil, err
			}

			if childOutput.Success {
				output.ChildrenDeleted++
			}
		}
	}

	//
	// STEP 6: Delete from cloud
	//

	s.logger.Debug("🗑️ Deleting collection from cloud",
		zap.String("collectionID", input.ID.String()))

	err = s.softSoftDeleteCollectionFromCloudUseCase.Execute(ctx, input.ID)
	if err != nil {
		s.transactionManager.Rollback()
		s.logger.Error("❌ failed to delete collection from cloud",
			zap.String("collectionID", input.ID.String()),
			zap.Error(err))
		return nil, errors.NewAppError("failed to delete collection from cloud", err)
	}

	output.DeletedFromCloud = true
	s.logger.Info("✅ successfully deleted collection from cloud",
		zap.String("collectionID", input.ID.String()))

	//
	// STEP 7: Delete from local storage if requested
	//

	if input.DeleteLocal {
		s.logger.Debug("🗑️ Deleting collection from local storage",
			zap.String("collectionID", input.ID.String()))

		err = s.deleteCollectionUseCase.Execute(ctx, input.ID)
		if err != nil {
			// Don't rollback cloud deletion for local storage failure
			s.logger.Warn("⚠️ failed to delete collection from local storage (cloud deletion succeeded)",
				zap.String("collectionID", input.ID.String()),
				zap.Error(err))
		} else {
			output.DeletedFromLocal = true
			s.logger.Info("✅ successfully deleted collection from local storage",
				zap.String("collectionID", input.ID.String()))
		}
	}

	//
	// STEP 8: Commit transaction and return result
	//

	if err := s.transactionManager.Commit(); err != nil {
		s.logger.Error("❌ failed to commit transaction", zap.Error(err))
		s.transactionManager.Rollback()
		return nil, errors.NewAppError("failed to commit transaction", err)
	}

	output.Success = true
	if input.DeleteWithChildren && output.ChildrenDeleted > 0 {
		output.Message = "Collection and children deleted successfully"
	} else {
		output.Message = "Collection deleted successfully"
	}

	s.logger.Info("✅ successfully completed collection deletion",
		zap.String("collectionID", input.ID.String()),
		zap.Bool("deletedFromCloud", output.DeletedFromCloud),
		zap.Bool("deletedFromLocal", output.DeletedFromLocal),
		zap.Int("childrenDeleted", output.ChildrenDeleted))

	return output, nil
}

// deleteFromCloudRecursive is a helper method for recursive deletion without transaction management
func (s *deleteFromCloudService) deleteFromCloudRecursive(ctx context.Context, input *DeleteFromCloudInput) (*DeleteFromCloudOutput, error) {
	output := &DeleteFromCloudOutput{
		Success:          false,
		DeletedFromCloud: false,
		DeletedFromLocal: false,
		ChildrenDeleted:  0,
	}

	// Handle child collections if requested
	if input.DeleteWithChildren {
		childCollections, err := s.listCollectionsUseCase.ListByParent(ctx, input.ID)
		if err != nil {
			return nil, err
		}

		// Recursively delete child collections
		for _, child := range childCollections {
			childInput := &DeleteFromCloudInput{
				ID:                 child.ID,
				DeleteLocal:        input.DeleteLocal,
				DeleteWithChildren: true,
			}

			childOutput, err := s.deleteFromCloudRecursive(ctx, childInput)
			if err != nil {
				return nil, err
			}

			if childOutput.Success {
				output.ChildrenDeleted++
			}
		}
	}

	// Delete from cloud
	err := s.softSoftDeleteCollectionFromCloudUseCase.Execute(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	output.DeletedFromCloud = true

	// Delete from local storage if requested
	if input.DeleteLocal {
		err = s.deleteCollectionUseCase.Execute(ctx, input.ID)
		if err != nil {
			// Log warning but don't fail for local deletion issues
			s.logger.Warn("⚠️ failed to delete collection from local storage during recursive deletion",
				zap.String("collectionID", input.ID.String()),
				zap.Error(err))
		} else {
			output.DeletedFromLocal = true
		}
	}

	output.Success = true
	return output, nil
}
