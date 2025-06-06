// cloud/backend/internal/maplefile/service/collection/create.go
package collection

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/keys"
	uc_federateduser "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/user"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// CreateCollectionRequestDTO represents a Data Transfer Object (DTO)
// used for transferring collection (folder or album) data between the local device and the cloud server.
// This data is end-to-end encrypted (E2EE) on the local device before transmission.
// The cloud server stores this encrypted data but cannot decrypt it.
// On the local device, this data is decrypted for use and storage (not stored in this encrypted DTO format locally).
// It can represent both root collections and embedded subcollections.
type CreateCollectionRequestDTO struct {
	ID                     gocql.UUID                    `bson:"_id" json:"id"`
	OwnerID                gocql.UUID                    `bson:"owner_id" json:"owner_id"`
	EncryptedName          string                        `bson:"encrypted_name" json:"encrypted_name"`
	CollectionType         string                        `bson:"collection_type" json:"collection_type"`
	EncryptedCollectionKey *keys.EncryptedCollectionKey  `bson:"encrypted_collection_key" json:"encrypted_collection_key"`
	Members                []*CollectionMembershipDTO    `bson:"members" json:"members"`
	ParentID               gocql.UUID                    `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
	AncestorIDs            []gocql.UUID                  `bson:"ancestor_ids,omitempty" json:"ancestor_ids,omitempty"`
	Children               []*CreateCollectionRequestDTO `bson:"children,omitempty" json:"children,omitempty"`
	CreatedAt              time.Time                     `bson:"created_at" json:"created_at"`
	CreatedByUserID        gocql.UUID                    `json:"created_by_user_id"`
	ModifiedAt             time.Time                     `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID       gocql.UUID                    `json:"modified_by_user_id"`
}

type CollectionMembershipDTO struct {
	ID                     gocql.UUID `bson:"_id" json:"id"`
	CollectionID           gocql.UUID `bson:"collection_id" json:"collection_id"`
	RecipientID            gocql.UUID `bson:"recipient_id" json:"recipient_id"`
	RecipientEmail         string     `bson:"recipient_email" json:"recipient_email"`
	GrantedByID            gocql.UUID `bson:"granted_by_id" json:"granted_by_id"`
	EncryptedCollectionKey []byte     `bson:"encrypted_collection_key" json:"encrypted_collection_key"`
	PermissionLevel        string     `bson:"permission_level" json:"permission_level"`
	CreatedAt              time.Time  `bson:"created_at" json:"created_at"`
	IsInherited            bool       `bson:"is_inherited" json:"is_inherited"`
	InheritedFromID        gocql.UUID `bson:"inherited_from_id,omitempty" json:"inherited_from_id,omitempty"`
}

type CollectionResponseDTO struct {
	ID                     gocql.UUID                   `json:"id"`
	OwnerID                gocql.UUID                   `json:"owner_id"`
	EncryptedName          string                       `json:"encrypted_name"`
	CollectionType         string                       `json:"collection_type"`
	ParentID               gocql.UUID                   `json:"parent_id,omitempty"`
	AncestorIDs            []gocql.UUID                 `json:"ancestor_ids,omitempty"`
	EncryptedCollectionKey *keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
	Children               []*CollectionResponseDTO     `json:"children,omitempty"`
	CreatedAt              time.Time                    `json:"created_at"`
	ModifiedAt             time.Time                    `json:"modified_at"`
	Members                []MembershipResponseDTO      `json:"members"`
}

type MembershipResponseDTO struct {
	ID             gocql.UUID `bson:"_id" json:"id"`
	CollectionID   gocql.UUID `bson:"collection_id" json:"collection_id"`     // ID of the collection (redundant but helpful for queries)
	RecipientID    gocql.UUID `bson:"recipient_id" json:"recipient_id"`       // User receiving access
	RecipientEmail string     `bson:"recipient_email" json:"recipient_email"` // Email for display purposes
	GrantedByID    gocql.UUID `bson:"granted_by_id" json:"granted_by_id"`     // User who shared the collection

	// Collection key encrypted with recipient's public key using box_seal. This matches the box_seal format which doesn't need a separate nonce.
	EncryptedCollectionKey []byte `bson:"encrypted_collection_key" json:"encrypted_collection_key"`

	// Access details
	PermissionLevel string    `bson:"permission_level" json:"permission_level"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`

	// Sharing origin tracking
	IsInherited     bool       `bson:"is_inherited" json:"is_inherited"`                               // Tracks whether access was granted directly or inherited from a parent
	InheritedFromID gocql.UUID `bson:"inherited_from_id,omitempty" json:"inherited_from_id,omitempty"` // InheritedFromID identifies which parent collection granted this access
}

