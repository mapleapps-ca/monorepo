// native/desktop/maplefile-cli/internal/usecase/recovery/verify.go
package recovery

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recoverydto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// VerifyRecoveryUseCase defines the interface for verifying recovery challenge
type VerifyRecoveryUseCase interface {
	Execute(ctx context.Context, sessionID string, recoveryKey string) (*recoverydto.RecoveryVerifyResponseDTO, error)
}

// verifyRecoveryUseCase implements the VerifyRecoveryUseCase interface
type verifyRecoveryUseCase struct {
	logger          *zap.Logger
	recoveryDTORepo recoverydto.RecoveryDTORepository
	recoveryRepo    recovery.RecoveryRepository
}

// NewVerifyRecoveryUseCase creates a new verify recovery use case
func NewVerifyRecoveryUseCase(
	logger *zap.Logger,
	recoveryDTORepo recoverydto.RecoveryDTORepository,
	recoveryRepo recovery.RecoveryRepository,
) VerifyRecoveryUseCase {
	logger = logger.Named("VerifyRecoveryUseCase")
	return &verifyRecoveryUseCase{
		logger:          logger,
		recoveryDTORepo: recoveryDTORepo,
		recoveryRepo:    recoveryRepo,
	}
}

// Execute verifies the recovery challenge with the provided recovery key
func (uc *verifyRecoveryUseCase) Execute(ctx context.Context, sessionID string, recoveryKey string) (*recoverydto.RecoveryVerifyResponseDTO, error) {
	//
	// STEP 1: Validate inputs
	//
	if sessionID == "" {
		return nil, errors.NewAppError("session ID is required", nil)
	}
	if recoveryKey == "" {
		return nil, errors.NewAppError("recovery key is required", nil)
	}

	// Sanitize inputs
	sessionID = strings.TrimSpace(sessionID)
	recoveryKey = strings.TrimSpace(recoveryKey)

	//
	// STEP 2: Get local recovery session (if exists)
	//
	var localSession *recovery.RecoverySession
	sessionUUID, err := gocql.ParseUUID(sessionID)
	if err == nil {
		localSession, _ = uc.recoveryRepo.GetSessionByID(ctx, sessionUUID)
	}

	if localSession != nil {
		// Check if session is expired
		if localSession.IsExpired() {
			uc.logger.Warn("Recovery session has expired", zap.String("sessionID", sessionID))
			// Continue anyway - let server validate
		}

		// Check if already verified
		if localSession.IsVerified {
			uc.logger.Info("Recovery session already verified", zap.String("sessionID", sessionID))
			// Continue anyway - server is source of truth
		}
	}

	//
	// STEP 3: Decode recovery key
	//
	recoveryKeyBytes, err := base64.StdEncoding.DecodeString(recoveryKey)
	if err != nil {
		// Try URL-safe base64 as fallback
		recoveryKeyBytes, err = base64.RawURLEncoding.DecodeString(recoveryKey)
		if err != nil {
			return nil, errors.NewAppError("invalid recovery key format", err)
		}
	}

	// Validate recovery key size
	if len(recoveryKeyBytes) != crypto.RecoveryKeySize {
		return nil, errors.NewAppError("invalid recovery key size", nil)
	}

	//
	// STEP 4: Get encrypted challenge from local session or initiate response
	//
	var encryptedChallenge []byte
	if localSession != nil && len(localSession.EncryptedChallenge) > 0 {
		encryptedChallenge = localSession.EncryptedChallenge
	}

	// If we don't have the encrypted challenge locally, we need to get it from the server
	// In this case, we'll need to decrypt it after getting the response
	if len(encryptedChallenge) == 0 {
		uc.logger.Warn("No encrypted challenge found locally, will rely on server validation",
			zap.String("sessionID", sessionID))
	}

	//
	// STEP 5: Attempt to decrypt challenge locally (if we have it)
	//
	// var decryptedChallenge []byte
	if len(encryptedChallenge) > 0 {
		// The challenge is encrypted with the user's public key
		// We need the recovery key to decrypt the master key first
		// But at this stage, we don't have the master key yet
		// So we'll send a placeholder and let the server validate
		uc.logger.Debug("Have encrypted challenge, but need server to validate with recovery key")
	}

	// For now, we'll send the recovery key hash as the decrypted challenge
	// The server will validate if this recovery key can decrypt the actual challenge
	challengeData := crypto.EncodeToBase64(recoveryKeyBytes)

	//
	// STEP 6: Create verify request
	//
	request := &recoverydto.RecoveryVerifyRequestDTO{
		SessionID:          sessionID,
		DecryptedChallenge: challengeData,
	}

	//
	// STEP 7: Call cloud service to verify recovery
	//
	uc.logger.Debug("Verifying recovery with cloud", zap.String("sessionID", sessionID))

	response, err := uc.recoveryDTORepo.VerifyRecoveryFromCloud(ctx, request)
	if err != nil {
		uc.logger.Error("Failed to verify recovery with cloud", zap.Error(err))
		return nil, err
	}

	//
	// STEP 8: Update local session as verified
	//
	if localSession != nil {
		now := time.Now()
		localSession.IsVerified = true
		localSession.VerifiedAt = &now

		if err := uc.recoveryRepo.UpdateSession(ctx, localSession); err != nil {
			uc.logger.Error("Failed to update local recovery session", zap.Error(err))
			// Continue anyway
		}
	}

	//
	// STEP 9: Create recovery token record locally
	//
	if localSession != nil {
		token := &recovery.RecoveryToken{
			Token:     response.RecoveryToken,
			SessionID: localSession.SessionID,
			UserID:    localSession.UserID,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Duration(response.ExpiresIn) * time.Second),
			Used:      false,
		}

		if err := uc.recoveryRepo.CreateToken(ctx, token); err != nil {
			uc.logger.Error("Failed to save recovery token locally", zap.Error(err))
			// Continue anyway
		}
	}

	uc.logger.Info("Successfully verified recovery challenge",
		zap.String("sessionID", sessionID),
		zap.Int("expiresIn", response.ExpiresIn))

	return response, nil
}
