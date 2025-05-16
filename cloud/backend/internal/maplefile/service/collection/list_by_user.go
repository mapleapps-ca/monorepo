// cloud/backend/internal/maplefile/service/collection/list_by_user.go
package collection

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

type CollectionsResponseDTO struct {
	Collections []*CollectionResponseDTO `json:"collections"`
}

type ListUserCollectionsService interface {
	Execute(sessCtx context.Context) (*CollectionsResponseDTO, error)
}

type listUserCollectionsServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewListUserCollectionsService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) ListUserCollectionsService {
	return &listUserCollectionsServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *listUserCollectionsServiceImpl) Execute(sessCtx context.Context) (*CollectionsResponseDTO, error) {
	//
	// STEP 1: Get user ID from context
	//
	userID, ok := sessCtx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, errors.New("user ID not found in context")
	}

	//
	// STEP 2: Get user's collections from repository
	//
	collections, err := svc.repo.GetAllByUserID(userID.Hex())
	if err != nil {
		svc.logger.Error("Failed to get user collections",
			zap.Any("error", err),
			zap.String("user_id", userID.Hex()))
		return nil, err
	}

	//
	// STEP 3: Map domain models to response DTOs
	//
	response := &CollectionsResponseDTO{
		Collections: make([]*CollectionResponseDTO, len(collections)),
	}

	for i, collection := range collections {
		collectionDTO := &CollectionResponseDTO{
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

		// Map members
		for j, member := range collection.Members {
			collectionDTO.Members[j] = struct {
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

		response.Collections[i] = collectionDTO
	}

	svc.logger.Debug("Retrieved user collections",
		zap.Int("count", len(collections)),
		zap.String("user_id", userID.Hex()))

	return response, nil
}
