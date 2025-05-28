// internal/service/syncstate/reset.go
package syncstate

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

// ResetOutput represents the result of resetting sync state
type ResetOutput struct {
	Message string `json:"message"`
}

// ResetService defines the interface for resetting sync state
type ResetService interface {
	ResetSyncState(ctx context.Context) (*ResetOutput, error)
	ResetCollectionSync(ctx context.Context) (*ResetOutput, error)
	ResetFileSync(ctx context.Context) (*ResetOutput, error)
}

// resetService implements the ResetService interface
type resetService struct {
	logger        *zap.Logger
	syncStateRepo syncstate.SyncStateRepository
}

// NewResetService creates a new service for resetting sync state
func NewResetService(
	logger *zap.Logger,
	syncStateRepo syncstate.SyncStateRepository,
) ResetService {
	logger = logger.Named("ResetService")
	return &resetService{
		logger:        logger,
		syncStateRepo: syncStateRepo,
	}
}

// ResetSyncState resets the entire sync state to default values
func (s *resetService) ResetSyncState(ctx context.Context) (*ResetOutput, error) {
	s.logger.Info("Resetting sync state to default values")

	// Reset sync state using repository
	if err := s.syncStateRepo.ResetSyncState(ctx); err != nil {
		s.logger.Error("failed to reset sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to reset sync state", err)
	}

	s.logger.Info("Successfully reset sync state")

	return &ResetOutput{
		Message: "Sync state has been reset to default values. Next sync will be a full synchronization.",
	}, nil
}

// ResetCollectionSync resets only the collection sync state while preserving file sync state
func (s *resetService) ResetCollectionSync(ctx context.Context) (*ResetOutput, error) {
	s.logger.Info("Resetting collection sync state")

	// Get current sync state
	currentState, err := s.syncStateRepo.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("failed to get current sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to get current sync state", err)
	}

	// Create updated state with reset collection sync but preserved file sync
	updatedState := &syncstate.SyncState{
		LastCollectionSync: currentState.LastCollectionSync.Truncate(0), // Reset to zero time
		LastFileSync:       currentState.LastFileSync,                   // Preserve file sync
		LastFileID:         currentState.LastFileID,                     // Preserve file sync ID
	}

	// Save the updated state
	if err := s.syncStateRepo.SaveSyncState(ctx, updatedState); err != nil {
		s.logger.Error("failed to save sync state after collection reset", zap.Error(err))
		return nil, errors.NewAppError("failed to save sync state after collection reset", err)
	}

	s.logger.Info("Successfully reset collection sync state")

	return &ResetOutput{
		Message: "Collection sync state has been reset. Next collection sync will be a full synchronization.",
	}, nil
}

// ResetFileSync resets only the file sync state while preserving collection sync state
func (s *resetService) ResetFileSync(ctx context.Context) (*ResetOutput, error) {
	s.logger.Info("Resetting file sync state")

	// Get current sync state
	currentState, err := s.syncStateRepo.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("failed to get current sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to get current sync state", err)
	}

	// Create updated state with reset file sync but preserved collection sync
	updatedState := &syncstate.SyncState{
		LastCollectionSync: currentState.LastCollectionSync,       // Preserve collection sync
		LastFileSync:       currentState.LastFileSync.Truncate(0), // Reset to zero time
		LastCollectionID:   currentState.LastCollectionID,         // Preserve collection sync ID
	}

	// Save the updated state
	if err := s.syncStateRepo.SaveSyncState(ctx, updatedState); err != nil {
		s.logger.Error("failed to save sync state after file reset", zap.Error(err))
		return nil, errors.NewAppError("failed to save sync state after file reset", err)
	}

	s.logger.Info("Successfully reset file sync state")

	return &ResetOutput{
		Message: "File sync state has been reset. Next file sync will be a full synchronization.",
	}, nil
}
