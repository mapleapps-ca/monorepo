// internal/service/syncstate/save.go
package syncstate

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncstate"
)

// SaveInput represents the input for saving sync state
type SaveInput struct {
	LastCollectionSync *time.Time  `json:"last_collection_sync,omitempty"`
	LastFileSync       *time.Time  `json:"last_file_sync,omitempty"`
	LastCollectionID   *gocql.UUID `json:"last_collection_id,omitempty"`
	LastFileID         *gocql.UUID `json:"last_file_id,omitempty"`
}

// SaveOutput represents the result of saving sync state
type SaveOutput struct {
	SyncState *syncstate.SyncState `json:"sync_state"`
	Message   string               `json:"message"`
}

// SaveService defines the interface for saving sync state
type SaveService interface {
	SaveSyncState(ctx context.Context, input *SaveInput) (*SaveOutput, error)
	UpdateCollectionSync(ctx context.Context, timestamp time.Time, lastID gocql.UUID) (*SaveOutput, error)
	UpdateFileSync(ctx context.Context, timestamp time.Time, lastID gocql.UUID) (*SaveOutput, error)
}

// saveService implements the SaveService interface
type saveService struct {
	logger        *zap.Logger
	syncStateRepo syncstate.SyncStateRepository
}

// NewSaveService creates a new service for saving sync state
func NewSaveService(
	logger *zap.Logger,
	syncStateRepo syncstate.SyncStateRepository,
) SaveService {
	logger = logger.Named("SaveService")
	return &saveService{
		logger:        logger,
		syncStateRepo: syncStateRepo,
	}
}

// SaveSyncState saves or updates the sync state
func (s *saveService) SaveSyncState(ctx context.Context, input *SaveInput) (*SaveOutput, error) {
	if input == nil {
		s.logger.Error("❌ input is required")
		return nil, errors.NewAppError("input is required", nil)
	}

	s.logger.Debug("🔄 Saving sync state", zap.Any("input", input))

	// Get current sync state first
	currentState, err := s.syncStateRepo.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("❌ failed to get current sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to get current sync state", err)
	}

	// Update only the fields that are provided
	updatedState := &syncstate.SyncState{
		LastCollectionSync: currentState.LastCollectionSync,
		LastFileSync:       currentState.LastFileSync,
		LastCollectionID:   currentState.LastCollectionID,
		LastFileID:         currentState.LastFileID,
	}

	if input.LastCollectionSync != nil {
		updatedState.LastCollectionSync = *input.LastCollectionSync
	}
	if input.LastFileSync != nil {
		updatedState.LastFileSync = *input.LastFileSync
	}
	if input.LastCollectionID != nil {
		updatedState.LastCollectionID = *input.LastCollectionID
	}
	if input.LastFileID != nil {
		updatedState.LastFileID = *input.LastFileID
	}

	// Save the updated state
	if err := s.syncStateRepo.SaveSyncState(ctx, updatedState); err != nil {
		s.logger.Error("❌ failed to save sync state", zap.Error(err))
		return nil, errors.NewAppError("failed to save sync state", err)
	}

	s.logger.Info("✅ Successfully saved sync state",
		zap.Time("last_collection_sync", updatedState.LastCollectionSync),
		zap.Time("last_file_sync", updatedState.LastFileSync))

	return &SaveOutput{
		SyncState: updatedState,
		Message:   "Sync state saved successfully",
	}, nil
}

// UpdateCollectionSync updates only the collection sync timestamp and ID
func (s *saveService) UpdateCollectionSync(ctx context.Context, timestamp time.Time, lastID gocql.UUID) (*SaveOutput, error) {
	s.logger.Debug("🔄 Updating collection sync state",
		zap.Time("timestamp", timestamp),
		zap.String("lastID", lastID.Hex()))

	input := &SaveInput{
		LastCollectionSync: &timestamp,
		LastCollectionID:   &lastID,
	}

	return s.SaveSyncState(ctx, input)
}

// UpdateFileSync updates only the file sync timestamp and ID
func (s *saveService) UpdateFileSync(ctx context.Context, timestamp time.Time, lastID gocql.UUID) (*SaveOutput, error) {
	s.logger.Debug("🔄 Updating file sync state",
		zap.Time("timestamp", timestamp),
		zap.String("lastID", lastID.Hex()))

	input := &SaveInput{
		LastFileSync: &timestamp,
		LastFileID:   &lastID,
	}

	return s.SaveSyncState(ctx, input)
}
