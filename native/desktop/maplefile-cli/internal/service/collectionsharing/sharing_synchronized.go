// internal/service/collectionsharing/sharing_synchronized.go
package collectionsharing

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_publiclookupdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/publiclookupdto"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectionsharingdto"
	uc_publiclookupdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/publiclookupdto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// SynchronizedCollectionSharingService ensures local state stays in sync after sharing
type SynchronizedCollectionSharingService interface {
	ExecuteWithSync(ctx context.Context, input *ShareCollectionInput, userPassword string) (*ShareCollectionOutput, error)
}

// synchronizedCollectionSharingService implements the SynchronizedCollectionSharingService interface
type synchronizedCollectionSharingService struct {
	logger *zap.Logger

	// Original sharing dependencies
	getCollectionUseCase            uc_collection.GetCollectionUseCase
	getPublicLookupFromCloudUseCase uc_publiclookupdto.GetPublicLookupFromCloudUseCase
	getUserByIsLoggedInUseCase      uc_user.GetByIsLoggedInUseCase
	shareCollectionUseCase          uc.ShareCollectionUseCase

	// Additional dependencies for local sync
	updateCollectionUseCase   uc_collection.UpdateCollectionUseCase
	localCollectionRepository collection.CollectionRepository

	// Crypto service
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService
}

// NewSynchronizedCollectionSharingService creates a new synchronized collection sharing service
func NewSynchronizedCollectionSharingService(
	logger *zap.Logger,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	getPublicLookupFromCloudUseCase uc_publiclookupdto.GetPublicLookupFromCloudUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	shareCollectionUseCase uc.ShareCollectionUseCase,
	updateCollectionUseCase uc_collection.UpdateCollectionUseCase,
	localCollectionRepository collection.CollectionRepository,
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
) SynchronizedCollectionSharingService {
	logger = logger.Named("SynchronizedCollectionSharingService")
	return &synchronizedCollectionSharingService{
		logger:                          logger,
		getCollectionUseCase:            getCollectionUseCase,
		getPublicLookupFromCloudUseCase: getPublicLookupFromCloudUseCase,
		getUserByIsLoggedInUseCase:      getUserByIsLoggedInUseCase,
		shareCollectionUseCase:          shareCollectionUseCase,
		updateCollectionUseCase:         updateCollectionUseCase,
		localCollectionRepository:       localCollectionRepository,
		collectionDecryptionService:     collectionDecryptionService,
	}
}

// ExecuteWithSync shares a collection and updates local state
func (s *synchronizedCollectionSharingService) ExecuteWithSync(
	ctx context.Context,
	input *ShareCollectionInput,
	userPassword string,
) (*ShareCollectionOutput, error) {
	s.logger.Info("🔄 Starting synchronized collection sharing using crypto service",
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.String("recipientEmail", input.RecipientEmail),
		zap.String("permissionLevel", input.PermissionLevel))

	//
	// STEP 1: Execute the original sharing logic (cloud update)
	//

	shareOutput, err := s.executeCloudSharing(ctx, input, userPassword)
	if err != nil {
		s.logger.Error("❌ Failed to share collection in cloud", zap.Error(err))
		return nil, err
	}

	s.logger.Info("✅ Successfully shared collection in cloud using crypto service",
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.Int("membershipsCreated", shareOutput.MembershipsCreated))

	//
	// STEP 2: Update local collection with new member
	//

	if err := s.updateLocalCollectionWithNewMember(ctx, input, userPassword); err != nil {
		s.logger.Error("⚠️ Failed to update local collection after sharing",
			zap.String("collectionID", input.CollectionID.Hex()),
			zap.Error(err))

		// Don't fail the entire operation since cloud sharing succeeded
		// But warn the user about potential sync issues
		s.logger.Warn("🚨 Collection shared successfully in cloud, but local sync failed. "+
			"Local collection may be out of sync. Consider running a manual sync.",
			zap.String("collectionID", input.CollectionID.Hex()))
	} else {
		s.logger.Info("✅ Successfully synchronized local collection with new member using crypto service",
			zap.String("collectionID", input.CollectionID.Hex()),
			zap.String("recipientEmail", input.RecipientEmail))
	}

	return shareOutput, nil
}

