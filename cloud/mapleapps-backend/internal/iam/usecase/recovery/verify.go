// monorepo/cloud/mapleapps-backend/internal/iam/usecase/recovery/verify.go
package recovery

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	dom_recovery "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/recovery"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type VerifyRecoveryUseCase interface {
	Execute(ctx context.Context, req *VerifyRecoveryRequest) (*VerifyRecoveryResult, error)
}

type VerifyRecoveryRequest struct {
	SessionID          string `json:"session_id"`
	DecryptedChallenge string `json:"decrypted_challenge"`
}

type VerifyRecoveryResult struct {
	UserID                   string    `json:"user_id"`
	Email                    string    `json:"email"`
	RecoveryToken            string    `json:"recovery_token"`
	MasterKeyWithRecoveryKey string    `json:"master_key_encrypted_with_recovery_key"`
	ExpiresAt                time.Time `json:"expires_at"`
}

type verifyRecoveryUseCaseImpl struct {
	config       *config.Configuration
	logger       *zap.Logger
	recoveryRepo dom_recovery.RecoveryRepository
	userRepo     dom_user.FederatedUserRepository
}

func NewVerifyRecoveryUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	recoveryRepo dom_recovery.RecoveryRepository,
	userRepo dom_user.FederatedUserRepository,
) VerifyRecoveryUseCase {
	logger = logger.Named("VerifyRecoveryUseCase")
	return &verifyRecoveryUseCaseImpl{
		config:       config,
		logger:       logger,
		recoveryRepo: recoveryRepo,
		userRepo:     userRepo,
	}
}

func (uc *verifyRecoveryUseCaseImpl) Execute(ctx context.Context, req *VerifyRecoveryRequest) (*VerifyRecoveryResult, error) {
	// Validate input
	if req.SessionID == "" {
		return nil, httperror.NewForBadRequestWithSingleField("session_id", "Session ID is required")
	}
	if req.DecryptedChallenge == "" {
		return nil, httperror.NewForBadRequestWithSingleField("decrypted_challenge", "Decrypted challenge is required")
	}

	// Get recovery session
	session, err := uc.recoveryRepo.GetRecoverySessionByID(ctx, req.SessionID)
	if err != nil {
		uc.logger.Error("Failed to get recovery session", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to verify recovery")
	}

	if session == nil {
		return nil, httperror.NewForBadRequestWithSingleField("session_id", "Invalid or expired session")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		uc.updateFailedAttempt(ctx, session, "Session expired")
		return nil, httperror.NewForBadRequestWithSingleField("session_id", "Session has expired")
	}

	// Check if already verified
	if session.IsVerified {
		return nil, httperror.NewForBadRequestWithSingleField("session_id", "Session already verified")
	}

	// Decode the decrypted challenge
	decryptedChallengeBytes, err := base64.RawURLEncoding.DecodeString(req.DecryptedChallenge)
	if err != nil {
		// Try standard encoding as fallback
		decryptedChallengeBytes, err = base64.StdEncoding.DecodeString(req.DecryptedChallenge)
		if err != nil {
			uc.logger.Warn("Failed to decode challenge", zap.Error(err))
			uc.updateFailedAttempt(ctx, session, "Invalid challenge format")
			return nil, httperror.NewForBadRequestWithSingleField("decrypted_challenge", "Invalid challenge format")
		}
	}

	// Verify the challenge matches
	if !bytes.Equal(session.EncryptedChallenge, decryptedChallengeBytes) {
		uc.logger.Warn("Challenge verification failed",
			zap.String("session_id", req.SessionID),
			zap.String("email", session.Email))
		uc.updateFailedAttempt(ctx, session, "Challenge verification failed")
		return nil, httperror.NewForBadRequestWithSingleField("decrypted_challenge", "Invalid challenge response")
	}

	uc.logger.Info("Recovery challenge verified successfully",
		zap.String("session_id", req.SessionID),
		zap.String("user_id", session.UserID.String()))

	// Mark session as verified
	session.IsVerified = true
	now := time.Now()
	session.VerifiedAt = &now

	if err := uc.recoveryRepo.UpdateRecoverySession(ctx, session); err != nil {
		uc.logger.Error("Failed to update recovery session", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to complete verification")
	}

	// Update recovery attempt as successful
	uc.updateSuccessfulAttempt(ctx, session)

	// Generate a recovery token for the next step
	recoveryToken := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", session.SessionID, session.ChallengeID)))

	// Return the encrypted master key (encrypted with recovery key)
	masterKeyWithRecoveryKeyBase64 := base64.StdEncoding.EncodeToString(session.MasterKeyWithRecoveryKey)

	return &VerifyRecoveryResult{
		UserID:                   session.UserID.String(),
		Email:                    session.Email,
		RecoveryToken:            recoveryToken,
		MasterKeyWithRecoveryKey: masterKeyWithRecoveryKeyBase64,
		ExpiresAt:                session.ExpiresAt,
	}, nil
}

func (uc *verifyRecoveryUseCaseImpl) updateFailedAttempt(ctx context.Context, session *dom_recovery.RecoverySession, reason string) {
	// Get the recovery attempt
	attempts, err := uc.recoveryRepo.GetRecentRecoveryAttempts(ctx, session.UserID, 10)
	if err != nil {
		uc.logger.Warn("Failed to get recovery attempts", zap.Error(err))
		return
	}

	// Find the attempt with matching challenge ID
	for _, attempt := range attempts {
		if attempt.ChallengeID == session.ChallengeID {
			attempt.Status = "failed"
			attempt.FailureReason = reason
			now := time.Now()
			attempt.CompletedAt = &now

			if err := uc.recoveryRepo.UpdateRecoveryAttempt(ctx, attempt); err != nil {
				uc.logger.Warn("Failed to update recovery attempt", zap.Error(err))
			}
			break
		}
	}
}

func (uc *verifyRecoveryUseCaseImpl) updateSuccessfulAttempt(ctx context.Context, session *dom_recovery.RecoverySession) {
	// Get the recovery attempt
	attempts, err := uc.recoveryRepo.GetRecentRecoveryAttempts(ctx, session.UserID, 10)
	if err != nil {
		uc.logger.Warn("Failed to get recovery attempts", zap.Error(err))
		return
	}

	// Find the attempt with matching challenge ID
	for _, attempt := range attempts {
		if attempt.ChallengeID == session.ChallengeID {
			attempt.Status = "succeeded"
			now := time.Now()
			attempt.CompletedAt = &now

			if err := uc.recoveryRepo.UpdateRecoveryAttempt(ctx, attempt); err != nil {
				uc.logger.Warn("Failed to update recovery attempt", zap.Error(err))
			}
			break
		}
	}
}
