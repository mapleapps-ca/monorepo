package recovery

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/authdto"
	uc_medto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/medto"
	uc_recovery "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/recovery"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// RecoveryService provides high-level functionality for account recovery
type RecoveryService interface {
	// InitiateRecovery starts the recovery process
	InitiateRecovery(ctx context.Context, email string) (*RecoveryInitiateOutput, error)

	// VerifyRecoveryKey verifies the recovery key and prepares for password reset
	VerifyRecoveryKey(ctx context.Context, sessionID string, recoveryKey string) (*RecoveryVerifyOutput, error)

	// CompleteRecovery sets new password and completes the recovery
	CompleteRecovery(ctx context.Context, recoveryToken string, newPassword string) (*RecoveryCompleteOutput, error)

	// GetRecoveryStatus returns the current recovery session status
	GetRecoveryStatus(ctx context.Context) (*RecoveryStatus, error)
}

// RecoveryInitiateOutput represents the output of recovery initiation
type RecoveryInitiateOutput struct {
	SessionID          string    `json:"session_id"`
	ChallengeID        string    `json:"challenge_id"`
	EncryptedChallenge string    `json:"encrypted_challenge"`
	ExpiresAt          time.Time `json:"expires_at"`
}

// RecoveryVerifyOutput represents the output of recovery verification
type RecoveryVerifyOutput struct {
	RecoveryToken                     string    `json:"recovery_token"`
	MasterKeyEncryptedWithRecoveryKey string    `json:"master_key_encrypted_with_recovery_key"`
	ExpiresAt                         time.Time `json:"expires_at"`
}

// RecoveryCompleteOutput represents the output of recovery completion
type RecoveryCompleteOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Email   string `json:"email"`
}

