// cloud/backend/internal/maplefile/service/collection/create.go
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

type CreateCollectionRequestDTO struct {
	Name                   string                      `json:"name"`
	Type                   string                      `json:"type"`
	Path                   string                      `json:"path"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key"`
}

type CollectionResponseDTO struct {
	ID                     string                      `json:"id"`
	OwnerID                string                      `json:"owner_id"`
	Name                   string                      `json:"name"`
	Path                   string                      `json:"path"`
	Type                   string                      `json:"type"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
	CreatedAt              time.Time                   `json:"created_at"`
	UpdatedAt              time.Time                   `json:"updated_at"`
	Members                []struct {
		RecipientID     string    `json:"recipient_id"`
		RecipientEmail  string    `json:"recipient_email"`
		PermissionLevel string    `json:"permission_level"`
		GrantedByID     string    `json:"granted_by_id"`
		CollectionID    string    `json:"collection_id"`
		CreatedAt       time.Time `json:"created_at"`
	} `json:"members"`
}

type CreateCollectionService interface {
	Execute(sessCtx context.Context, req *CreateCollectionRequestDTO) (*CollectionResponseDTO, error)
}

type createCollectionServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewCreateCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) CreateCollectionService {
	return &createCollectionServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *createCollectionServiceImpl) Execute(sessCtx context.Context, req *CreateCollectionRequestDTO) (*CollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Collection details are required")
	}

	e := make(map[string]string)
	if req.Name == "" {
		e["name"] = "Collection name is required"
	}
	if req.Type == "" {
		e["type"] = "Collection type is required"
	} else if req.Type != dom_collection.CollectionTypeFolder && req.Type != dom_collection.CollectionTypeAlbum {
		e["type"] = "Collection type must be either 'folder' or 'album'"
	}
	if req.EncryptedCollectionKey.Ciphertext == nil || len(req.EncryptedCollectionKey.Ciphertext) == 0 {
		e["encrypted_collection_key"] = "Encrypted collection key is required"
	}
	if req.EncryptedCollectionKey.Nonce == nil || len(req.EncryptedCollectionKey.Nonce) == 0 {
		e["encrypted_collection_key"] = "Encrypted collection key nonce is required"
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
	// STEP 3: Create collection object
	//
	now := time.Now()
	collection := &dom_collection.Collection{
		OwnerID:                userID.Hex(), // Convert ObjectID to string
		Name:                   req.Name,
		Path:                   req.Path,
		Type:                   req.Type,
		CreatedAt:              now,
		UpdatedAt:              now,
		EncryptedCollectionKey: req.EncryptedCollectionKey,
		Members:                []dom_collection.CollectionMembership{}, // Initialize empty members slice
	}

	//
	// STEP 4: Create collection in repository
	//
	err := svc.repo.Create(collection)
	if err != nil {
		svc.logger.Error("Failed to create collection",
			zap.Any("error", err),
			zap.String("owner_id", collection.OwnerID),
			zap.String("name", collection.Name))
		return nil, err
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
		}, 0),
	}

	svc.logger.Debug("Collection created successfully",
		zap.String("collection_id", collection.ID),
		zap.String("owner_id", collection.OwnerID))

	return response, nil
}
