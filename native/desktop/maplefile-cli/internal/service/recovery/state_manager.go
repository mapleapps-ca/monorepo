// internal/service/recovery/state_manager.go
package recovery

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
	uc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage"
)

const (
	recoveryStateKey = "current_recovery_state"
	recoveryDataKey  = "current_recovery_data"
)

// RecoveryStateManager handles persistent recovery state
type RecoveryStateManager interface {
	SaveState(ctx context.Context, status *RecoveryStatus) error
	LoadState(ctx context.Context) (*RecoveryStatus, error)
	ClearState(ctx context.Context) error
	FindActiveSession(ctx context.Context) (*RecoveryStatus, error)

	// Recovery data persistence methods
	SaveRecoveryData(ctx context.Context, data *uc_authdto.RecoveryData, recoveryToken string) error
	LoadRecoveryData(ctx context.Context) (*uc_authdto.RecoveryData, string, error)
	ClearRecoveryData(ctx context.Context) error
}

// PersistentRecoveryState represents the recovery state stored in the database
type PersistentRecoveryState struct {
	InProgress bool       `json:"in_progress"`
	SessionID  string     `json:"session_id,omitempty"`
	Email      string     `json:"email,omitempty"`
	Stage      string     `json:"stage,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	SavedAt    time.Time  `json:"saved_at"`
}

// PersistentRecoveryData represents the recovery data stored in the database
type PersistentRecoveryData struct {
	Email         string    `json:"email"`
	RecoveryToken string    `json:"recovery_token,omitempty"`
	MasterKey     string    `json:"master_key,omitempty"` // Base64 encoded for persistence
	SavedAt       time.Time `json:"saved_at"`
}

type recoveryStateManager struct {
	logger       *zap.Logger
	storage      storage.Storage
	recoveryRepo recovery.RecoveryRepository
}

// NewRecoveryStateManager creates a new recovery state manager
func NewRecoveryStateManager(
	logger *zap.Logger,
	storage storage.Storage,
	recoveryRepo recovery.RecoveryRepository,
) RecoveryStateManager {
	logger = logger.Named("RecoveryStateManager")
	return &recoveryStateManager{
		logger:       logger,
		storage:      storage,
		recoveryRepo: recoveryRepo,
	}
}

// SaveState saves the current recovery state to persistent storage
func (rsm *recoveryStateManager) SaveState(ctx context.Context, status *RecoveryStatus) error {
	if status == nil {
		return rsm.ClearState(ctx)
	}

	persistentState := &PersistentRecoveryState{
		InProgress: status.InProgress,
		SessionID:  status.SessionID,
		Email:      status.Email,
		Stage:      status.Stage,
		ExpiresAt:  status.ExpiresAt,
		SavedAt:    time.Now(),
	}

	data, err := json.Marshal(persistentState)
	if err != nil {
		rsm.logger.Error("Failed to marshal recovery state", zap.Error(err))
		return errors.NewAppError("failed to save recovery state", err)
	}

	if err := rsm.storage.Set(recoveryStateKey, data); err != nil {
		rsm.logger.Error("Failed to save recovery state to storage", zap.Error(err))
		return errors.NewAppError("failed to save recovery state", err)
	}

	rsm.logger.Debug("Successfully saved recovery state",
		zap.String("sessionID", status.SessionID),
		zap.String("stage", status.Stage))

	return nil
}

// LoadState loads the recovery state from persistent storage
func (rsm *recoveryStateManager) LoadState(ctx context.Context) (*RecoveryStatus, error) {
	data, err := rsm.storage.Get(recoveryStateKey)
	if err != nil {
		rsm.logger.Error("Failed to load recovery state from storage", zap.Error(err))
		return nil, errors.NewAppError("failed to load recovery state", err)
	}

	if data == nil {
		rsm.logger.Debug("No recovery state found in storage")
		return nil, nil
	}

	var persistentState PersistentRecoveryState
	if err := json.Unmarshal(data, &persistentState); err != nil {
		rsm.logger.Error("Failed to unmarshal recovery state", zap.Error(err))
		return nil, errors.NewAppError("failed to parse recovery state", err)
	}

	// Check if the state has expired
	if persistentState.ExpiresAt != nil && time.Now().After(*persistentState.ExpiresAt) {
		rsm.logger.Info("Loaded recovery state has expired, clearing it")
		_ = rsm.ClearState(ctx)
		return &RecoveryStatus{InProgress: false}, nil
	}

	status := &RecoveryStatus{
		InProgress: persistentState.InProgress,
		SessionID:  persistentState.SessionID,
		Email:      persistentState.Email,
		Stage:      persistentState.Stage,
		ExpiresAt:  persistentState.ExpiresAt,
	}

	rsm.logger.Debug("Successfully loaded recovery state",
		zap.String("sessionID", status.SessionID),
		zap.String("stage", status.Stage))

	return status, nil
}

// ClearState removes the recovery state from persistent storage
func (rsm *recoveryStateManager) ClearState(ctx context.Context) error {
	if err := rsm.storage.Delete(recoveryStateKey); err != nil {
		rsm.logger.Error("Failed to clear recovery state from storage", zap.Error(err))
		return errors.NewAppError("failed to clear recovery state", err)
	}

	rsm.logger.Debug("Successfully cleared recovery state")
	return nil
}

// SaveRecoveryData saves the recovery data to persistent storage
func (rsm *recoveryStateManager) SaveRecoveryData(ctx context.Context, data *uc_authdto.RecoveryData, recoveryToken string) error {
	if data == nil {
		return rsm.ClearRecoveryData(ctx)
	}

	// Encode master key as base64 for persistence
	var masterKeyB64 string
	if data.MasterKey != nil {
		masterKeyB64 = base64.StdEncoding.EncodeToString(data.MasterKey)
	}

	persistentData := &PersistentRecoveryData{
		Email:         data.Email,
		RecoveryToken: recoveryToken,
		MasterKey:     masterKeyB64,
		SavedAt:       time.Now(),
	}

	dataBytes, err := json.Marshal(persistentData)
	if err != nil {
		rsm.logger.Error("Failed to marshal recovery data", zap.Error(err))
		return errors.NewAppError("failed to save recovery data", err)
	}

	if err := rsm.storage.Set(recoveryDataKey, dataBytes); err != nil {
		rsm.logger.Error("Failed to save recovery data to storage", zap.Error(err))
		return errors.NewAppError("failed to save recovery data", err)
	}

	rsm.logger.Debug("Successfully saved recovery data")
	return nil
}

// LoadRecoveryData loads the recovery data from persistent storage
func (rsm *recoveryStateManager) LoadRecoveryData(ctx context.Context) (*uc_authdto.RecoveryData, string, error) {
	dataBytes, err := rsm.storage.Get(recoveryDataKey)
	if err != nil {
		rsm.logger.Error("Failed to load recovery data from storage", zap.Error(err))
		return nil, "", errors.NewAppError("failed to load recovery data", err)
	}

	if dataBytes == nil {
		rsm.logger.Debug("No recovery data found in storage")
		return nil, "", nil
	}

	var persistentData PersistentRecoveryData
	if err := json.Unmarshal(dataBytes, &persistentData); err != nil {
		rsm.logger.Error("Failed to unmarshal recovery data", zap.Error(err))
		return nil, "", errors.NewAppError("failed to parse recovery data", err)
	}

	// Decode master key from base64
	var masterKey []byte
	if persistentData.MasterKey != "" {
		masterKey, err = base64.StdEncoding.DecodeString(persistentData.MasterKey)
		if err != nil {
			rsm.logger.Error("Failed to decode master key", zap.Error(err))
			return nil, "", errors.NewAppError("failed to decode master key", err)
		}
	}

	data := &uc_authdto.RecoveryData{
		Email:     persistentData.Email,
		MasterKey: masterKey,
	}

	rsm.logger.Debug("Successfully loaded recovery data")
	return data, persistentData.RecoveryToken, nil
}

// ClearRecoveryData removes the recovery data from persistent storage
func (rsm *recoveryStateManager) ClearRecoveryData(ctx context.Context) error {
	if err := rsm.storage.Delete(recoveryDataKey); err != nil {
		rsm.logger.Error("Failed to clear recovery data from storage", zap.Error(err))
		return errors.NewAppError("failed to clear recovery data", err)
	}

	rsm.logger.Debug("Successfully cleared recovery data")
	return nil
}

// FindActiveSession searches for an active recovery session in the recovery repository
func (rsm *recoveryStateManager) FindActiveSession(ctx context.Context) (*RecoveryStatus, error) {
	rsm.logger.Debug("Searching for active recovery sessions")

	// This is a simplified implementation. In a production system, you might want
	// to index sessions by user or have a more efficient way to find active sessions.
	// For now, we'll rely on the persistent state storage as the primary mechanism.

	// Try to load from persistent state first
	status, err := rsm.LoadState(ctx)
	if err != nil {
		return nil, err
	}

	if status != nil && status.InProgress {
		// Verify the session still exists in the recovery repository
		if status.SessionID != "" {
			session, err := rsm.recoveryRepo.GetSessionBySessionIDString(ctx, status.SessionID)
			if err != nil {
				rsm.logger.Error("Failed to verify session in repository", zap.Error(err))
				// Clear invalid state
				_ = rsm.ClearState(ctx)
				return &RecoveryStatus{InProgress: false}, nil
			}

			if session == nil {
				rsm.logger.Info("Session not found in repository, clearing state")
				_ = rsm.ClearState(ctx)
				return &RecoveryStatus{InProgress: false}, nil
			}

			if session.IsExpired() {
				rsm.logger.Info("Session has expired, clearing state")
				_ = rsm.ClearState(ctx)
				return &RecoveryStatus{InProgress: false}, nil
			}

			// Update status based on actual session state
			status.Stage = session.GetState()
			if session.GetState() == "verified" {
				status.Stage = "verified"
			}

			rsm.logger.Debug("Found and verified active recovery session",
				zap.String("sessionID", status.SessionID),
				zap.String("stage", status.Stage))

			return status, nil
		}
	}

	return &RecoveryStatus{InProgress: false}, nil
}
