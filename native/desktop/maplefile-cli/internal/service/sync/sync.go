// native/desktop/maplefile-cli/internal/service/sync/sync.go
package sync

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/sync"
	uc_sync "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/sync"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// SyncService defines the interface for sync operations
type SyncService interface {
	SyncCollections(ctx context.Context) (*sync.SyncResult, error)
	SyncFiles(ctx context.Context) (*sync.SyncResult, error)
	FullSync(ctx context.Context) (*sync.SyncResult, error)
	ResetSync(ctx context.Context) error
}

// syncService implements the SyncService interface
type syncService struct {
	logger                     *zap.Logger
	syncUseCase                uc_sync.SyncUseCase
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
}

// NewSyncService creates a new service for sync operations
func NewSyncService(
	logger *zap.Logger,
	syncUseCase uc_sync.SyncUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
) SyncService {
	return &syncService{
		logger:                     logger,
		syncUseCase:                syncUseCase,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
	}
}

func (s *syncService) SyncCollections(ctx context.Context) (*sync.SyncResult, error) {
	// Verify user is logged in
	if err := s.ensureUserLoggedIn(ctx); err != nil {
		return nil, err
	}

	s.logger.Info("Starting collection sync from service layer")

	result, err := s.syncUseCase.SyncCollections(ctx)
	if err != nil {
		s.logger.Error("Collection sync failed", zap.Error(err))
		return nil, errors.NewAppError("collection sync failed", err)
	}

	s.logger.Info("Collection sync completed successfully",
		zap.Int("collections_processed", result.CollectionsProcessed),
		zap.Int("collections_added", result.CollectionsAdded),
		zap.Int("collections_updated", result.CollectionsUpdated),
		zap.Int("collections_deleted", result.CollectionsDeleted),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}

func (s *syncService) SyncFiles(ctx context.Context) (*sync.SyncResult, error) {
	// Verify user is logged in
	if err := s.ensureUserLoggedIn(ctx); err != nil {
		return nil, err
	}

	s.logger.Info("Starting file sync from service layer")

	result, err := s.syncUseCase.SyncFiles(ctx)
	if err != nil {
		s.logger.Error("File sync failed", zap.Error(err))
		return nil, errors.NewAppError("file sync failed", err)
	}

	s.logger.Info("File sync completed successfully",
		zap.Int("files_processed", result.FilesProcessed),
		zap.Int("files_added", result.FilesAdded),
		zap.Int("files_updated", result.FilesUpdated),
		zap.Int("files_deleted", result.FilesDeleted),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}

func (s *syncService) FullSync(ctx context.Context) (*sync.SyncResult, error) {
	// Verify user is logged in
	if err := s.ensureUserLoggedIn(ctx); err != nil {
		return nil, err
	}

	s.logger.Info("Starting full sync from service layer")

	result, err := s.syncUseCase.FullSync(ctx)
	if err != nil {
		s.logger.Error("Full sync failed", zap.Error(err))
		return nil, errors.NewAppError("full sync failed", err)
	}

	s.logger.Info("Full sync completed successfully",
		zap.Int("collections_processed", result.CollectionsProcessed),
		zap.Int("collections_added", result.CollectionsAdded),
		zap.Int("collections_updated", result.CollectionsUpdated),
		zap.Int("collections_deleted", result.CollectionsDeleted),
		zap.Int("files_processed", result.FilesProcessed),
		zap.Int("files_added", result.FilesAdded),
		zap.Int("files_updated", result.FilesUpdated),
		zap.Int("files_deleted", result.FilesDeleted),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}

func (s *syncService) ResetSync(ctx context.Context) error {
	// Verify user is logged in
	if err := s.ensureUserLoggedIn(ctx); err != nil {
		return err
	}

	s.logger.Info("Resetting sync state from service layer")

	err := s.syncUseCase.ResetSync(ctx)
	if err != nil {
		s.logger.Error("Reset sync failed", zap.Error(err))
		return errors.NewAppError("reset sync failed", err)
	}

	s.logger.Info("Sync state reset successfully")
	return nil
}

func (s *syncService) ensureUserLoggedIn(ctx context.Context) error {
	userData, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("Failed to get authenticated user for sync", zap.Error(err))
		return errors.NewAppError("failed to get user data", err)
	}

	if userData == nil {
		s.logger.Error("User not authenticated for sync operation")
		return errors.NewAppError("user not authenticated; please login first", nil)
	}

	return nil
}
