// cloud/backend/internal/papercloud/service/collection/get.go
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetCollectionService interface {
	Execute(sessCtx context.Context, collectionID string) (*CollectionResponseDTO, error)
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

func (svc *getCollectionServiceImpl) Execute(sessCtx context.Context, collectionID string) (*CollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if collectionID == "" {
		svc.logger.Warn("Empty collection ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("collection_id", "Collection ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := sessCtx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Get collection from repository
	//
	collection, err := svc.repo.Get(collectionID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.String("collection_id", collectionID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.String("collection_id", collectionID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	//
	// STEP 4: Check if the user has access to this collection
	//
	// First check if user is owner
	hasAccess := collection.OwnerID == userID.Hex()

	// If not owner, check if user is a member
	if !hasAccess {
		for _, member := range collection.Members {
			if member.RecipientID == userID.Hex() {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized collection access attempt",
			zap.String("user_id", userID.Hex()),
			zap.String("collection_id", collectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have access to this collection")
	}

	//
	// STEP 5: Map domain model to response DTO
	//
	response := &CollectionResponseDTO{
		ID:                     collection.ID,
		OwnerID:                collection.OwnerID,
		Name:                   collection.Name,
		Path:                   collection.Path,
		Type:                   collection.Type,
		EncryptedCollectionKey: collection.EncryptedCollectionKey,
		CreatedAt:              collection.CreatedAt,
		UpdatedAt:              collection.UpdatedAt,
		Members: make([]struct {
			RecipientID     string    `json:"recipient_id"`
			RecipientEmail  string    `json:"recipient_email"`
			PermissionLevel string    `json:"permission_level"`
			GrantedByID     string    `json:"granted_by_id"`
			CollectionID    string    `json:"collection_id"`
			CreatedAt       time.Time `json:"created_at"`
		}, len(collection.Members)),
	}

	// Map members from domain model to response DTO
	for i, member := range collection.Members {
		response.Members[i] = struct {
			RecipientID     string    `json:"recipient_id"`
			RecipientEmail  string    `json:"recipient_email"`
			PermissionLevel string    `json:"permission_level"`
			GrantedByID     string    `json:"granted_by_id"`
			CollectionID    string    `json:"collection_id"`
			CreatedAt       time.Time `json:"created_at"`
		}{
			RecipientID:     member.RecipientID,
			RecipientEmail:  member.RecipientEmail,
			PermissionLevel: member.PermissionLevel,
			GrantedByID:     member.GrantedByID,
			CollectionID:    member.CollectionID,
			CreatedAt:       member.CreatedAt,
		}
	}

	return response, nil
}
