// internal/service/collectionsharing/sharing.go
package collectionsharing

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_publiclookupdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/publiclookupdto"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectionsharingdto"
	uc_publiclookupdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/publiclookupdto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// ShareCollectionInput represents input for sharing a collection at the service level
type ShareCollectionInput struct {
	CollectionID         primitive.ObjectID `json:"collection_id"`
	RecipientEmail       string             `json:"recipient_email"`
	PermissionLevel      string             `json:"permission_level"`
	ShareWithDescendants bool               `json:"share_with_descendants"`
}

// ShareCollectionOutput represents the output from sharing a collection
type ShareCollectionOutput struct {
	Success            bool   `json:"success"`
	Message            string `json:"message"`
	MembershipsCreated int    `json:"memberships_created"`
}

// CollectionSharingService defines the interface for collection sharing operations
type CollectionSharingService interface {
	Execute(ctx context.Context, input *ShareCollectionInput, userPassword string) (*ShareCollectionOutput, error)
}

// collectionSharingService implements the CollectionSharingService interface
type collectionSharingService struct {
	logger                          *zap.Logger
	getCollectionUseCase            uc_collection.GetCollectionUseCase
	getPublicLookupFromCloudUseCase uc_publiclookupdto.GetPublicLookupFromCloudUseCase
	getUserByIsLoggedInUseCase      uc_user.GetByIsLoggedInUseCase
	shareCollectionUseCase          uc.ShareCollectionUseCase
}

// NewCollectionSharingService creates a new collection sharing service
func NewCollectionSharingService(
	logger *zap.Logger,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	getPublicLookupFromCloudUseCase uc_publiclookupdto.GetPublicLookupFromCloudUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	shareCollectionUseCase uc.ShareCollectionUseCase,
) CollectionSharingService {
	logger = logger.Named("CollectionSharingService")
	return &collectionSharingService{
		logger:                          logger,
		getCollectionUseCase:            getCollectionUseCase,
		getPublicLookupFromCloudUseCase: getPublicLookupFromCloudUseCase,
		getUserByIsLoggedInUseCase:      getUserByIsLoggedInUseCase,
		shareCollectionUseCase:          shareCollectionUseCase,
	}
}

