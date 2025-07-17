// monorepo/cloud/mapleapps-backend/internal/iam/usecase/recovery/initiate.go
package recovery

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	dom_recovery "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/recovery"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/crypto"
)

type InitiateRecoveryUseCase interface {
	Execute(ctx context.Context, email string, method dom_recovery.RecoveryMethod) (*InitiateRecoveryResult, error)
}

type InitiateRecoveryResult struct {
	SessionID          string    `json:"session_id"`
	ChallengeID        string    `json:"challenge_id"`
	EncryptedChallenge string    `json:"encrypted_challenge"`
	CreatedAt          time.Time `json:"created_at"`
	ExpiresAt          time.Time `json:"expires_at"`
}

type initiateRecoveryUseCaseImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	recoveryRepo   dom_recovery.RecoveryRepository
	userRepo       dom_user.FederatedUserRepository
	maxAttempts    int
	attemptWindow  time.Duration
	sessionTimeout time.Duration
}

func NewInitiateRecoveryUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	recoveryRepo dom_recovery.RecoveryRepository,
	userRepo dom_user.FederatedUserRepository,
) InitiateRecoveryUseCase {
	logger = logger.Named("InitiateRecoveryUseCase")
	return &initiateRecoveryUseCaseImpl{
		config:         config,
		logger:         logger,
		recoveryRepo:   recoveryRepo,
		userRepo:       userRepo,
		maxAttempts:    5,                // Max 5 attempts
		attemptWindow:  15 * time.Minute, // Within 15 minutes
		sessionTimeout: 10 * time.Minute, // Session valid for 10 minutes
	}
}

func (uc *initiateRecoveryUseCaseImpl) Execute(ctx context.Context, email string, method dom_recovery.RecoveryMethod) (*InitiateRecoveryResult, error) {
	// Get client info from context
	ipAddress, _ := ctx.Value(constants.SessionIPAddress).(string)
	userAgent := "Unknown" // TODO: Get from request headers

	// Check rate limiting
	failedAttempts, err := uc.recoveryRepo.CountFailedAttemptsInWindow(ctx, email, uc.attemptWindow)
	if err != nil {
		uc.logger.Error("Failed to count recovery attempts", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to process recovery request")
	}

	if failedAttempts >= uc.maxAttempts {
		uc.logger.Warn("Too many recovery attempts",
			zap.String("email", email),
			zap.Int("attempts", failedAttempts))
		return nil, httperror.NewForBadRequestWithSingleField("email", "Too many recovery attempts. Please try again later.")
	}

	// Get user
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		uc.logger.Error("Failed to get user", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to process recovery request")
	}

	if user == nil {
		// Don't reveal if user exists - create a fake attempt for security
		uc.createFailedAttempt(ctx, gocql.UUID{}, email, method, ipAddress, userAgent, "User not found")
		return nil, httperror.NewForBadRequestWithSingleField("email", "If this email exists, recovery instructions will be sent.")
	}

	// Verify user has recovery key set up
	if user.SecurityData == nil || user.SecurityData.MasterKeyEncryptedWithRecoveryKey.Ciphertext == nil {
		uc.createFailedAttempt(ctx, user.ID, email, method, ipAddress, userAgent, "No recovery key configured")
		return nil, httperror.NewForBadRequestWithSingleField("recovery", "Recovery key not configured for this account")
	}

	// Create recovery attempt
	attempt := &dom_recovery.RecoveryAttempt{
		ID:          gocql.TimeUUID(),
		UserID:      user.ID,
		Email:       email,
		Method:      method,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Status:      "initiated",
		AttemptedAt: time.Now(),
		ExpiresAt:   time.Now().Add(uc.sessionTimeout),
	}

	if err := uc.recoveryRepo.CreateRecoveryAttempt(ctx, attempt); err != nil {
		uc.logger.Error("Failed to create recovery attempt", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to initiate recovery")
	}

	// Generate challenge
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		uc.logger.Error("Failed to generate challenge", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to generate security challenge")
	}

	// Encrypt challenge with user's public key
	encryptedChallenge, err := crypto.EncryptWithPublicKey(challenge, user.SecurityData.PublicKey.Key)
	if err != nil {
		uc.logger.Error("Failed to encrypt challenge", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to prepare security challenge")
	}

	// Create recovery session
	sessionID := gocql.TimeUUID().String()
	challengeID := gocql.TimeUUID().String()

	session := &dom_recovery.RecoverySession{
		SessionID:                sessionID,
		UserID:                   user.ID,
		Email:                    email,
		Method:                   method,
		EncryptedChallenge:       challenge, // Store original challenge for verification
		ChallengeID:              challengeID,
		PublicKey:                user.SecurityData.PublicKey.Key,
		EncryptedMasterKey:       append(user.SecurityData.EncryptedMasterKey.Nonce, user.SecurityData.EncryptedMasterKey.Ciphertext...),
		EncryptedPrivateKey:      append(user.SecurityData.EncryptedPrivateKey.Nonce, user.SecurityData.EncryptedPrivateKey.Ciphertext...),
		MasterKeyWithRecoveryKey: append(user.SecurityData.MasterKeyEncryptedWithRecoveryKey.Nonce, user.SecurityData.MasterKeyEncryptedWithRecoveryKey.Ciphertext...),
		CreatedAt:                time.Now(),
		ExpiresAt:                time.Now().Add(uc.sessionTimeout),
		IsVerified:               false,
	}

	if err := uc.recoveryRepo.CreateRecoverySession(ctx, session); err != nil {
		uc.logger.Error("Failed to create recovery session", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to create recovery session")
	}

	// Update attempt with challenge ID
	attempt.ChallengeID = challengeID
	attempt.Status = "challenged"
	if err := uc.recoveryRepo.UpdateRecoveryAttempt(ctx, attempt); err != nil {
		uc.logger.Warn("Failed to update recovery attempt with challenge", zap.Error(err))
		// Continue anyway
	}

	return &InitiateRecoveryResult{
		SessionID:          sessionID,
		ChallengeID:        challengeID,
		EncryptedChallenge: base64.StdEncoding.EncodeToString(encryptedChallenge),
		CreatedAt:          session.CreatedAt,
		ExpiresAt:          session.ExpiresAt,
	}, nil
}

func (uc *initiateRecoveryUseCaseImpl) createFailedAttempt(ctx context.Context, userID gocql.UUID, email string, method dom_recovery.RecoveryMethod, ipAddress, userAgent, reason string) {
	now := time.Now()
	attempt := &dom_recovery.RecoveryAttempt{
		ID:            gocql.TimeUUID(),
		UserID:        userID,
		Email:         email,
		Method:        method,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Status:        "failed",
		FailureReason: reason,
		AttemptedAt:   now,
		CompletedAt:   &now,
		ExpiresAt:     time.Now().Add(24 * time.Hour), // Keep for audit
	}

	if err := uc.recoveryRepo.CreateRecoveryAttempt(ctx, attempt); err != nil {
		uc.logger.Warn("Failed to create failed recovery attempt", zap.Error(err))
	}
}
