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
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
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
	userRepo        user.Repository // Added user repository
}

// NewVerifyRecoveryUseCase creates a new verify recovery use case
func NewVerifyRecoveryUseCase(
	logger *zap.Logger,
	recoveryDTORepo recoverydto.RecoveryDTORepository,
	recoveryRepo recovery.RecoveryRepository,
	userRepo user.Repository, // Added user repository parameter
) VerifyRecoveryUseCase {
	logger = logger.Named("VerifyRecoveryUseCase")
	return &verifyRecoveryUseCase{
		logger:          logger,
		recoveryDTORepo: recoveryDTORepo,
		recoveryRepo:    recoveryRepo,
		userRepo:        userRepo, // Store user repository
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
	// STEP 2: Get local recovery session
	//
	sessionUUID, err := gocql.ParseUUID(sessionID)
	if err != nil {
		return nil, errors.NewAppError("invalid session ID format", err)
	}

	localSession, err := uc.recoveryRepo.GetSessionByID(ctx, sessionUUID)
	if err != nil {
		uc.logger.Error("Failed to get recovery session", zap.Error(err))
		return nil, errors.NewAppError("failed to get recovery session", err)
	}

	if localSession == nil {
		return nil, errors.NewAppError("recovery session not found", nil)
	}

	// Check if session is expired
	if localSession.IsExpired() {
		uc.logger.Warn("Recovery session has expired", zap.String("sessionID", sessionID))
		return nil, errors.NewAppError("recovery session has expired", nil)
	}

	//
	// STEP 3: Get user data to access encrypted keys
	//
	user, err := uc.userRepo.GetByEmail(ctx, localSession.Email)
	if err != nil {
		uc.logger.Error("Failed to get user", zap.Error(err))
		return nil, errors.NewAppError("failed to get user data", err)
	}

	if user == nil {
		return nil, errors.NewAppError("user not found locally. Please ensure you have logged in before attempting recovery.", nil)
	}

	//
	// STEP 4: Decode and validate recovery key
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
	// STEP 5: Decrypt master key using recovery key
	//
	if len(user.MasterKeyEncryptedWithRecoveryKey.Ciphertext) == 0 {
		return nil, errors.NewAppError("no recovery key data found for user", nil)
	}

	masterKey, err := crypto.DecryptWithSecretBox(
		user.MasterKeyEncryptedWithRecoveryKey.Ciphertext,
		user.MasterKeyEncryptedWithRecoveryKey.Nonce,
		recoveryKeyBytes,
	)
	if err != nil {
		uc.logger.Error("Failed to decrypt master key with recovery key", zap.Error(err))
		return nil, errors.NewAppError("invalid recovery key", nil)
	}

	// Clear master key after use
	defer crypto.ClearBytes(masterKey)

	//
	// STEP 6: Decrypt private key using master key
	//
	privateKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedPrivateKey.Ciphertext,
		user.EncryptedPrivateKey.Nonce,
		masterKey,
	)
	if err != nil {
		uc.logger.Error("Failed to decrypt private key", zap.Error(err))
		return nil, errors.NewAppError("failed to decrypt private key", err)
	}

	// Clear private key after use
	defer crypto.ClearBytes(privateKey)

	//
	// STEP 7: Decrypt the challenge using the private key
	//
	if len(localSession.EncryptedChallenge) == 0 {
		return nil, errors.NewAppError("no encrypted challenge found in session", nil)
	}

	// The encrypted challenge from the server should be base64 encoded
	encryptedChallengeStr := string(localSession.EncryptedChallenge)

	// Decode the base64 encrypted challenge
	encryptedChallengeBytes, err := base64.StdEncoding.DecodeString(encryptedChallengeStr)
	if err != nil {
		// Try URL-safe base64
		encryptedChallengeBytes, err = base64.RawURLEncoding.DecodeString(encryptedChallengeStr)
		if err != nil {
			return nil, errors.NewAppError("invalid encrypted challenge format", err)
		}
	}

	// Decrypt the challenge using box_open (public key cryptography)
	// The challenge is encrypted with the user's public key, so we need the private key
	decryptedChallenge, err := crypto.DecryptWithBoxAnonymous(
		encryptedChallengeBytes,
		user.PublicKey.Key,
		privateKey,
	)
	if err != nil {
		uc.logger.Error("Failed to decrypt challenge with private key", zap.Error(err))
		return nil, errors.NewAppError("failed to decrypt challenge", err)
	}

	// Encode the decrypted challenge as base64 for transmission
	decryptedChallengeBase64 := base64.StdEncoding.EncodeToString(decryptedChallenge)

	uc.logger.Debug("Successfully decrypted challenge",
		zap.Int("challengeLength", len(decryptedChallenge)))

	//
	// STEP 8: Create verify request with the properly decrypted challenge
	//
	request := &recoverydto.RecoveryVerifyRequestDTO{
		SessionID:          sessionID,
		DecryptedChallenge: decryptedChallengeBase64,
	}

	//
	// STEP 9: Call cloud service to verify recovery
	//
	uc.logger.Debug("Verifying recovery with cloud", zap.String("sessionID", sessionID))

	response, err := uc.recoveryDTORepo.VerifyRecoveryFromCloud(ctx, request)
	if err != nil {
		uc.logger.Error("Failed to verify recovery with cloud", zap.Error(err))
		return nil, err
	}

	//
	// STEP 10: Update local session as verified
	//
	now := time.Now()
	localSession.IsVerified = true
	localSession.VerifiedAt = &now

	if err := uc.recoveryRepo.UpdateSession(ctx, localSession); err != nil {
		uc.logger.Error("Failed to update local recovery session", zap.Error(err))
		// Continue anyway
	}

	//
	// STEP 11: Create recovery token record locally
	//
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

	uc.logger.Info("Successfully verified recovery challenge",
		zap.String("sessionID", sessionID),
		zap.Int("expiresIn", response.ExpiresIn))

	return response, nil
}