// executeCloudSharing performs the original cloud sharing logic using crypto service
func (s *synchronizedCollectionSharingService) executeCloudSharing(
	ctx context.Context,
	input *ShareCollectionInput,
	userPassword string,
) (*ShareCollectionOutput, error) {
	// Validate inputs
	if input == nil {
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.CollectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.RecipientEmail == "" {
		return nil, errors.NewAppError("recipient email is required", nil)
	}
	if input.PermissionLevel == "" {
		return nil, errors.NewAppError("permission level is required", nil)
	}
	if userPassword == "" {
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
	}

	// Validate permission level
	if err := collectionsharingdto.ValidatePermissionLevel(input.PermissionLevel); err != nil {
		return nil, errors.NewAppError("invalid permission level", err)
	}

	// Get recipient public key
	publicLookupRequest := &dom_publiclookupdto.PublicLookupRequestDTO{
		Email: input.RecipientEmail,
	}
	publicLookupResponse, err := s.getPublicLookupFromCloudUseCase.Execute(ctx, publicLookupRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup recipient: %w", err)
	}

	// Get collection and validate sharing permissions
	collectionToShare, err := s.getCollectionUseCase.Execute(ctx, input.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection", err)
	}

	currentUser, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get current user", err)
	}

	// Validate sharing permissions (simplified - use your existing logic)
	canShare := collectionToShare.OwnerID == currentUser.ID
	if !canShare {
		// Check for admin permissions in existing logic
		for _, member := range collectionToShare.Members {
			if member.RecipientID == currentUser.ID &&
				member.PermissionLevel == collectionsharingdto.CollectionDTOPermissionAdmin {
				canShare = true
				break
			}
		}
	}
	if !canShare {
		return nil, errors.NewAppError("you don't have permission to share this collection", nil)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicLookupResponse.PublicKeyInBase64)
	if err != nil {
		return nil, errors.NewAppError("failed to decode recipient public key", err)
	}

	encryptedCollectionKey, err := s.encryptCollectionKeyForRecipient(
		ctx, currentUser, publicKeyBytes, collectionToShare, userPassword,
	)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt collection key for recipient", err)
	}

	// Execute cloud sharing
	useCaseInput := &uc.ShareCollectionInputDTO{
		CollectionID:           input.CollectionID,
		RecipientID:            publicLookupResponse.UserID,
		RecipientEmail:         publicLookupResponse.Email,
		PermissionLevel:        input.PermissionLevel,
		EncryptedCollectionKey: encryptedCollectionKey,
		ShareWithDescendants:   input.ShareWithDescendants,
	}

	s.logger.Debug("🔍 Sharing request details",
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.String("recipientEmail", input.RecipientEmail),
		zap.Int("encryptedKeyLength", len(encryptedCollectionKey.ToBoxSealBytes())))

	response, err := s.shareCollectionUseCase.Execute(ctx, useCaseInput, userPassword)
	if err != nil {
		return nil, err
	}

	return &ShareCollectionOutput{
		Success:            response.Success,
		Message:            response.Message,
		MembershipsCreated: response.MembershipsCreated,
	}, nil
}

