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
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	EncryptedPathSegments  []string                    `json:"encrypted_path_segments,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key"`
}

type CollectionResponseDTO struct {
	ID                     primitive.ObjectID          `json:"id"`
	OwnerID                primitive.ObjectID          `json:"owner_id"`
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	AncestorIDs            []primitive.ObjectID        `json:"ancestor_ids,omitempty"`
	EncryptedPathSegments  []string                    `json:"encrypted_path_segments,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
	Children               []*CollectionResponseDTO    `json:"children,omitempty"`
	CreatedAt              time.Time                   `json:"created_at"`
	ModifiedAt             time.Time                   `json:"modified_at"`
	Members                []MembershipResponseDTO     `json:"members"`
}

type MembershipResponseDTO struct {
	ID              primitive.ObjectID `json:"id"`
	RecipientID     primitive.ObjectID `json:"recipient_id"`
	RecipientEmail  string             `json:"recipient_email"`
	PermissionLevel string             `json:"permission_level"`
	GrantedByID     primitive.ObjectID `json:"granted_by_id"`
	CollectionID    primitive.ObjectID `json:"collection_id"`
	IsInherited     bool               `json:"is_inherited"`
	InheritedFromID primitive.ObjectID `json:"inherited_from_id,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
}

type CreateCollectionService interface {
	Execute(ctx context.Context, req *CreateCollectionRequestDTO) (*CollectionResponseDTO, error)
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

func (svc *createCollectionServiceImpl) Execute(ctx context.Context, req *CreateCollectionRequestDTO) (*CollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Collection details are required")
	}

	e := make(map[string]string)
	if req.EncryptedName == "" {
		e["encrypted_name"] = "Collection name is required"
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
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Create collection object
	//
	now := time.Now()
	collection := &dom_collection.Collection{
		ID:                     primitive.NewObjectID(),
		OwnerID:                userID,
		EncryptedName:          req.EncryptedName,
		Type:                   req.Type,
		CreatedAt:              now,
		ModifiedAt:             now,
		EncryptedCollectionKey: req.EncryptedCollectionKey,
		Members:                []dom_collection.CollectionMembership{}, // Initialize empty members slice
		EncryptedPathSegments:  req.EncryptedPathSegments,
	}

	// If parent ID is provided, set it
	if !req.ParentID.IsZero() {
		collection.ParentID = req.ParentID
	}

	//
	// STEP 4: Create collection in repository
	//
	err := svc.repo.Create(ctx, collection)
	if err != nil {
		svc.logger.Error("Failed to create collection",
			zap.Any("error", err),
			zap.Any("owner_id", collection.OwnerID),
			zap.String("name", collection.EncryptedName))
		return nil, err
	}

	//
	// STEP 5: Map domain model to response DTO
	//
	response := mapCollectionToDTO(collection)

	svc.logger.Debug("Collection created successfully",
		zap.Any("collection_id", collection.ID),
		zap.Any("owner_id", collection.OwnerID))

	return response, nil
}

// Helper function to map a Collection domain model to a CollectionResponseDTO
func mapCollectionToDTO(collection *dom_collection.Collection) *CollectionResponseDTO {
	responseDTO := &CollectionResponseDTO{
		ID:                     collection.ID,
		OwnerID:                collection.OwnerID,
		EncryptedName:          collection.EncryptedName,
		Type:                   collection.Type,
		ParentID:               collection.ParentID,
		AncestorIDs:            collection.AncestorIDs,
		EncryptedPathSegments:  collection.EncryptedPathSegments,
		EncryptedCollectionKey: collection.EncryptedCollectionKey,
		CreatedAt:              collection.CreatedAt,
		ModifiedAt:             collection.ModifiedAt,
		Members:                make([]MembershipResponseDTO, len(collection.Members)),
	}

	// Map members
	for i, member := range collection.Members {
		responseDTO.Members[i] = MembershipResponseDTO{
			ID:              member.ID,
			RecipientID:     member.RecipientID,
			RecipientEmail:  member.RecipientEmail,
			PermissionLevel: member.PermissionLevel,
			GrantedByID:     member.GrantedByID,
			CollectionID:    member.CollectionID,
			IsInherited:     member.IsInherited,
			InheritedFromID: member.InheritedFromID,
			CreatedAt:       member.CreatedAt,
		}
	}

	// Map children if present
	if len(collection.Children) > 0 {
		responseDTO.Children = make([]*CollectionResponseDTO, len(collection.Children))
		for i, child := range collection.Children {
			responseDTO.Children[i] = mapCollectionToDTO(child)
		}
	}

	return responseDTO
}
