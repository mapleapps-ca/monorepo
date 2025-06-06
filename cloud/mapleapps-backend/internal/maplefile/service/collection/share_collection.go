// cloud/backend/internal/maplefile/service/collection/share_collection.go
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ShareCollectionRequestDTO struct {
	CollectionID           gocql.UUID `json:"collection_id"`
	RecipientID            gocql.UUID `json:"recipient_id"`
	RecipientEmail         string             `json:"recipient_email"`
	PermissionLevel        string             `json:"permission_level"`
	EncryptedCollectionKey []byte             `json:"encrypted_collection_key"`
	ShareWithDescendants   bool               `json:"share_with_descendants"`
}

type ShareCollectionResponseDTO struct {
	Success            bool   `json:"success"`
	Message            string `json:"message"`
	MembershipsCreated int    `json:"memberships_created,omitempty"`
}

type ShareCollectionService interface {
	Execute(ctx context.Context, req *ShareCollectionRequestDTO) (*ShareCollectionResponseDTO, error)
}

type shareCollectionServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewShareCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) ShareCollectionService {
	logger = logger.Named("ShareCollectionService")
	return &shareCollectionServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *shareCollectionServiceImpl) Execute(ctx context.Context, req *ShareCollectionRequestDTO) (*ShareCollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Share details are required")
	}

	e := make(map[string]string)
	if req.CollectionID.IsZero() {
		e["collection_id"] = "Collection ID is required"
	}
	if req.RecipientID.IsZero() {
		e["recipient_id"] = "Recipient ID is required"
	}
	if req.RecipientEmail == "" {
		e["recipient_email"] = "Recipient email is required"
	}
	if req.PermissionLevel == "" {
		// Will default to read-only in repository
	} else if req.PermissionLevel != dom_collection.CollectionPermissionReadOnly &&
		req.PermissionLevel != dom_collection.CollectionPermissionReadWrite &&
		req.PermissionLevel != dom_collection.CollectionPermissionAdmin {
		e["permission_level"] = "Invalid permission level"
	}
	if len(req.EncryptedCollectionKey) == 0 {
		e["encrypted_collection_key"] = "Encrypted collection key is required"
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

	//
	// STEP 3: Retrieve existing collection
	//
	collection, err := svc.repo.Get(ctx, req.CollectionID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.Any("collection_id", req.CollectionID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.Any("collection_id", req.CollectionID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	//
	// STEP 4: Check if user has rights to share this collection
	//
	hasSharePermission := false

	// Owner always has share permission
	if collection.OwnerID == userID {
		hasSharePermission = true
	} else {
		// Check if user is an admin member
		for _, member := range collection.Members {
			if member.RecipientID == userID && member.PermissionLevel == dom_collection.CollectionPermissionAdmin {
				hasSharePermission = true
				break
			}
		}
	}

	if !hasSharePermission {
		svc.logger.Warn("Unauthorized collection sharing attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to share this collection")
	}

	//
	// STEP 5: Create membership
	//
	membership := &dom_collection.CollectionMembership{
		ID:                     primitive.NewObjectID(),
		CollectionID:           req.CollectionID,
		RecipientID:            req.RecipientID,
		RecipientEmail:         req.RecipientEmail,
		GrantedByID:            userID,
		EncryptedCollectionKey: req.EncryptedCollectionKey,
		PermissionLevel:        req.PermissionLevel,
		CreatedAt:              time.Now(),
		IsInherited:            false,
	}

	//
	// STEP 6: Add membership to collection
	//
	var membershipsCreated int = 1

	if req.ShareWithDescendants {
		// Add member to collection and all descendants
		err = svc.repo.AddMemberToHierarchy(ctx, req.CollectionID, membership)
		if err != nil {
			svc.logger.Error("Failed to add member to collection hierarchy",
				zap.Any("error", err),
				zap.Any("collection_id", req.CollectionID),
				zap.Any("recipient_id", req.RecipientID))
			return nil, err
		}

		// Get the number of descendants to report how many memberships were created
		descendants, err := svc.repo.FindDescendants(ctx, req.CollectionID)
		if err == nil {
			membershipsCreated += len(descendants)
		}
	} else {
		// Add member just to this collection
		err = svc.repo.AddMember(ctx, req.CollectionID, membership)
		if err != nil {
			svc.logger.Error("Failed to add member to collection",
				zap.Any("error", err),
				zap.Any("collection_id", req.CollectionID),
				zap.Any("recipient_id", req.RecipientID))
			return nil, err
		}
	}

	svc.logger.Info("Collection shared successfully",
		zap.Any("collection_id", req.CollectionID),
		zap.Any("recipient_id", req.RecipientID),
		zap.Any("granted_by", userID),
		zap.String("permission_level", req.PermissionLevel),
		zap.Bool("shared_with_descendants", req.ShareWithDescendants),
		zap.Int("memberships_created", membershipsCreated))

	return &ShareCollectionResponseDTO{
		Success:            true,
		Message:            "Collection shared successfully",
		MembershipsCreated: membershipsCreated,
	}, nil
}
