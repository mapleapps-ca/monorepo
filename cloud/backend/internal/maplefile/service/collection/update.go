// cloud/backend/internal/maplefile/service/collection/update.go
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/keys"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type UpdateCollectionRequestDTO struct {
	ID                     primitive.ObjectID           `json:"id"`
	EncryptedName          string                       `json:"encrypted_name"`
	CollectionType         string                       `json:"collection_type,omitempty"`
	EncryptedCollectionKey *keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
}

type UpdateCollectionService interface {
	Execute(ctx context.Context, req *UpdateCollectionRequestDTO) (*CollectionResponseDTO, error)
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

func (svc *updateCollectionServiceImpl) Execute(ctx context.Context, req *UpdateCollectionRequestDTO) (*CollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Collection details are required")
	}

	e := make(map[string]string)
	if req.ID.IsZero() {
		e["id"] = "Collection ID is required"
	}
	if req.EncryptedName == "" {
		e["encrypted_name"] = "Collection name is required"
	}
	if req.CollectionType != "" && req.CollectionType != dom_collection.CollectionTypeFolder && req.CollectionType != dom_collection.CollectionTypeAlbum {
		e["collection_type"] = "Collection type must be either 'folder' or 'album'"
	}

	if len(e) != 0 {
		svc.logger.Warn("Failed validation",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
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
	// STEP 3: Retrieve existing collection
	//
	collection, err := svc.repo.Get(ctx, req.ID)
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
	// STEP 4: Check if user has rights to update this collection
	//
	if collection.OwnerID != userID {
		// Check if user is a member with admin permissions
		isAdmin := false
		for _, member := range collection.Members {
			if member.RecipientID == userID && member.PermissionLevel == dom_collection.CollectionPermissionAdmin {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			svc.logger.Warn("Unauthorized collection update attempt",
				zap.Any("user_id", userID),
				zap.Any("collection_id", req.ID))
			return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to update this collection")
		}
	}

	//
	// STEP 5: Update collection
	//
	collection.EncryptedName = req.EncryptedName
	collection.ModifiedAt = time.Now()

	// Only update optional fields if they are provided
	if req.CollectionType != "" {
		collection.CollectionType = req.CollectionType
	}
	if req.EncryptedCollectionKey.Ciphertext != nil && len(req.EncryptedCollectionKey.Ciphertext) > 0 &&
		req.EncryptedCollectionKey.Nonce != nil && len(req.EncryptedCollectionKey.Nonce) > 0 {
		collection.EncryptedCollectionKey = req.EncryptedCollectionKey
	}

	//
	// STEP 6: Save updated collection
	//
	err = svc.repo.Update(ctx, collection)
	if err != nil {
		svc.logger.Error("Failed to update collection",
			zap.Any("error", err),
			zap.Any("collection_id", collection.ID))
		return nil, err
	}

	//
	// STEP 7: Map domain model to response DTO
	//
	response := mapCollectionToDTO(collection)

	svc.logger.Debug("Collection updated successfully",
		zap.Any("collection_id", collection.ID))

	return response, nil
}