// RecoveryStatus represents the current state of recovery
type RecoveryStatus struct {
	InProgress bool       `json:"in_progress"`
	SessionID  string     `json:"session_id,omitempty"`
	Email      string     `json:"email,omitempty"`
	Stage      string     `json:"stage,omitempty"` // "initiated", "verified", "completed"
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// recoveryService implements the RecoveryService interface
type recoveryService struct {
	logger                      *zap.Logger
	configService               config.ConfigService
	userRepo                    user.Repository
	authRecoveryUseCase         uc_authdto.RecoveryUseCase
	initiateRecoveryUseCase     uc_recovery.InitiateRecoveryUseCase
	verifyRecoveryUseCase       uc_recovery.VerifyRecoveryUseCase
	completeRecoveryUseCase     uc_recovery.CompleteRecoveryUseCase
	checkRateLimitUseCase       uc_recovery.CheckRateLimitUseCase
	trackRecoveryAttemptUseCase uc_recovery.TrackRecoveryAttemptUseCase
	getRecoverySessionUseCase   uc_recovery.GetRecoverySessionUseCase
	getMeFromCloudUseCase       uc_medto.GetMeFromCloudUseCase
	stateManager                RecoveryStateManager

	// In-memory storage for recovery session state
	mu            sync.RWMutex
	currentStatus *RecoveryStatus
	recoveryData  *uc_authdto.RecoveryData // Store decrypted keys temporarily
	recoveryToken string                   // Store recovery token
}

// NewRecoveryService creates a new recovery service
func NewRecoveryService(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	authRecoveryUseCase uc_authdto.RecoveryUseCase,
	initiateRecoveryUseCase uc_recovery.InitiateRecoveryUseCase,
	verifyRecoveryUseCase uc_recovery.VerifyRecoveryUseCase,
	completeRecoveryUseCase uc_recovery.CompleteRecoveryUseCase,
	checkRateLimitUseCase uc_recovery.CheckRateLimitUseCase,
	trackRecoveryAttemptUseCase uc_recovery.TrackRecoveryAttemptUseCase,
	getRecoverySessionUseCase uc_recovery.GetRecoverySessionUseCase,
	getMeFromCloudUseCase uc_medto.GetMeFromCloudUseCase,
	stateManager RecoveryStateManager,
) RecoveryService {
	logger = logger.Named("RecoveryService")
	return &recoveryService{
		logger:                      logger,
		configService:               configService,
		userRepo:                    userRepo,
		authRecoveryUseCase:         authRecoveryUseCase,
		initiateRecoveryUseCase:     initiateRecoveryUseCase,
		verifyRecoveryUseCase:       verifyRecoveryUseCase,
		completeRecoveryUseCase:     completeRecoveryUseCase,
		checkRateLimitUseCase:       checkRateLimitUseCase,
		trackRecoveryAttemptUseCase: trackRecoveryAttemptUseCase,
		getRecoverySessionUseCase:   getRecoverySessionUseCase,
		getMeFromCloudUseCase:       getMeFromCloudUseCase,
		stateManager:                stateManager,
	}
}

// InitiateRecovery starts the recovery process
func (s *recoveryService) InitiateRecovery(ctx context.Context, email string) (*RecoveryInitiateOutput, error) {
	s.logger.Info("🔐 Initiating account recovery", zap.String("email", email))

	// Get IP address for rate limiting (in real app, get from request context)
	ipAddress := "127.0.0.1" // Default for CLI
	userAgent := "maplefile-cli"

	//
	// STEP 1: Check rate limit
	//
	if err := s.checkRateLimitUseCase.Execute(ctx, email, ipAddress); err != nil {
		// Track failed attempt
		_ = s.trackRecoveryAttemptUseCase.Execute(ctx, email, ipAddress, "recovery_key", false, userAgent)

		s.logger.Warn("⚠️ Recovery rate limit exceeded",
			zap.String("email", email),
			zap.String("ipAddress", ipAddress))
		return nil, errors.NewAppError("too many recovery attempts, please try again later", err)
	}

	//
	// STEP 2: Initiate recovery
	//
	response, err := s.initiateRecoveryUseCase.Execute(ctx, email, "recovery_key")
	if err != nil {
		// Track failed attempt
		_ = s.trackRecoveryAttemptUseCase.Execute(ctx, email, ipAddress, "recovery_key", false, userAgent)

		s.logger.Error("❌ Failed to initiate recovery", zap.Error(err))
		return nil, err
	}

	// Track successful initiation
	_ = s.trackRecoveryAttemptUseCase.Execute(ctx, email, ipAddress, "recovery_key", true, userAgent)

	//
	// STEP 3: Update local status and save to persistent storage
	//
	expiresAt := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)

	s.mu.Lock()
	s.currentStatus = &RecoveryStatus{
		InProgress: true,
		SessionID:  response.SessionID,
		Email:      email,
		Stage:      "initiated",
		ExpiresAt:  &expiresAt,
	}
	s.mu.Unlock()

	// Save state to persistent storage
	if err := s.stateManager.SaveState(ctx, s.currentStatus); err != nil {
		s.logger.Warn("Failed to save recovery state", zap.Error(err))
		// Continue anyway - this is not critical for the recovery process
	}

	s.logger.Info("✅ Recovery initiated successfully",
		zap.String("sessionID", response.SessionID),
		zap.String("challengeID", response.ChallengeID),
		zap.Time("expiresAt", expiresAt))

	return &RecoveryInitiateOutput{
		SessionID:          response.SessionID,
		ChallengeID:        response.ChallengeID,
		EncryptedChallenge: response.EncryptedChallenge,
		ExpiresAt:          expiresAt,
	}, nil
}

