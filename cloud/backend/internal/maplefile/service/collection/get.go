// cloud/backend/internal/maplefile/service/collection/get.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetCollectionService interface {
	Execute(ctx context.Context, collectionID primitive.ObjectID) (*CollectionResponseDTO, error)
}

type getCollectionServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewGetCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) GetCollectionService {
	return &getCollectionServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *getCollectionServiceImpl) Execute(ctx context.Context, collectionID primitive.ObjectID) (*CollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if collectionID.IsZero() {
		svc.logger.Warn("Empty collection ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("collection_id", "Collection ID is required")
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
	// STEP 3: Get collection from repository
	//
	collection, err := svc.repo.Get(ctx, collectionID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.Any("collection_id", collectionID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.Any("collection_id", collectionID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	//
	// STEP 4: Check if the user has access to this collection
	//
	// First check if user is owner
	hasAccess := collection.OwnerID == userID

	// If not owner, check if user is a member
	if !hasAccess {
		for _, member := range collection.Members {
			if member.RecipientID == userID {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized collection access attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", collectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have access to this collection")
	}

	//
	// STEP 5: Map domain model to response DTO
	//
	response := mapCollectionToDTO(collection)

	return response, nil
}
