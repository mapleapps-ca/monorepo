// native/desktop/maplefile-cli/internal/usecase/collectionsharingdto/sharing.go
package collectionsharingdto

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// ShareCollectionInput represents input for sharing a collection
type ShareCollectionInputDTO struct {
	CollectionID           primitive.ObjectID `json:"collection_id"`
	RecipientID            primitive.ObjectID `json:"recipient_id"`
	RecipientEmail         string             `json:"recipient_email"`
	PermissionLevel        string             `json:"permission_level"`
	EncryptedCollectionKey *keys.EncryptedCollectionKey
	ShareWithDescendants   bool `json:"share_with_descendants"`
}

// ShareCollectionUseCase defines the interface for sharing collections
type ShareCollectionUseCase interface {
	Execute(ctx context.Context, input *ShareCollectionInputDTO, userPassword string) (*collectionsharingdto.ShareCollectionResponseDTO, error)
}

// shareCollectionUseCase implements the ShareCollectionUseCase interface
type shareCollectionUseCase struct {
	logger      *zap.Logger
	sharingRepo collectionsharingdto.CollectionSharingDTORepository
}

// NewShareCollectionUseCase creates a new use case for sharing collections
func NewShareCollectionUseCase(
	logger *zap.Logger,
	sharingRepo collectionsharingdto.CollectionSharingDTORepository,
) ShareCollectionUseCase {
	logger = logger.Named("ShareCollectionUseCase")
	return &shareCollectionUseCase{
		logger:      logger,
		sharingRepo: sharingRepo,
	}
}

// Execute shares a collection with another user
func (uc *shareCollectionUseCase) Execute(ctx context.Context, input *ShareCollectionInputDTO, userPassword string) (*collectionsharingdto.ShareCollectionResponseDTO, error) {
	// Validate inputs
	if input.CollectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.RecipientID.IsZero() {
		return nil, errors.NewAppError("recipient ID is required", nil)
	}
	if input.RecipientEmail == "" {
		return nil, errors.NewAppError("recipient email is required", nil)
	}
	if input.PermissionLevel == "" {
		return nil, errors.NewAppError("permission level is required", nil)
	}
	if input.PermissionLevel == "" {
		return nil, errors.NewAppError("permission level is required", nil)
	}
	if input.EncryptedCollectionKey == nil {
		return nil, errors.NewAppError("encrypted collection key is required for E2EE operations", nil)
	}
	if err := collectionsharingdto.ValidatePermissionLevel(input.PermissionLevel); err != nil {
		return nil, errors.NewAppError("invalid permission level", err)
	}

	// Create share request
	shareRequest := &collectionsharingdto.ShareCollectionRequestDTO{
		CollectionID:           input.CollectionID,
		RecipientID:            input.RecipientID,
		RecipientEmail:         input.RecipientEmail,
		PermissionLevel:        input.PermissionLevel,
		EncryptedCollectionKey: input.EncryptedCollectionKey,
		ShareWithDescendants:   input.ShareWithDescendants,
	}

	// Execute share operation via repository
	response, err := uc.sharingRepo.ShareCollectionInCloud(ctx, shareRequest)
	if err != nil {
		uc.logger.Error("❌ Failed to share collection", zap.Error(err))
		return nil, err
	}

	uc.logger.Info("✅ Successfully shared collection",
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.String("recipientEmail", input.RecipientEmail),
		zap.String("permissionLevel", input.PermissionLevel))

	return response, nil
}