// VerifyRecoveryKey verifies the recovery key and prepares for password reset
func (s *recoveryService) VerifyRecoveryKey(ctx context.Context, sessionID string, recoveryKey string) (*RecoveryVerifyOutput, error) {
	s.logger.Info("🔐 Verifying recovery key", zap.String("sessionID", sessionID))

	//
	// STEP 1: First try to get session from repository (persistent storage)
	//
	session, err := s.getRecoverySessionUseCase.Execute(ctx, sessionID)
	if err != nil {
		s.logger.Error("❌ Failed to get recovery session", zap.Error(err))
		return nil, err
	}

	if session == nil {
		return nil, errors.NewAppError("recovery session not found", nil)
	}

	// Check if session has expired
	if session.IsExpired() {
		s.mu.Lock()
		s.currentStatus = nil
		s.recoveryData = nil
		s.recoveryToken = ""
		s.mu.Unlock()
		return nil, errors.NewAppError("recovery session has expired", nil)
	}

	// Check if session can be verified
	if !session.CanVerify() {
		return nil, errors.NewAppError("recovery session cannot be verified (expired or already verified)", nil)
	}

	//
	// STEP 2: Clean and normalize the recovery key format
	//
	cleanRecoveryKey := s.normalizeRecoveryKey(recoveryKey)
	s.logger.Debug("Normalized recovery key format")

	//
	// STEP 3: Get the user from local storage
	//
	user, err := s.userRepo.GetByEmail(ctx, session.Email)
	if err != nil {
		s.logger.Error("❌ Failed to get user", zap.Error(err))
		return nil, errors.NewAppError("failed to get user data", err)
	}

	if user == nil {
		s.logger.Error("❌ User not found locally", zap.String("email", session.Email))
		return nil, errors.NewAppError("user not found locally. Please ensure you have logged in before attempting recovery.", nil)
	}

	//
	// STEP 4: Validate the recovery key against local user data
	//
	if err := s.validateRecoveryKeyLocally(ctx, user, cleanRecoveryKey); err != nil {
		s.logger.Error("❌ Recovery key validation failed", zap.Error(err))
		return nil, errors.NewAppError("invalid recovery key", err)
	}

	//
	// STEP 5: Decrypt master key using recovery key to prepare recovery data
	//
	recoveryData, err := s.prepareRecoveryData(ctx, user, cleanRecoveryKey)
	if err != nil {
		s.logger.Error("❌ Failed to prepare recovery data", zap.Error(err))
		return nil, err
	}

	//
	// STEP 6: Update in-memory status to match the found session
	//
	s.mu.Lock()
	s.currentStatus = &RecoveryStatus{
		InProgress: true,
		SessionID:  sessionID,
		Email:      session.Email,
		Stage:      "initiated",
		ExpiresAt:  &session.ExpiresAt,
	}
	s.mu.Unlock()

	//
	// STEP 7: Verify with cloud service
	//
	response, err := s.verifyRecoveryUseCase.Execute(ctx, sessionID, cleanRecoveryKey)
	if err != nil {
		s.logger.Error("❌ Failed to verify recovery with cloud", zap.Error(err))
		return nil, err
	}

	//
	// STEP 8: Store recovery data for completion phase and save state
	//
	s.mu.Lock()
	s.recoveryData = recoveryData
	s.recoveryToken = response.RecoveryToken
	s.currentStatus.Stage = "verified"
	expiresAt := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
	s.currentStatus.ExpiresAt = &expiresAt
	s.mu.Unlock()

	// Save updated state to persistent storage
	if err := s.stateManager.SaveState(ctx, s.currentStatus); err != nil {
		s.logger.Warn("Failed to save recovery state after verification", zap.Error(err))
		// Continue anyway - this is not critical for the recovery process
	}

	// Save recovery data to persistent storage
	if err := s.stateManager.SaveRecoveryData(ctx, recoveryData, response.RecoveryToken); err != nil {
		s.logger.Warn("Failed to save recovery data after verification", zap.Error(err))
		// Continue anyway - this is not critical for the recovery process
	}

	s.logger.Info("✅ Recovery key verified successfully",
		zap.String("sessionID", sessionID),
		zap.Time("expiresAt", expiresAt))

	return &RecoveryVerifyOutput{
		RecoveryToken:                     response.RecoveryToken,
		MasterKeyEncryptedWithRecoveryKey: response.MasterKeyEncryptedWithRecoveryKey,
		ExpiresAt:                         expiresAt,
	}, nil
}

// validateRecoveryKeyLocally validates the recovery key against local user data
func (s *recoveryService) validateRecoveryKeyLocally(ctx context.Context, user *user.User, recoveryKey string) error {
	// Decode recovery key
	recoveryKeyBytes, err := base64.StdEncoding.DecodeString(recoveryKey)
	if err != nil {
		// Try URL-safe encoding
		recoveryKeyBytes, err = base64.RawURLEncoding.DecodeString(recoveryKey)
		if err != nil {
			return errors.NewAppError("invalid recovery key format", err)
		}
	}

	// Check if user has encrypted master key with recovery key
	if len(user.MasterKeyEncryptedWithRecoveryKey.Ciphertext) == 0 {
		return errors.NewAppError("no recovery key configured for this account", nil)
	}

	// Try to decrypt master key with the provided recovery key
	_, err = crypto.DecryptWithSecretBox(
		user.MasterKeyEncryptedWithRecoveryKey.Ciphertext,
		user.MasterKeyEncryptedWithRecoveryKey.Nonce,
		recoveryKeyBytes,
	)

	if err != nil {
		return errors.NewAppError("invalid recovery key", nil)
	}

	return nil
}