// Execute shares a collection with another user
func (s *collectionSharingService) Execute(ctx context.Context, input *ShareCollectionInput, userPassword string) (*ShareCollectionOutput, error) {
	//
	// STEP 1: Validate inputs
	//

	if input == nil {
		s.logger.Error("❌ Input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.CollectionID.IsZero() {
		s.logger.Error("❌ Collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.RecipientEmail == "" {
		s.logger.Error("❌ Recipient email is required")
		return nil, errors.NewAppError("recipient email is required", nil)
	}
	if input.PermissionLevel == "" {
		s.logger.Error("❌ Permission level is required")
		return nil, errors.NewAppError("permission level is required", nil)
	}
	if userPassword == "" {
		s.logger.Error("❌ User password is required for E2EE operations")
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
	}

	// Validate permission level
	if err := collectionsharingdto.ValidatePermissionLevel(input.PermissionLevel); err != nil {
		s.logger.Error("❌ Invalid permission level", zap.String("level", input.PermissionLevel), zap.Error(err))
		return nil, errors.NewAppError("invalid permission level", err)
	}

	//
	// STEP 2: Lookup the receipients email from the cloud so to
	// (1) confirm email exists and (2) get the public key.
	//

	publicLookupRequest := &dom_publiclookupdto.PublicLookupRequestDTO{
		Email: input.RecipientEmail,
	}
	publicLookupResponse, err := s.getPublicLookupFromCloudUseCase.Execute(ctx, publicLookupRequest)
	if err != nil {
		if strings.Contains(err.Error(), "email") {
			err := fmt.Errorf("email does not exist: %v", input.RecipientEmail)
			s.logger.Error("Failed lookup up email",
				zap.String("email", input.RecipientEmail), zap.Error(err))
			return nil, err
		}
		s.logger.Error("Failed lookup up email",
			zap.String("email", input.RecipientEmail), zap.Error(err))
		return nil, err
	}
	if publicLookupResponse == nil {
		err := fmt.Errorf("nothing returned from cloud for email: %s", input.RecipientEmail)
		s.logger.Error("Failed lookup up email",
			zap.String("email", input.RecipientEmail), zap.Error(err))
		return nil, err
	}

	//
	// STEP 3: Get any related records.
	//

	// Get the collection to share
	collectionToShare, err := s.getCollectionUseCase.Execute(ctx, input.CollectionID)
	if err != nil {
		s.logger.Error("❌ Failed to get collection to share", zap.Error(err))
		return nil, errors.NewAppError("failed to get collection", err)
	}
	if collectionToShare == nil {
		s.logger.Error("❌ Collection not found")
		return nil, errors.NewAppError("collection not found", nil)
	}

	// Get current user (the one sharing)
	currentUser, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("❌ Failed to get current user", zap.Error(err))
		return nil, errors.NewAppError("failed to get current user", err)
	}
	if currentUser == nil {
		return nil, errors.NewAppError("user not authenticated", nil)
	}

	//
	// STEP 4: Validation of related records.
	//

	// Check if current (logged in) user has permission to share this collection.
	// User must be owner or have admin permission.
	canShare := collectionToShare.OwnerID == currentUser.ID
	if !canShare {
		s.logger.Debug("🔍 Checking if user is an admin member")
		// Check if user is an admin member
		for _, member := range collectionToShare.Members {
			s.logger.Debug("🔍 Member sharing check",
				zap.Any("member.RecipientID", member.RecipientID),
				zap.Any("currentUser.ID", currentUser.ID),
				zap.Any("member.PermissionLevel", member.PermissionLevel),
				zap.Any("collectionsharingdto.CollectionDTOPermissionAdmin", collectionsharingdto.CollectionDTOPermissionAdmin),
			)
			if member.RecipientID == currentUser.ID && member.PermissionLevel == collectionsharingdto.CollectionDTOPermissionAdmin {
				s.logger.Debug("✅ Member sharing check passed!")
				canShare = true
				break
			}
		}
	}
	if !canShare {
		s.logger.Error("🚫 You don't have permission to share this collection",
			zap.Any("collectionToShare.OwnerID", collectionToShare.OwnerID),
			zap.Any("currentUser.ID", currentUser.ID),
		)
		return nil, errors.NewAppError("you don't have permission to share this collection", nil)
	}

	// Check if user is trying to share with themselves
	if publicLookupResponse.UserID == currentUser.ID {
		return nil, errors.NewAppError("cannot share collection with yourself", nil)
	}

	// Check if recipient already has access
	for _, member := range collectionToShare.Members {
		if member.RecipientID == publicLookupResponse.UserID {
			return nil, errors.NewAppError("recipient already has access to this collection", nil)
		}
	}

	//
	// STEP 5: Encrypt collection key for recipient (E2EE)
	//

	// Decode the public key with fallback to multiple base64 encodings
	publicKeyBytes, err := s.decodePublicKeyFromBase64(publicLookupResponse.PublicKeyInBase64)
	if err != nil {
		s.logger.Error("❌ Failed to decode recipient public key", zap.Error(err))
		return nil, fmt.Errorf("failed to decode recipient public key: %v", err)
	}

	encryptedCollectionKey, err := s.encryptCollectionKeyForRecipient(
		ctx,
		currentUser,
		publicKeyBytes,
		collectionToShare,
		userPassword,
	)
	if err != nil {
		s.logger.Error("❌ Failed to encrypt collection key for recipient", zap.Error(err))
		return nil, errors.NewAppError("failed to encrypt collection key for recipient", err)
	}
	if encryptedCollectionKey == nil {
		return nil, errors.NewAppError("could not encrypt collection key", nil)
	}

	//
	// STEP 6: Submit our share request to the cloud backend.
	//

	// Create use case input
	useCaseInput := &uc.ShareCollectionInputDTO{
		CollectionID:           input.CollectionID,
		RecipientID:            publicLookupResponse.UserID,
		RecipientEmail:         publicLookupResponse.Email,
		PermissionLevel:        input.PermissionLevel,
		EncryptedCollectionKey: encryptedCollectionKey, // Pass the encrypted collection key struct for E2EE
		ShareWithDescendants:   input.ShareWithDescendants,
	}

	s.logger.Debug("🔍 Sharing request details",
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.String("recipientEmail", input.RecipientEmail),
		zap.Int("encryptedKeyLength", len(encryptedCollectionKey.ToBoxSealBytes())))

	// Execute use case
	response, err := s.shareCollectionUseCase.Execute(ctx, useCaseInput, userPassword)
	if err != nil {
		s.logger.Error("❌ Failed to share collection",
			zap.String("collectionID", input.CollectionID.Hex()),
			zap.String("recipientEmail", input.RecipientEmail),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("✅ Successfully shared collection",
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.String("recipientEmail", input.RecipientEmail),
		zap.String("permissionLevel", input.PermissionLevel))

	return &ShareCollectionOutput{
		Success:            response.Success,
		Message:            response.Message,
		MembershipsCreated: response.MembershipsCreated,
	}, nil
}

// decodePublicKeyFromBase64 attempts to decode a public key using multiple base64 encodings
// This handles different base64 formats that might be returned from the server
func (s *collectionSharingService) decodePublicKeyFromBase64(publicKeyBase64 string) ([]byte, error) {
	if publicKeyBase64 == "" {
		return nil, fmt.Errorf("public key cannot be empty")
	}

	// Try URL-safe base64 without padding first (most common in our system)
	publicKeyBytes, err := base64.RawURLEncoding.DecodeString(publicKeyBase64)
	if err == nil {
		s.logger.Debug("✅ Successfully decoded public key using RawURLEncoding")
		return publicKeyBytes, nil
	}

	s.logger.Debug("🔄 RawURLEncoding failed, trying StdEncoding", zap.Error(err))

	// Try standard base64 encoding as fallback
	publicKeyBytes, err = base64.StdEncoding.DecodeString(publicKeyBase64)
	if err == nil {
		s.logger.Debug("✅ Successfully decoded public key using StdEncoding")
		return publicKeyBytes, nil
	}

	s.logger.Debug("🔄 StdEncoding failed, trying URLEncoding", zap.Error(err))

	// Try URL-safe base64 with padding as another fallback
	publicKeyBytes, err = base64.URLEncoding.DecodeString(publicKeyBase64)
	if err == nil {
		s.logger.Debug("✅ Successfully decoded public key using URLEncoding")
		return publicKeyBytes, nil
	}

	s.logger.Error("❌ All base64 decoding attempts failed for public key",
		zap.String("publicKeyBase64", publicKeyBase64),
		zap.Error(err))

	return nil, fmt.Errorf("failed to decode public key with any base64 encoding (tried RawURL, Std, URL): %v", err)
}

// encryptCollectionKeyForRecipient encrypts the collection key for the recipient using their public key
// Returns the EncryptedCollectionKey struct instead of base64 string
func (s *collectionSharingService) encryptCollectionKeyForRecipient(
	ctx context.Context,
	currentUser *dom_user.User,
	recipientUserPublicKey []byte,
	collectionToShare *collection.Collection,
	userPassword string,
) (*keys.EncryptedCollectionKey, error) {
	// Step 1: Derive keyEncryptionKey from current user's password
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(userPassword, currentUser.PasswordSalt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ClearBytes(keyEncryptionKey)

	// Step 2: Decrypt current user's master key
	masterKey, err := crypto.DecryptWithSecretBox(
		currentUser.EncryptedMasterKey.Ciphertext,
		currentUser.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt master key: %w", err)
	}
	defer crypto.ClearBytes(masterKey)

	// Step 3: Decrypt collection key using master key
	collectionKey, err := crypto.DecryptWithSecretBox(
		collectionToShare.EncryptedCollectionKey.Ciphertext,
		collectionToShare.EncryptedCollectionKey.Nonce,
		masterKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt collection key: %w", err)
	}
	defer crypto.ClearBytes(collectionKey)

	// Step 4: Encrypt collection key with recipient's public key using box_seal
	encryptedForRecipient, err := crypto.EncryptWithBoxSeal(collectionKey, recipientUserPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt collection key for recipient: %w", err)
	}

	// Step 5: Create EncryptedCollectionKey struct from box_seal bytes
	encryptedCollectionKey := keys.NewEncryptedCollectionKeyFromBoxSeal(encryptedForRecipient)

	if err := s.verifyEncryptedKey(encryptedCollectionKey, recipientUserPublicKey); err != nil {
		return nil, fmt.Errorf("failed to verify encrypt collection key for recipient: %w", err)
	}

	return encryptedCollectionKey, nil
}

func (s *collectionSharingService) verifyEncryptedKey(encryptedKey *keys.EncryptedCollectionKey, recipientPublicKey []byte) error {
	// Get the box_seal bytes
	encryptedBytes := encryptedKey.ToBoxSealBytes()
	if encryptedBytes == nil {
		return fmt.Errorf("encrypted key is nil")
	}

	// Verify it's the right length for box_seal format
	expectedMinLength := crypto.BoxPublicKeySize + crypto.BoxNonceSize + crypto.BoxOverhead
	if len(encryptedBytes) < expectedMinLength {
		return fmt.Errorf("encrypted key too short: got %d, expected at least %d",
			len(encryptedBytes), expectedMinLength)
	}

	s.logger.Debug("✅ Encrypted key format validation passed",
		zap.Int("encryptedKeyLength", len(encryptedBytes)),
		zap.Int("expectedMinLength", expectedMinLength))

	return nil
}