// updateLocalCollectionWithNewMember adds the new member to the local collection
func (s *synchronizedCollectionSharingService) updateLocalCollectionWithNewMember(
	ctx context.Context,
	input *ShareCollectionInput,
	userPassword string,
) error {
	s.logger.Debug("🔄 Updating local collection with new member using crypto service",
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.String("recipientEmail", input.RecipientEmail))

	// Get the current local collection
	localCollection, err := s.getCollectionUseCase.Execute(ctx, input.CollectionID)
	if err != nil {
		return fmt.Errorf("failed to get local collection: %w", err)
	}

	// Get recipient user information
	publicLookupRequest := &dom_publiclookupdto.PublicLookupRequestDTO{
		Email: input.RecipientEmail,
	}
	publicLookupResponse, err := s.getPublicLookupFromCloudUseCase.Execute(ctx, publicLookupRequest)
	if err != nil {
		return fmt.Errorf("failed to lookup recipient for local update: %w", err)
	}

	// Get current user for encryption
	currentUser, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current user for local update: %w", err)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicLookupResponse.PublicKeyInBase64)
	if err != nil {
		return fmt.Errorf("failed to decode public key for local update: %w", err)
	}

	encryptedKeyForMember, err := s.encryptCollectionKeyForRecipient(
		ctx, currentUser, publicKeyBytes, localCollection, userPassword,
	)
	if err != nil {
		return fmt.Errorf("failed to encrypt key for member: %w", err)
	}

	// Create new membership
	newMember := &collection.CollectionMembership{
		ID:                     primitive.NewObjectID(), // Generate new membership ID
		CollectionID:           input.CollectionID,
		RecipientID:            publicLookupResponse.UserID,
		RecipientEmail:         input.RecipientEmail,
		GrantedByID:            currentUser.ID,
		EncryptedCollectionKey: encryptedKeyForMember,
		PermissionLevel:        input.PermissionLevel,
		CreatedAt:              time.Now(),
		IsInherited:            false,
		InheritedFromID:        primitive.NilObjectID,
	}

	// Check if member already exists (shouldn't happen, but be defensive)
	for _, existingMember := range localCollection.Members {
		if existingMember.RecipientID == publicLookupResponse.UserID {
			s.logger.Warn("Member already exists in local collection, skipping local update",
				zap.String("collectionID", input.CollectionID.Hex()),
				zap.String("recipientEmail", input.RecipientEmail))
			return nil
		}
	}

	// Add new member to local collection
	localCollection.Members = append(localCollection.Members, newMember)

	// Update modification timestamp
	localCollection.ModifiedAt = time.Now()

	// Save updated collection locally
	if err := s.localCollectionRepository.Save(ctx, localCollection); err != nil {
		return fmt.Errorf("failed to save updated local collection: %w", err)
	}

	s.logger.Info("✅ Successfully added new member to local collection using crypto service",
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.String("newMemberEmail", input.RecipientEmail),
		zap.String("permissionLevel", input.PermissionLevel),
		zap.Int("totalMembers", len(localCollection.Members)))

	return nil
}

func (s *synchronizedCollectionSharingService) encryptCollectionKeyForRecipient(
	ctx context.Context,
	currentUser *dom_user.User,
	recipientPublicKey []byte,
	collectionToShare *collection.Collection,
	userPassword string,
) (*keys.EncryptedCollectionKey, error) {

	s.logger.Debug("🔐 Starting E2EE collection key encryption for recipient using crypto service")

	collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, currentUser, collectionToShare, userPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt collection key chain: %w", err)
	}
	defer crypto.ClearBytes(collectionKey)

	s.logger.Debug("✅ Successfully decrypted collection key using crypto service")

	s.logger.Debug("🔐 Encrypting collection key for recipient using BoxSeal")
	encryptedForRecipient, err := crypto.EncryptWithBoxSeal(collectionKey, recipientPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt collection key for recipient: %w", err)
	}

	// Create EncryptedCollectionKey struct from box_seal bytes
	encryptedCollectionKey := keys.NewEncryptedCollectionKeyFromBoxSeal(encryptedForRecipient)

	if err := s.verifyEncryptedKey(encryptedCollectionKey, recipientPublicKey); err != nil {
		return nil, fmt.Errorf("failed to verify encrypted collection key for recipient: %w", err)
	}

	s.logger.Debug("✅ Successfully encrypted collection key for recipient using crypto service")
	return encryptedCollectionKey, nil
}

func (s *synchronizedCollectionSharingService) verifyEncryptedKey(encryptedKey *keys.EncryptedCollectionKey, recipientPublicKey []byte) error {
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