type CreateCollectionService interface {
	Execute(ctx context.Context, req *CreateCollectionRequestDTO) (*CollectionResponseDTO, error)
}

type createCollectionServiceImpl struct {
	config                      *config.Configuration
	logger                      *zap.Logger
	userGetByIDUseCase          uc_user.UserGetByIDUseCase
	federatedUserGetByIDUseCase uc_federateduser.FederatedUserGetByIDUseCase
	repo                        dom_collection.CollectionRepository
}

func NewCreateCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	userGetByIDUseCase uc_user.UserGetByIDUseCase,
	federatedUserGetByIDUseCase uc_federateduser.FederatedUserGetByIDUseCase,
	repo dom_collection.CollectionRepository,
) CreateCollectionService {
	logger = logger.Named("CreateCollectionService")
	return &createCollectionServiceImpl{
		config:                      config,
		logger:                      logger,
		userGetByIDUseCase:          userGetByIDUseCase,
		federatedUserGetByIDUseCase: federatedUserGetByIDUseCase,
		repo:                        repo,
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
	if req.ID.IsZero() {
		e["encrypted_name"] = "Client-side generated ID is required"
	}
	if req.EncryptedName == "" {
		e["encrypted_name"] = "Collection name is required"
	}
	if req.CollectionType == "" {
		e["collection_type"] = "Collection type is required"
	} else if req.CollectionType != dom_collection.CollectionTypeFolder && req.CollectionType != dom_collection.CollectionTypeAlbum {
		e["collection_type"] = "Collection type must be either 'folder' or 'album'"
	}
	// Check pointer and then content
	if req.EncryptedCollectionKey == nil || req.EncryptedCollectionKey.Ciphertext == nil || len(req.EncryptedCollectionKey.Ciphertext) == 0 {
		e["encrypted_collection_key"] = "Encrypted collection key ciphertext is required"
	}
	if req.EncryptedCollectionKey == nil || req.EncryptedCollectionKey.Nonce == nil || len(req.EncryptedCollectionKey.Nonce) == 0 {
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
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	federateduser, err := svc.federatedUserGetByIDUseCase.Execute(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Failed getting federated user from database: %v", err)
	}
	if federateduser == nil {
		return nil, fmt.Errorf("User does not exist for federated iam id: %v", userID.Hex())
	}

	//
	// STEP 3: Create collection object by mapping DTO and applying server-side logic
	//
	now := time.Now()

	// Map all fields from the request DTO to the domain object.
	// This copies client-provided values including potential ID, OwnerID, timestamps, etc.
	collection := mapCollectionDTOToDomain(req, userID, now)

	// Apply server-side mandatory fields/overrides for the top-level collection.
	// These values are managed by the backend regardless of what the client provides in the DTO.
	// This ensures data integrity and reflects the server's perspective of the creation event.
	collection.ID = primitive.NewObjectID()                 // Always generate a new ID on the server for a new creation
	collection.OwnerID = userID                             // The authenticated user is the authoritative owner
	collection.CreatedAt = now                              // Server timestamp for creation
	collection.ModifiedAt = now                             // Server timestamp for modification
	collection.CreatedByUserID = userID                     // The authenticated user is the creator
	collection.ModifiedByUserID = userID                    // The authenticated user is the initial modifier
	collection.Version = 1                                  // Collection creation **always** starts mutation version at 1.
	collection.State = dom_collection.CollectionStateActive // Collection creation **always** starts in active state.

	// Ensure owner membership exists with Admin permissions.
	// Check if the owner is already present in the members list copied from the DTO.
	ownerAlreadyMember := false
	for i := range collection.Members { // Iterate by index to allow modification if needed
		if collection.Members[i].RecipientID == userID {
			// Owner is found. Ensure they have Admin permission and correct granted_by/is_inherited status.
			collection.Members[i].RecipientEmail = federateduser.Email
			collection.Members[i].PermissionLevel = dom_collection.CollectionPermissionAdmin
			collection.Members[i].GrantedByID = userID
			collection.Members[i].IsInherited = false
			// Optionally update membership CreatedAt here if server should control it, otherwise keep DTO value.
			// collection.Members[i].CreatedAt = now
			ownerAlreadyMember = true
			svc.logger.Debug("✅ Owner membership exists with Admin permissions")
			break
		}
	}

	// If owner is not in the members list, add their mandatory membership.
	if !ownerAlreadyMember {
		svc.logger.Debug("☑️ Owner is not in the members list, add their mandatory membership now")
		ownerMembership := dom_collection.CollectionMembership{
			ID:              primitive.NewObjectID(), // Unique ID for this specific membership record
			RecipientID:     userID,
			RecipientEmail:  federateduser.Email,
			CollectionID:    collection.ID,                            // Link to the newly created collection ID
			PermissionLevel: dom_collection.CollectionPermissionAdmin, // Owner must have Admin
			GrantedByID:     userID,                                   // Owner implicitly grants themselves permission
			IsInherited:     false,                                    // Owner membership is never inherited
			CreatedAt:       now,                                      // Server timestamp for membership creation
			// InheritedFromID is nil for direct membership.
		}
		// Append the mandatory owner membership. If req.Members was empty, this initializes the slice.
		collection.Members = append(collection.Members, ownerMembership)

		svc.logger.Debug("✅ Owner membership added with Admin permissions")
	}

	svc.logger.Debug("🔍 Collection debugging info",
		zap.String("collectionID", collection.ID.Hex()),
		zap.String("collectionOwnerID", collection.OwnerID.Hex()),
		zap.String("currentUserID", userID.Hex()),
		zap.Int("totalMembers", len(collection.Members)),
		zap.String("encryptedName", collection.EncryptedName))

	for i, memberDTO := range collection.Members {
		svc.logger.Debug("🔍 Cloud collection member DTO",
			zap.Int("memberIndex", i),
			zap.String("memberID", memberDTO.ID.Hex()),
			zap.String("recipientID", memberDTO.RecipientID.Hex()),
			zap.String("recipientEmail", memberDTO.RecipientEmail),
			zap.String("permissionLevel", memberDTO.PermissionLevel),
			zap.Bool("isInherited", memberDTO.IsInherited),
			zap.Int("encryptedKeyLength", len(memberDTO.EncryptedCollectionKey)))
	}

	// ENHANCED DEBUGGING: Log current user info for comparison
	svc.logger.Debug("🔍 Current user info for comparison",
		zap.String("currentUserID", federateduser.ID.Hex()),
		zap.String("currentUserEmail", federateduser.Email),
		zap.String("currentUserName", federateduser.Name))

	// Note: Fields like ParentID, AncestorIDs, EncryptedCollectionKey,
	// EncryptedName, CollectionType, and recursively mapped Children are copied directly from the DTO
	// by the mapCollectionDTOToDomain function before server overrides. This fulfills the
	// prompt's requirement to copy these fields from the DTO.

	//
	// STEP 4: Create collection in repository
	//

	if err := svc.repo.Create(ctx, collection); err != nil {
		svc.logger.Error("Failed to create collection",
			zap.Any("error", err),
			zap.Any("owner_id", collection.OwnerID),
			zap.String("name", collection.EncryptedName))
		return nil, err
	}

	//
	// STEP 5: Map domain model to response DTO
	//
	// The mapCollectionToDTO helper is used here to convert the created domain object back
	// into the response DTO format, potentially excluding sensitive fields like keys
	// or specific membership details not meant for the general response.
	response := mapCollectionToDTO(collection)

	svc.logger.Debug("Collection created successfully",
		zap.Any("collection_id", collection.ID),
		zap.Any("owner_id", collection.OwnerID))

	return response, nil
}
