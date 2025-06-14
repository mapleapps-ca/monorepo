// native/desktop/maplefile-cli/internal/service/recovery/recovery.go
package recovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/authdto"
	uc_medto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/medto"
	uc_recovery "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/recovery"
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

	// In-memory storage for recovery session state
	mu            sync.RWMutex
	currentStatus *RecoveryStatus
	recoveryData  *uc_authdto.RecoveryData // Store decrypted keys temporarily
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
	// STEP 3: Update local status
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
	// STEP 1: Check current status
	//
	s.mu.RLock()
	status := s.currentStatus
	s.mu.RUnlock()

	if status == nil || !status.InProgress || status.SessionID != sessionID {
		return nil, errors.NewAppError("no active recovery session or session mismatch", nil)
	}

	if status.ExpiresAt != nil && time.Now().After(*status.ExpiresAt) {
		s.mu.Lock()
		s.currentStatus = nil
		s.recoveryData = nil
		s.mu.Unlock()
		return nil, errors.NewAppError("recovery session has expired", nil)
	}

	//
	// STEP 2: Use auth recovery use case to get decrypted keys
	//
	recoveryData, err := s.authRecoveryUseCase.InitiateRecovery(ctx, status.Email, recoveryKey)
	if err != nil {
		s.logger.Error("❌ Failed to verify recovery key", zap.Error(err))
		return nil, err
	}

	//
	// STEP 3: Verify with cloud service
	//
	response, err := s.verifyRecoveryUseCase.Execute(ctx, sessionID, recoveryKey)
	if err != nil {
		s.logger.Error("❌ Failed to verify recovery with cloud", zap.Error(err))
		return nil, err
	}

	//
	// STEP 4: Store recovery data for completion phase
	//
	s.mu.Lock()
	s.recoveryData = recoveryData
	s.currentStatus.Stage = "verified"
	expiresAt := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
	s.currentStatus.ExpiresAt = &expiresAt
	s.mu.Unlock()

	s.logger.Info("✅ Recovery key verified successfully",
		zap.String("sessionID", sessionID),
		zap.Time("expiresAt", expiresAt))

	return &RecoveryVerifyOutput{
		RecoveryToken:                     response.RecoveryToken,
		MasterKeyEncryptedWithRecoveryKey: response.MasterKeyEncryptedWithRecoveryKey,
		ExpiresAt:                         expiresAt,
	}, nil
}

// CompleteRecovery sets new password and completes the recovery
func (s *recoveryService) CompleteRecovery(ctx context.Context, recoveryToken string, newPassword string) (*RecoveryCompleteOutput, error) {
	s.logger.Info("🔐 Completing account recovery")

	//
	// STEP 1: Check current status and get recovery data
	//
	s.mu.RLock()
	status := s.currentStatus
	recoveryData := s.recoveryData
	s.mu.RUnlock()

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
		s.mu.Unlock()
		return nil, errors.NewAppError("recovery session has expired", nil)
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
	response, err := s.completeRecoveryUseCase.Execute(ctx, recoveryToken, newPassword, recoveryData.MasterKey)
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
	// STEP 9: Clear recovery state
	//
	s.mu.Lock()
	s.currentStatus = nil
	s.recoveryData = nil
	s.mu.Unlock()

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

// GetRecoveryStatus returns the current recovery session status
func (s *recoveryService) GetRecoveryStatus(ctx context.Context) (*RecoveryStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.currentStatus == nil {
		return &RecoveryStatus{
			InProgress: false,
		}, nil
	}

	// Check if expired
	if s.currentStatus.ExpiresAt != nil && time.Now().After(*s.currentStatus.ExpiresAt) {
		// Clear expired status
		go func() {
			s.mu.Lock()
			s.currentStatus = nil
			s.recoveryData = nil
			s.mu.Unlock()
		}()

		return &RecoveryStatus{
			InProgress: false,
		}, nil
	}

	// Return a copy of the status
	return &RecoveryStatus{
		InProgress: s.currentStatus.InProgress,
		SessionID:  s.currentStatus.SessionID,
		Email:      s.currentStatus.Email,
		Stage:      s.currentStatus.Stage,
		ExpiresAt:  s.currentStatus.ExpiresAt,
	}, nil
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