// prepareRecoveryData prepares the recovery data needed for password reset
func (s *recoveryService) prepareRecoveryData(ctx context.Context, user *user.User, recoveryKey string) (*uc_authdto.RecoveryData, error) {
	// Decode recovery key
	recoveryKeyBytes, err := base64.StdEncoding.DecodeString(recoveryKey)
	if err != nil {
		// Try URL-safe encoding
		recoveryKeyBytes, err = base64.RawURLEncoding.DecodeString(recoveryKey)
		if err != nil {
			return nil, errors.NewAppError("invalid recovery key format", err)
		}
	}

	// Decrypt master key using recovery key
	masterKey, err := crypto.DecryptWithSecretBox(
		user.MasterKeyEncryptedWithRecoveryKey.Ciphertext,
		user.MasterKeyEncryptedWithRecoveryKey.Nonce,
		recoveryKeyBytes,
	)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt master key with recovery key", err)
	}

	// Prepare recovery data
	recoveryData := &uc_authdto.RecoveryData{
		Email:     user.Email,
		MasterKey: masterKey,
		// Add other fields as needed based on the RecoveryData structure
	}

	return recoveryData, nil
}

// normalizeRecoveryKey cleans up the recovery key format to be a proper base64 string
func (s *recoveryService) normalizeRecoveryKey(recoveryKey string) string {
	// Remove any whitespace
	cleanKey := strings.TrimSpace(recoveryKey)

	// Remove hyphens and spaces that might be used for formatting
	cleanKey = strings.ReplaceAll(cleanKey, "-", "")
	cleanKey = strings.ReplaceAll(cleanKey, " ", "")

	s.logger.Debug("Recovery key normalized",
		zap.String("original_length", fmt.Sprintf("%d", len(recoveryKey))),
		zap.String("cleaned_length", fmt.Sprintf("%d", len(cleanKey))))

	return cleanKey
}

