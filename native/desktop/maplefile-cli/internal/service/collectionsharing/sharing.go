// internal/service/collectionsharing/sharing.go (UPDATED)
package collectionsharing

import (
	"context"
	"encoding/base64"
	"fmt"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
	dom_publiclookupdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/publiclookupdto"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectionsharingdto"
	uc_publiclookupdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/publiclookupdto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// ShareCollectionInput represents input for sharing a collection at the service level
type ShareCollectionInput struct {
	CollectionID         gocql.UUID `json:"collection_id"`
	RecipientEmail       string     `json:"recipient_email"`
	PermissionLevel      string     `json:"permission_level"`
	ShareWithDescendants bool       `json:"share_with_descendants"`
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
	ExecuteBatchSharing(ctx context.Context, input *BatchShareCollectionInput, userPassword string) (*BatchShareCollectionOutput, error)
}

// Batch sharing input for multiple recipients
type BatchShareCollectionInput struct {
	CollectionID         gocql.UUID      `json:"collection_id"`
	Recipients           []RecipientInfo `json:"recipients"`
	ShareWithDescendants bool            `json:"share_with_descendants"`
}

type RecipientInfo struct {
	Email           string `json:"email"`
	PermissionLevel string `json:"permission_level"`
}

type BatchShareCollectionOutput struct {
	Success                 bool                      `json:"success"`
	Message                 string                    `json:"message"`
	TotalMembershipsCreated int                       `json:"total_memberships_created"`
	Results                 []IndividualSharingResult `json:"results"`
}

type IndividualSharingResult struct {
	RecipientEmail     string `json:"recipient_email"`
	Success            bool   `json:"success"`
	MembershipsCreated int    `json:"memberships_created"`
	Error              string `json:"error,omitempty"`
}

// collectionSharingService implements the enhanced CollectionSharingService interface
type collectionSharingService struct {
	logger                          *zap.Logger
	getCollectionUseCase            uc_collection.GetCollectionUseCase
	getPublicLookupFromCloudUseCase uc_publiclookupdto.GetPublicLookupFromCloudUseCase
	getUserByIsLoggedInUseCase      uc_user.GetByIsLoggedInUseCase
	shareCollectionUseCase          uc.ShareCollectionUseCase
	collectionEncryptionService     svc_collectioncrypto.CollectionEncryptionService
}

// NewCollectionSharingService creates a new enhanced collection sharing service
func NewCollectionSharingService(
	logger *zap.Logger,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	getPublicLookupFromCloudUseCase uc_publiclookupdto.GetPublicLookupFromCloudUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	shareCollectionUseCase uc.ShareCollectionUseCase,
	collectionEncryptionService svc_collectioncrypto.CollectionEncryptionService,
) CollectionSharingService {
	logger = logger.Named("CollectionSharingService")
	return &collectionSharingService{
		logger:                          logger,
		getCollectionUseCase:            getCollectionUseCase,
		getPublicLookupFromCloudUseCase: getPublicLookupFromCloudUseCase,
		getUserByIsLoggedInUseCase:      getUserByIsLoggedInUseCase,
		shareCollectionUseCase:          shareCollectionUseCase,
		collectionEncryptionService:     collectionEncryptionService,
	}
}

// Execute shares a collection with another user using extended crypto service
func (s *collectionSharingService) Execute(ctx context.Context, input *ShareCollectionInput, userPassword string) (*ShareCollectionOutput, error) {
	//
	// STEP 1: Validate inputs
	//
	if input == nil {
		s.logger.Error("❌ Input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.CollectionID.String() == "" {
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
	// STEP 2: Get and validate related records
	//
	publicLookupResponse, currentUser, collectionToShare, err := s.validateAndGetSharingData(ctx, input)
	if err != nil {
		return nil, err
	}

	//
	// STEP 3: Get public key of the other user
	//
	publicKeyBytes, err := s.decodePublicKey(publicLookupResponse.PublicKeyInBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode recipient public key: %v", err)
	}

	// E2EE key chain + BoxSeal encryption logic
	encryptedCollectionKey, err := s.collectionEncryptionService.EncryptCollectionKeyForSharing(
		ctx,
		currentUser,
		collectionToShare,
		publicKeyBytes,
		userPassword,
	)
	if err != nil {
		s.logger.Error("❌ Failed to encrypt collection key for sharing", zap.Error(err))
		return nil, errors.NewAppError("failed to encrypt collection key for sharing", err)
	}

	s.logger.Debug("✅ Successfully encrypted collection key using extended crypto service",
		zap.String("collectionID", input.CollectionID.String()),
		zap.String("recipientEmail", input.RecipientEmail))

	//
	// STEP 4: Submit share request to cloud
	//
	useCaseInput := &uc.ShareCollectionInputDTO{
		CollectionID:           input.CollectionID,
		RecipientID:            publicLookupResponse.UserID,
		RecipientEmail:         publicLookupResponse.Email,
		PermissionLevel:        input.PermissionLevel,
		EncryptedCollectionKey: encryptedCollectionKey,
		ShareWithDescendants:   input.ShareWithDescendants,
	}

	response, err := s.shareCollectionUseCase.Execute(ctx, useCaseInput, userPassword)
	if err != nil {
		s.logger.Error("❌ Failed to share collection", zap.Error(err))
		return nil, err
	}

	s.logger.Info("✅ Successfully shared collection using extended crypto service",
		zap.String("collectionID", input.CollectionID.String()),
		zap.String("recipientEmail", input.RecipientEmail),
		zap.String("permissionLevel", input.PermissionLevel))

	return &ShareCollectionOutput{
		Success:            response.Success,
		Message:            response.Message,
		MembershipsCreated: response.MembershipsCreated,
	}, nil
}

// Batch sharing using extended crypto service efficiency
func (s *collectionSharingService) ExecuteBatchSharing(ctx context.Context, input *BatchShareCollectionInput, userPassword string) (*BatchShareCollectionOutput, error) {
	s.logger.Info("🚀 Starting batch collection sharing using extended crypto service",
		zap.String("collectionID", input.CollectionID.String()),
		zap.Int("recipientCount", len(input.Recipients)))

	// STEP 1: Validate inputs
	if input == nil || len(input.Recipients) == 0 {
		return nil, errors.NewAppError("input and recipients are required", nil)
	}

	// STEP 2: Get collection and user data
	collectionToShare, err := s.getCollectionUseCase.Execute(ctx, input.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection", err)
	}

	currentUser, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get current user", err)
	}

	// STEP 3: Lookup all recipients and build crypto service input
	cryptoRecipients := make([]svc_collectioncrypto.SharingRecipient, 0, len(input.Recipients))
	recipientMap := make(map[string]RecipientInfo)

	for _, recipient := range input.Recipients {
		// Validate permission level
		if err := collectionsharingdto.ValidatePermissionLevel(recipient.PermissionLevel); err != nil {
			s.logger.Warn("⚠️ Skipping recipient with invalid permission level",
				zap.String("email", recipient.Email),
				zap.String("permissionLevel", recipient.PermissionLevel))
			continue
		}

		// Lookup recipient
		publicLookupRequest := &dom_publiclookupdto.PublicLookupRequestDTO{
			Email: recipient.Email,
		}
		publicLookupResponse, err := s.getPublicLookupFromCloudUseCase.Execute(ctx, publicLookupRequest)
		if err != nil {
			s.logger.Warn("⚠️ Skipping recipient due to lookup failure",
				zap.String("email", recipient.Email),
				zap.Error(err))
			continue
		}

		// Decode public key
		publicKeyBytes, err := s.decodePublicKey(publicLookupResponse.PublicKeyInBase64)
		if err != nil {
			s.logger.Warn("⚠️ Skipping recipient due to invalid public key",
				zap.String("email", recipient.Email),
				zap.Error(err))
			continue
		}

		cryptoRecipients = append(cryptoRecipients, svc_collectioncrypto.SharingRecipient{
			Email:     recipient.Email,
			PublicKey: publicKeyBytes,
			UserID:    publicLookupResponse.UserID.String(),
		})
		recipientMap[recipient.Email] = recipient
	}

	// STEP 4: ✅ MAJOR EFFICIENCY GAIN: Batch encrypt using extended crypto service
	// This decrypts the collection key ONCE and encrypts for all recipients
	encryptedKeys, err := s.collectionEncryptionService.EncryptCollectionKeyForMultipleRecipients(
		ctx,
		currentUser,
		collectionToShare,
		cryptoRecipients,
		userPassword,
	)
	if err != nil {
		return nil, fmt.Errorf("batch encryption failed: %w", err)
	}

	s.logger.Info("✅ Successfully batch encrypted collection keys using extended crypto service",
		zap.String("collectionID", input.CollectionID.String()),
		zap.Int("successfulRecipients", len(encryptedKeys)))

	// STEP 5: Submit individual share requests to cloud
	output := &BatchShareCollectionOutput{
		Results: make([]IndividualSharingResult, 0, len(encryptedKeys)),
	}

	for email, encryptedKey := range encryptedKeys {
		recipientInfo := recipientMap[email]

		// Submit to cloud (individual requests)
		// In a real implementation, you might want to batch these cloud requests too
		useCaseInput := &uc.ShareCollectionInputDTO{
			CollectionID:           input.CollectionID,
			RecipientEmail:         email,
			PermissionLevel:        recipientInfo.PermissionLevel,
			EncryptedCollectionKey: encryptedKey,
			ShareWithDescendants:   input.ShareWithDescendants,
		}

		response, err := s.shareCollectionUseCase.Execute(ctx, useCaseInput, userPassword)
		if err != nil {
			output.Results = append(output.Results, IndividualSharingResult{
				RecipientEmail: email,
				Success:        false,
				Error:          err.Error(),
			})
			continue
		}

		output.Results = append(output.Results, IndividualSharingResult{
			RecipientEmail:     email,
			Success:            response.Success,
			MembershipsCreated: response.MembershipsCreated,
		})
		output.TotalMembershipsCreated += response.MembershipsCreated
	}

	// Calculate success
	successCount := 0
	for _, result := range output.Results {
		if result.Success {
			successCount++
		}
	}

	output.Success = successCount > 0
	output.Message = fmt.Sprintf("Successfully shared with %d of %d recipients", successCount, len(input.Recipients))

	s.logger.Info("✅ Completed batch collection sharing using extended crypto service",
		zap.String("collectionID", input.CollectionID.String()),
		zap.Int("successfulShares", successCount),
		zap.Int("totalRecipients", len(input.Recipients)))

	return output, nil
}

// Helper methods remain the same but simplified...
func (s *collectionSharingService) validateAndGetSharingData(ctx context.Context, input *ShareCollectionInput) (*dom_publiclookupdto.PublicLookupResponseDTO, *dom_user.User, *collection.Collection, error) {
	// Lookup recipient
	publicLookupRequest := &dom_publiclookupdto.PublicLookupRequestDTO{
		Email: input.RecipientEmail,
	}
	publicLookupResponse, err := s.getPublicLookupFromCloudUseCase.Execute(ctx, publicLookupRequest)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to lookup recipient: %w", err)
	}

	// Get collection
	collectionToShare, err := s.getCollectionUseCase.Execute(ctx, input.CollectionID)
	if err != nil {
		return nil, nil, nil, errors.NewAppError("failed to get collection", err)
	}

	// Get current user
	currentUser, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, nil, nil, errors.NewAppError("failed to get current user", err)
	}

	// Validate sharing permissions (simplified)
	canShare := collectionToShare.OwnerID == currentUser.ID
	if !canShare {
		for _, member := range collectionToShare.Members {
			if member.RecipientID == currentUser.ID && member.PermissionLevel == collectionsharingdto.CollectionDTOPermissionAdmin {
				canShare = true
				break
			}
		}
	}
	if !canShare {
		return nil, nil, nil, errors.NewAppError("you don't have permission to share this collection", nil)
	}

	return publicLookupResponse, currentUser, collectionToShare, nil
}

func (s *collectionSharingService) decodePublicKey(publicKeyBase64 string) ([]byte, error) {
	if publicKeyBase64 == "" {
		return nil, fmt.Errorf("public key cannot be empty")
	}

	// Try most common encodings
	encodings := []struct {
		name     string
		encoding *base64.Encoding
	}{
		{"RawURL", base64.RawURLEncoding},
		{"Standard", base64.StdEncoding},
		{"URL", base64.URLEncoding},
	}

	for _, enc := range encodings {
		if publicKeyBytes, err := enc.encoding.DecodeString(publicKeyBase64); err == nil {
			s.logger.Debug("✅ Successfully decoded public key",
				zap.String("encoding", enc.name))
			return publicKeyBytes, nil
		}
	}

	return nil, fmt.Errorf("failed to decode public key with any base64 encoding")
}
