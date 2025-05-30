// native/desktop/maplefile-cli/internal/usecase/collectiondsharingdto/sharing.go
package collectiondsharingdto

import (
	"context"
	"encoding/base64"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// ShareCollectionInput represents input for sharing a collection
type ShareCollectionInputDTO struct {
	CollectionID         primitive.ObjectID `json:"collection_id"`
	RecipientEmail       string             `json:"recipient_email"`
	PermissionLevel      string             `json:"permission_level"`
	ShareWithDescendants bool               `json:"share_with_descendants"`
}

// ShareCollectionUseCase defines the interface for sharing collections
type ShareCollectionUseCase interface {
	Execute(ctx context.Context, input *ShareCollectionInputDTO, userPassword string) (*collectionsharingdto.ShareCollectionResponseDTO, error)
}

// shareCollectionUseCase implements the ShareCollectionUseCase interface
type shareCollectionUseCase struct {
	logger                     *zap.Logger
	collectionRepo             collectiondto.CollectionDTORepository
	sharingRepo                collectionsharingdto.CollectionSharingDTORepository
	getCollectionUseCase       uc_collectiondto.GetCollectionFromCloudUseCase
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
	getUserByEmailUseCase      uc_user.GetByEmailUseCase
}

// NewShareCollectionUseCase creates a new use case for sharing collections
func NewShareCollectionUseCase(
	logger *zap.Logger,
	collectionRepo collectiondto.CollectionDTORepository,
	sharingRepo collectionsharingdto.CollectionSharingDTORepository,
	getCollectionUseCase uc_collectiondto.GetCollectionFromCloudUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getUserByEmailUseCase uc_user.GetByEmailUseCase,
) ShareCollectionUseCase {
	logger = logger.Named("ShareCollectionUseCase")
	return &shareCollectionUseCase{
		logger:                     logger,
		collectionRepo:             collectionRepo,
		sharingRepo:                sharingRepo,
		getCollectionUseCase:       getCollectionUseCase,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
		getUserByEmailUseCase:      getUserByEmailUseCase,
	}
}

// Execute shares a collection with another user
func (uc *shareCollectionUseCase) Execute(ctx context.Context, input *ShareCollectionInputDTO, userPassword string) (*collectionsharingdto.ShareCollectionResponseDTO, error) {
	// Validate inputs
	if input.CollectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.RecipientEmail == "" {
		return nil, errors.NewAppError("recipient email is required", nil)
	}
	if userPassword == "" {
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
	}
	if err := collectionsharingdto.ValidatePermissionLevel(input.PermissionLevel); err != nil {
		return nil, errors.NewAppError("invalid permission level", err)
	}

	// Get the collection to share
	collectionToShare, err := uc.getCollectionUseCase.Execute(ctx, input.CollectionID)
	if err != nil {
		uc.logger.Error("❌ Failed to get collection to share", zap.Error(err))
		return nil, errors.NewAppError("failed to get collection", err)
	}
	if collectionToShare == nil {
		return nil, errors.NewAppError("collection not found", nil)
	}

	// Get current user (the one sharing)
	currentUser, err := uc.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		uc.logger.Error("❌ Failed to get current user", zap.Error(err))
		return nil, errors.NewAppError("failed to get current user", err)
	}
	if currentUser == nil {
		return nil, errors.NewAppError("user not authenticated", nil)
	}

	// Check if current user has permission to share this collection
	// User must be owner or have admin permission
	canShare := collectionToShare.OwnerID == currentUser.ID
	if !canShare {
		// Check if user is an admin member
		for _, member := range collectionToShare.Members {
			if member.RecipientID == currentUser.ID && member.PermissionLevel == collectionsharingdto.CollectionDTOPermissionAdmin {
				canShare = true
				break
			}
		}
	}
	if !canShare {
		return nil, errors.NewAppError("you don't have permission to share this collection", nil)
	}

	// Get recipient user information
	recipientUser, err := uc.getUserByEmailUseCase.Execute(ctx, input.RecipientEmail)
	if err != nil {
		uc.logger.Error("❌ Failed to get recipient user", zap.String("email", input.RecipientEmail), zap.Error(err))
		return nil, errors.NewAppError("failed to get recipient user", err)
	}
	if recipientUser == nil {
		return nil, errors.NewAppError("recipient user not found", nil)
	}

	// Check if user is trying to share with themselves
	if recipientUser.ID == currentUser.ID {
		return nil, errors.NewAppError("cannot share collection with yourself", nil)
	}

	// Check if recipient already has access
	for _, member := range collectionToShare.Members {
		if member.RecipientID == recipientUser.ID {
			return nil, errors.NewAppError("recipient already has access to this collection", nil)
		}
	}

	// Encrypt collection key for recipient (E2EE)
	encryptedCollectionKey, err := uc.encryptCollectionKeyForRecipient(
		ctx,
		currentUser,
		recipientUser,
		collectionToShare,
		userPassword,
	)
	if err != nil {
		uc.logger.Error("❌ Failed to encrypt collection key for recipient", zap.Error(err))
		return nil, errors.NewAppError("failed to encrypt collection key for recipient", err)
	}

	// Create share request
	shareRequest := &collectionsharingdto.ShareCollectionRequestDTO{
		CollectionID:           input.CollectionID,
		RecipientID:            recipientUser.ID,
		RecipientEmail:         input.RecipientEmail,
		PermissionLevel:        input.PermissionLevel,
		EncryptedCollectionKey: encryptedCollectionKey,
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

// encryptCollectionKeyForRecipient encrypts the collection key for the recipient using their public key
func (uc *shareCollectionUseCase) encryptCollectionKeyForRecipient(
	ctx context.Context,
	currentUser *dom_user.User,
	recipientUser *dom_user.User,
	collectionToShare *collectiondto.CollectionDTO,
	userPassword string,
) (string, error) {
	// Step 1: Derive keyEncryptionKey from current user's password
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(userPassword, currentUser.PasswordSalt)
	if err != nil {
		return "", fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ClearBytes(keyEncryptionKey)

	// Step 2: Decrypt current user's master key
	masterKey, err := crypto.DecryptWithSecretBox(
		currentUser.EncryptedMasterKey.Ciphertext,
		currentUser.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt master key: %w", err)
	}
	defer crypto.ClearBytes(masterKey)

	// Step 3: Decrypt collection key using master key
	collectionKey, err := crypto.DecryptWithSecretBox(
		collectionToShare.EncryptedCollectionKey.Ciphertext,
		collectionToShare.EncryptedCollectionKey.Nonce,
		masterKey,
	)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt collection key: %w", err)
	}
	defer crypto.ClearBytes(collectionKey)

	// Step 4: Encrypt collection key with recipient's public key using box_seal
	encryptedForRecipient, err := crypto.EncryptWithBoxSeal(collectionKey, recipientUser.PublicKey.Key)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt collection key for recipient: %w", err)
	}

	// Step 5: Encode to base64 for transmission
	return base64.StdEncoding.EncodeToString(encryptedForRecipient), nil
}