// CompleteRecovery sets new password and completes the recovery
func (s *recoveryService) CompleteRecovery(ctx context.Context, recoveryToken string, newPassword string) (*RecoveryCompleteOutput, error) {
	s.logger.Info("🔐 Completing account recovery")

	//
	// STEP 1: Try to restore recovery state from persistent storage if not in memory
	//
	s.mu.RLock()
	status := s.currentStatus
	recoveryData := s.recoveryData
	storedRecoveryToken := s.recoveryToken
	s.mu.RUnlock()

	// If no in-memory state, try to restore from persistent storage
	if status == nil || !status.InProgress || recoveryData == nil {
		s.logger.Info("🔄 No in-memory recovery state found, attempting to restore from persistent storage")

		restoredStatus, err := s.stateManager.FindActiveSession(ctx)
		if err != nil {
			s.logger.Error("❌ Failed to find active recovery session", zap.Error(err))
			return nil, errors.NewAppError("failed to find active recovery session", err)
		}

		if restoredStatus == nil || !restoredStatus.InProgress {
			return nil, errors.NewAppError("no active recovery session found. Please start the recovery process again.", nil)
		}

		if restoredStatus.Stage != "verified" {
			return nil, errors.NewAppError(fmt.Sprintf("recovery session not verified (current stage: %s). Please verify your recovery key first.", restoredStatus.Stage), nil)
		}

		// Restore recovery data from persistent storage
		if err := s.restoreRecoveryData(ctx, restoredStatus); err != nil {
			s.logger.Error("❌ Failed to restore recovery data", zap.Error(err))
			return nil, err
		}

		// Update in-memory state
		s.mu.Lock()
		s.currentStatus = restoredStatus
		s.mu.Unlock()

		// Re-read the state after restoration
		s.mu.RLock()
		status = s.currentStatus
		recoveryData = s.recoveryData
		storedRecoveryToken = s.recoveryToken
		s.mu.RUnlock()
	}

	if status == nil || !status.InProgress {
		return nil, errors.NewAppError("no active recovery session", nil)
	}

	if status.Stage != "verified" {
		return nil, errors.NewAppError("recovery session not verified", nil)
	}

	if recoveryData == nil {
		return nil, errors.NewAppError("recovery data not found", nil)
	}

	if status.ExpiresAt != nil && time.Now().After(*status.ExpiresAt) {
		s.mu.Lock()
		s.currentStatus = nil
		s.recoveryData = nil
		s.recoveryToken = ""
		s.mu.Unlock()
		_ = s.stateManager.ClearState(ctx)
		_ = s.stateManager.ClearRecoveryData(ctx)
		return nil, errors.NewAppError("recovery session has expired", nil)
	}

	// Use provided recovery token or stored one
	finalRecoveryToken := recoveryToken
	if finalRecoveryToken == "" && storedRecoveryToken != "" {
		finalRecoveryToken = storedRecoveryToken
		s.logger.Debug("Using stored recovery token")
	}

	if finalRecoveryToken == "" {
		return nil, errors.NewAppError("recovery token is required and not found in storage", nil)
	}

	//
	// STEP 2: Start transaction
	//
	if err := s.userRepo.OpenTransaction(); err != nil {
		return nil, errors.NewAppError("failed to open transaction", err)
	}

	// Ensure transaction cleanup
	defer func() {
		if s.userRepo.OpenTransaction() == nil { // Check if still in transaction
			s.userRepo.DiscardTransaction()
		}
	}()

	//
	// STEP 3: Complete recovery with cloud
	//
	response, err := s.completeRecoveryUseCase.Execute(ctx, finalRecoveryToken, newPassword, recoveryData.MasterKey)
	if err != nil {
		s.logger.Error("❌ Failed to complete recovery with cloud", zap.Error(err))
		return nil, err
	}

	if !response.Success {
		return nil, errors.NewAppError(fmt.Sprintf("recovery failed: %s", response.Message), nil)
	}

	//
	// STEP 4: Use auth recovery use case to update local user
	//
	authResponse, updatedUser, err := s.authRecoveryUseCase.CompleteRecovery(ctx, recoveryData, newPassword)
	if err != nil {
		s.logger.Error("❌ Failed to complete local recovery", zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Save updated user
	//
	if err := s.userRepo.UpsertByEmail(ctx, updatedUser); err != nil {
		s.logger.Error("❌ Failed to save updated user", zap.Error(err))
		return nil, errors.NewAppError("failed to update user data", err)
	}

	//
	// STEP 6: Save authenticated credentials if provided
	//
	if authResponse.AccessToken != "" && authResponse.RefreshToken != "" {
		if err := s.configService.SetLoggedInUserCredentials(
			ctx,
			recoveryData.Email,
			authResponse.AccessToken,
			&authResponse.AccessTokenExpiryTime,
			authResponse.RefreshToken,
			&authResponse.RefreshTokenExpiryTime,
		); err != nil {
			s.logger.Warn("⚠️ Failed to save credentials after recovery", zap.Error(err))
			// Continue anyway - recovery was successful
		}
	}

	//
	// STEP 7: Fetch and update profile from cloud
	//
	if authResponse.AccessToken != "" {
		meDTO, err := s.getMeFromCloudUseCase.Execute(ctx)
		if err != nil {
			s.logger.Warn("⚠️ Failed to fetch profile after recovery", zap.Error(err))
		} else if meDTO != nil {
			// Update local user profile
			updatedUser.ID = meDTO.ID
			updatedUser.FirstName = meDTO.FirstName
			updatedUser.LastName = meDTO.LastName
			updatedUser.Name = meDTO.Name
			updatedUser.LexicalName = meDTO.LexicalName
			updatedUser.Role = meDTO.Role
			updatedUser.WasEmailVerified = meDTO.WasEmailVerified
			updatedUser.Phone = meDTO.Phone
			updatedUser.Country = meDTO.Country
			updatedUser.Timezone = meDTO.Timezone
			updatedUser.Region = meDTO.Region
			updatedUser.City = meDTO.City
			updatedUser.PostalCode = meDTO.PostalCode
			updatedUser.AddressLine1 = meDTO.AddressLine1
			updatedUser.AddressLine2 = meDTO.AddressLine2
			updatedUser.AgreePromotions = meDTO.AgreePromotions
			updatedUser.AgreeToTrackingAcrossThirdPartyAppsAndServices = meDTO.AgreeToTrackingAcrossThirdPartyAppsAndServices
			updatedUser.CreatedAt = meDTO.CreatedAt
			updatedUser.Status = meDTO.Status

			if err := s.userRepo.UpsertByEmail(ctx, updatedUser); err != nil {
				s.logger.Warn("⚠️ Failed to update user profile after recovery", zap.Error(err))
			}
		}
	}

	//
	// STEP 8: Commit transaction
	//
	if err := s.userRepo.CommitTransaction(); err != nil {
		return nil, errors.NewAppError("failed to commit transaction", err)
	}

	//
	// STEP 9: Clear recovery state and data
	//
	s.mu.Lock()
	s.currentStatus = nil
	s.recoveryData = nil
	s.recoveryToken = ""
	s.mu.Unlock()

	// Clear persistent state and data
	if err := s.stateManager.ClearState(ctx); err != nil {
		s.logger.Warn("Failed to clear recovery state", zap.Error(err))
	}
	if err := s.stateManager.ClearRecoveryData(ctx); err != nil {
		s.logger.Warn("Failed to clear recovery data", zap.Error(err))
	}

	// Display new recovery key to user
	newRecoveryKey := s.generateRecoveryKeyDisplay(updatedUser)

	s.logger.Info("✅ Account recovery completed successfully",
		zap.String("email", recoveryData.Email))

	return &RecoveryCompleteOutput{
		Success: true,
		Message: fmt.Sprintf("Password reset successfully. Your new recovery key: %s", newRecoveryKey),
		Email:   recoveryData.Email,
	}, nil
}

// restoreRecoveryData attempts to restore recovery data from the session and user
func (s *recoveryService) restoreRecoveryData(ctx context.Context, status *RecoveryStatus) error {
	s.logger.Debug("🔄 Attempting to restore recovery data from persistent storage")

	if status.Email == "" {
		return errors.NewAppError("no email in recovery status", nil)
	}

	// Load recovery data from persistent storage
	recoveryData, recoveryToken, err := s.stateManager.LoadRecoveryData(ctx)
	if err != nil {
		return errors.NewAppError("failed to load recovery data from storage", err)
	}

	if recoveryData == nil {
		s.logger.Debug("No recovery data found in persistent storage")
		// Try to restore basic data without master key
		user, err := s.userRepo.GetByEmail(ctx, status.Email)
		if err != nil {
			return errors.NewAppError("failed to get user for recovery restoration", err)
		}

		if user == nil {
			return errors.NewAppError("user not found for recovery restoration", nil)
		}

		s.mu.Lock()
		s.recoveryData = &uc_authdto.RecoveryData{
			Email: user.Email,
			// MasterKey will need to be provided again during completion
		}
		s.mu.Unlock()

		return errors.NewAppError("recovery data not found in memory. Please provide your recovery key again to complete the process.", nil)
	}

	// Restore full recovery data
	s.mu.Lock()
	s.recoveryData = recoveryData
	s.recoveryToken = recoveryToken
	s.mu.Unlock()

	s.logger.Info("✅ Successfully restored recovery data from persistent storage")
	return nil
}

// GetRecoveryStatus returns the current recovery session status
func (s *recoveryService) GetRecoveryStatus(ctx context.Context) (*RecoveryStatus, error) {
	s.mu.RLock()
	memoryStatus := s.currentStatus
	s.mu.RUnlock()

	// If we have in-memory status, use it
	if memoryStatus != nil && memoryStatus.InProgress {
		// Check if expired
		if memoryStatus.ExpiresAt != nil && time.Now().After(*memoryStatus.ExpiresAt) {
			// Clear expired status
			s.mu.Lock()
			s.currentStatus = nil
			s.recoveryData = nil
			s.recoveryToken = ""
			s.mu.Unlock()
			_ = s.stateManager.ClearState(ctx)
			_ = s.stateManager.ClearRecoveryData(ctx)

			return &RecoveryStatus{
				InProgress: false,
			}, nil
		}

		// Return a copy of the status
		return &RecoveryStatus{
			InProgress: memoryStatus.InProgress,
			SessionID:  memoryStatus.SessionID,
			Email:      memoryStatus.Email,
			Stage:      memoryStatus.Stage,
			ExpiresAt:  memoryStatus.ExpiresAt,
		}, nil
	}

	// If no in-memory status, try to load from persistent storage
	persistentStatus, err := s.stateManager.FindActiveSession(ctx)
	if err != nil {
		s.logger.Error("Failed to find active session", zap.Error(err))
		return &RecoveryStatus{InProgress: false}, nil
	}

	if persistentStatus != nil && persistentStatus.InProgress {
		// Update in-memory state
		s.mu.Lock()
		s.currentStatus = persistentStatus
		s.mu.Unlock()
	}

	return persistentStatus, nil
}

// generateRecoveryKeyDisplay generates a display-friendly recovery key
func (s *recoveryService) generateRecoveryKeyDisplay(user *user.User) string {
	// In a real implementation, this would decrypt and display the actual recovery key
	// For now, return a placeholder
	if user.EncryptedRecoveryKey.Ciphertext != nil {
		// The actual recovery key is encrypted, we'd need to decrypt it
		// This is just for display purposes
		return "XXXX-XXXX-XXXX-XXXX-XXXX-XXXX"
	}
	return "No recovery key found"
}
