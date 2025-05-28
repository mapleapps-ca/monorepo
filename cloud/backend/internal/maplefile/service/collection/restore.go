// cloud/backend/internal/maplefile/service/collection/restore.go
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type RestoreCollectionRequestDTO struct {
	ID primitive.ObjectID `json:"id"`
}

type RestoreCollectionResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RestoreCollectionService interface {
	Execute(ctx context.Context, req *RestoreCollectionRequestDTO) (*RestoreCollectionResponseDTO, error)
}

type restoreCollectionServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewRestoreCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) RestoreCollectionService {
	logger = logger.Named("RestoreCollectionService")
	return &restoreCollectionServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *restoreCollectionServiceImpl) Execute(ctx context.Context, req *RestoreCollectionRequestDTO) (*RestoreCollectionResponseDTO, error) {
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
	// STEP 3: Retrieve existing collection (including non-active states for restoration)
	//
	collection, err := svc.repo.GetWithAnyState(ctx, req.ID)
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
	// STEP 4: Check if user has rights to restore this collection
	//
	if collection.OwnerID != userID {
		svc.logger.Warn("Unauthorized collection restore attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.ID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "Only the collection owner can restore a collection")
	}

	//
	// STEP 5: Validate state transition
	//
	err = dom_collection.IsValidStateTransition(collection.State, dom_collection.CollectionStateActive)
	if err != nil {
		svc.logger.Warn("Invalid state transition for collection restore",
			zap.Any("collection_id", req.ID),
			zap.String("current_state", collection.State),
			zap.String("target_state", dom_collection.CollectionStateActive),
			zap.Error(err))
		return nil, httperror.NewForBadRequestWithSingleField("state", err.Error())
	}

	//
	// STEP 6: Restore the collection
	//
	collection.State = dom_collection.CollectionStateActive
	collection.Version++ // Update mutation means we increment version.
	collection.ModifiedAt = time.Now()
	collection.ModifiedByUserID = userID
	err = svc.repo.Update(ctx, collection)
	if err != nil {
		svc.logger.Error("Failed to restore collection",
			zap.Any("error", err),
			zap.Any("collection_id", req.ID))
		return nil, err
	}

	svc.logger.Info("Collection restored successfully",
		zap.Any("collection_id", req.ID),
		zap.Any("user_id", userID))

	return &RestoreCollectionResponseDTO{
		Success: true,
		Message: "Collection restored successfully",
	}, nil
}
