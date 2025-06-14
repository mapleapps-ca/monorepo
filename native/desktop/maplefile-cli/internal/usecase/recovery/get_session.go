// native/desktop/maplefile-cli/internal/usecase/recovery/get_session.go
package recovery

import (
	"context"
	"strings"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
)

// GetRecoverySessionUseCase defines the interface for getting a recovery session
type GetRecoverySessionUseCase interface {
	Execute(ctx context.Context, sessionID string) (*recovery.RecoverySession, error)
}

// getRecoverySessionUseCase implements the GetRecoverySessionUseCase interface
type getRecoverySessionUseCase struct {
	logger       *zap.Logger
	recoveryRepo recovery.RecoveryRepository
}

// NewGetRecoverySessionUseCase creates a new get recovery session use case
func NewGetRecoverySessionUseCase(
	logger *zap.Logger,
	recoveryRepo recovery.RecoveryRepository,
) GetRecoverySessionUseCase {
	logger = logger.Named("GetRecoverySessionUseCase")
	return &getRecoverySessionUseCase{
		logger:       logger,
		recoveryRepo: recoveryRepo,
	}
}

// Execute retrieves a recovery session by ID
func (uc *getRecoverySessionUseCase) Execute(ctx context.Context, sessionID string) (*recovery.RecoverySession, error) {
	//
	// STEP 1: Validate input
	//
	if sessionID == "" {
		return nil, errors.NewAppError("session ID is required", nil)
	}

	sessionID = strings.TrimSpace(sessionID)

	//
	// STEP 2: Parse session ID
	//
	sessionUUID, err := gocql.ParseUUID(sessionID)
	if err != nil {
		uc.logger.Error("Invalid session ID format",
			zap.String("sessionID", sessionID),
			zap.Error(err))
		return nil, errors.NewAppError("invalid session ID format", err)
	}

	//
	// STEP 3: Get session from repository
	//
	uc.logger.Debug("Getting recovery session", zap.String("sessionID", sessionID))

	session, err := uc.recoveryRepo.GetSessionByID(ctx, sessionUUID)
	if err != nil {
		uc.logger.Error("Failed to get recovery session", zap.Error(err))
		return nil, errors.NewAppError("failed to get recovery session", err)
	}

	//
	// STEP 4: Check if session exists
	//
	if session == nil {
		uc.logger.Warn("Recovery session not found", zap.String("sessionID", sessionID))
		return nil, recovery.NewSessionNotFoundError(sessionID)
	}

	//
	// STEP 5: Log session state
	//
	uc.logger.Info("Retrieved recovery session",
		zap.String("sessionID", sessionID),
		zap.String("email", session.Email),
		zap.String("state", session.GetState()),
		zap.Bool("isExpired", session.IsExpired()),
		zap.Bool("isVerified", session.IsVerified))

	return session, nil
}
