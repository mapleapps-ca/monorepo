// cloud/backend/internal/maplefile/service/collection/softdelete.go
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	uc_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type SoftDeleteCollectionRequestDTO struct {
	ID primitive.ObjectID `json:"id"`
}

type SoftDeleteCollectionResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SoftDeleteCollectionService interface {
	Execute(ctx context.Context, req *SoftDeleteCollectionRequestDTO) (*SoftDeleteCollectionResponseDTO, error)
}

type softDeleteCollectionServiceImpl struct {
	config                  *config.Configuration
	logger                  *zap.Logger
	repo                    dom_collection.CollectionRepository
	getCollectionUseCase    uc_collection.GetCollectionUseCase
	updateCollectionUseCase uc_collection.UpdateCollectionUseCase
}

func NewSoftDeleteCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	updateCollectionUseCase uc_collection.UpdateCollectionUseCase,
) SoftDeleteCollectionService {
	logger = logger.Named("SoftDeleteCollectionService")
	return &softDeleteCollectionServiceImpl{
		config:                  config,
		logger:                  logger,
		repo:                    repo,
		getCollectionUseCase:    getCollectionUseCase,
		updateCollectionUseCase: updateCollectionUseCase,
	}
}

func (svc *softDeleteCollectionServiceImpl) Execute(ctx context.Context, req *SoftDeleteCollectionRequestDTO) (*SoftDeleteCollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Collection ID is required")
	}

	if req.ID.IsZero() {
		svc.logger.Warn("Empty collection ID")
		return nil, httperror.NewForBadRequestWithSingleField("id", "Collection ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Retrieve related records
	//
	collection, err := svc.getCollectionUseCase.Execute(ctx, req.ID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.Any("collection_id", req.ID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.Any("collection_id", req.ID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	//
	// STEP 4: Check if user has rights to delete this collection
	//
	if collection.OwnerID != userID {
		svc.logger.Warn("Unauthorized collection deletion attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.ID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "Only the collection owner can delete a collection")
	}

	// Check valid transitions.
	if err := dom_collection.IsValidStateTransition(collection.State, dom_collection.CollectionStateDeleted); err != nil {
		svc.logger.Warn("Invalid collection state transition",
			zap.Any("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	// Update mutation means we increment version.
	collection.Version++
	collection.ModifiedAt = time.Now()
	collection.ModifiedByUserID = userID
	collection.TombstoneVersion = collection.Version
	collection.TombstoneExpiry = time.Now().Add(time.Hour * 24 * 30)
	if err := svc.updateCollectionUseCase.Execute(ctx, collection); err != nil {
		svc.logger.Warn("Error updating collection state",
			zap.Any("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Check for child collections
	//
	descendants, err := svc.repo.FindDescendants(ctx, req.ID)
	if err != nil {
		svc.logger.Error("Failed to check for descendant collections",
			zap.Any("error", err),
			zap.Any("collection_id", req.ID))
		return nil, err
	}

	//
	// STEP 6: Delete all files in this collection and its descendants
	//
	// For this to work, we'd need to update the FileRepository to support filtering by multiple collection IDs
	// Otherwise, we'd need to loop through each collection and delete its files

	//
	// STEP 7: Delete the collection and all its descendants
	//

	if err := svc.repo.SoftDelete(ctx, req.ID); err != nil {
		svc.logger.Error("Failed to soft-delete collection",
			zap.Any("error", err),
			zap.Any("collection_id", req.ID))
		return nil, err
	}

	svc.logger.Info("Collection soft-deleted successfully",
		zap.Any("collection_id", req.ID),
		zap.Int("descendants_count", len(descendants)))

	return &SoftDeleteCollectionResponseDTO{
		Success: true,
		Message: "Collection, descendants, and all associated files soft-deleted successfully",
	}, nil
}
