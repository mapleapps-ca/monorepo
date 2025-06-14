// native/desktop/maplefile-cli/internal/usecase/recovery/initiate.go
package recovery

import (
	"context"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recoverydto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// InitiateRecoveryUseCase defines the interface for initiating account recovery
type InitiateRecoveryUseCase interface {
	Execute(ctx context.Context, email string, method string) (*recoverydto.RecoveryInitiateResponseDTO, error)
}

// initiateRecoveryUseCase implements the InitiateRecoveryUseCase interface
type initiateRecoveryUseCase struct {
	logger          *zap.Logger
	recoveryDTORepo recoverydto.RecoveryDTORepository
	recoveryRepo    recovery.RecoveryRepository
	userRepo        user.Repository
}

// NewInitiateRecoveryUseCase creates a new initiate recovery use case
func NewInitiateRecoveryUseCase(
	logger *zap.Logger,
	recoveryDTORepo recoverydto.RecoveryDTORepository,
	recoveryRepo recovery.RecoveryRepository,
	userRepo user.Repository,
) InitiateRecoveryUseCase {
	logger = logger.Named("InitiateRecoveryUseCase")
	return &initiateRecoveryUseCase{
		logger:          logger,
		recoveryDTORepo: recoveryDTORepo,
		recoveryRepo:    recoveryRepo,
		userRepo:        userRepo,
	}
}

// Execute initiates the recovery process
func (uc *initiateRecoveryUseCase) Execute(ctx context.Context, email string, method string) (*recoverydto.RecoveryInitiateResponseDTO, error) {
	//
	// STEP 1: Validate inputs
	//
	if email == "" {
		return nil, errors.NewAppError("email is required", nil)
	}
	if method == "" {
		method = recovery.DefaultRecoveryMethod
	}

	// Sanitize inputs
	email = strings.ToLower(strings.TrimSpace(email))

	//
	// STEP 2: Check if user exists locally
	//
	existingUser, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		uc.logger.Error("Failed to check user existence", zap.String("email", email), zap.Error(err))
		return nil, errors.NewAppError("failed to check user existence", err)
	}

	// If user doesn't exist locally, we still proceed with cloud recovery
	// This allows recovery for users who may have lost their local data
	if existingUser == nil {
		uc.logger.Info("User not found locally, proceeding with cloud recovery", zap.String("email", email))
	}

	//
	// STEP 3: Create recovery request
	//
	request := &recoverydto.RecoveryInitiateRequestDTO{
		Email:  email,
		Method: method,
	}

	//
	// STEP 4: Call cloud service to initiate recovery
	//
	uc.logger.Debug("Initiating recovery from cloud",
		zap.String("email", email),
		zap.String("method", method))

	response, err := uc.recoveryDTORepo.InitiateRecoveryFromCloud(ctx, request)
	if err != nil {
		uc.logger.Error("Failed to initiate recovery from cloud", zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Store recovery session locally for tracking
	//
	sessionUUID, err := gocql.ParseUUID(response.SessionID)
	if err != nil {
		uc.logger.Error("Invalid session ID from cloud", zap.String("sessionID", response.SessionID), zap.Error(err))
		return nil, errors.NewAppError("invalid session ID from server", err)
	}

	challengeUUID, err := gocql.ParseUUID(response.ChallengeID)
	if err != nil {
		uc.logger.Error("Invalid challenge ID from cloud", zap.String("challengeID", response.ChallengeID), zap.Error(err))
		return nil, errors.NewAppError("invalid challenge ID from server", err)
	}

	// Create local recovery session
	localSession := &recovery.RecoverySession{
		SessionID:          sessionUUID,
		ChallengeID:        challengeUUID,
		Email:              email,
		UserID:             gocql.TimeUUID(),                    // Generate a temporary UUID if user doesn't exist
		EncryptedChallenge: []byte(response.EncryptedChallenge), // Store as reference
		ExpiresAt:          time.Now().Add(time.Duration(response.ExpiresIn) * time.Second),
		IsVerified:         false,
		CreatedAt:          time.Now(),
	}

	// If we have a local user, use their actual user ID
	if existingUser != nil {
		localSession.UserID = existingUser.ID
	}

	// Save session locally
	if err := uc.recoveryRepo.CreateSession(ctx, localSession); err != nil {
		uc.logger.Error("Failed to save recovery session locally", zap.Error(err))
		// Continue anyway - local storage failure shouldn't block recovery
	}

	uc.logger.Info("Successfully initiated recovery",
		zap.String("email", email),
		zap.String("sessionID", response.SessionID),
		zap.Int("expiresIn", response.ExpiresIn))

	return response, nil
}
