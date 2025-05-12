// cloud/backend/internal/papercloud/service/collection/update.go
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/keys"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type UpdateCollectionRequestDTO struct {
	ID                     string                      `json:"id"`
	Name                   string                      `json:"name"`
	Path                   string                      `json:"path,omitempty"`
	Type                   string                      `json:"type,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key"`
}

type UpdateCollectionService interface {
	Execute(sessCtx context.Context, req *UpdateCollectionRequestDTO) (*CollectionResponseDTO, error)
}

type updateCollectionServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewUpdateCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) UpdateCollectionService {
	return &updateCollectionServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *updateCollectionServiceImpl) Execute(sessCtx context.Context, req *UpdateCollectionRequestDTO) (*CollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Collection details are required")
	}

	e := make(map[string]string)
	if req.ID == "" {
		e["id"] = "Collection ID is required"
	}
	if req.Name == "" {
		e["name"] = "Collection name is required"
	}
	if req.Type != "" && req.Type != dom_collection.CollectionTypeFolder && req.Type != dom_collection.CollectionTypeAlbum {
		e["type"] = "Collection type must be either 'folder' or 'album'"
	}

	if len(e) != 0 {
		svc.logger.Warn("Failed validation",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
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
	// STEP 3: Retrieve existing collection
	//
	collection, err := svc.repo.Get(req.ID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.String("collection_id", req.ID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.String("collection_id", req.ID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	//
	// STEP 4: Check if user has rights to update this collection
	//
	if collection.OwnerID != userID.Hex() {
		// Check if user is a member with admin permissions
		isAdmin := false
		for _, member := range collection.Members {
			if member.RecipientID == userID.Hex() && member.PermissionLevel == dom_collection.CollectionPermissionAdmin {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			svc.logger.Warn("Unauthorized collection update attempt",
				zap.String("user_id", userID.Hex()),
				zap.String("collection_id", req.ID))
			return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to update this collection")
		}
	}

	//
	// STEP 5: Update collection
	//
	collection.Name = req.Name
	collection.UpdatedAt = time.Now()

	// Only update optional fields if they are provided
	if req.Path != "" {
		collection.Path = req.Path
	}
	if req.Type != "" {
		collection.Type = req.Type
	}
	if req.EncryptedCollectionKey.Ciphertext != nil && len(req.EncryptedCollectionKey.Ciphertext) > 0 &&
		req.EncryptedCollectionKey.Nonce != nil && len(req.EncryptedCollectionKey.Nonce) > 0 {
		collection.EncryptedCollectionKey = req.EncryptedCollectionKey
	}

	//
	// STEP 6: Save updated collection
	//
	err = svc.repo.Update(collection)
	if err != nil {
		svc.logger.Error("Failed to update collection",
			zap.Any("error", err),
			zap.String("collection_id", collection.ID))
		return nil, err
	}

	//
	// STEP 7: Map domain model to response DTO
	//
	response := &CollectionResponseDTO{
		ID:        collection.ID,
		OwnerID:   collection.OwnerID,
		Name:      collection.Name,
		Path:      collection.Path,
		Type:      collection.Type,
		CreatedAt: collection.CreatedAt,
		UpdatedAt: collection.UpdatedAt,
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

	svc.logger.Debug("Collection updated successfully",
		zap.String("collection_id", collection.ID))

	return response, nil
}
