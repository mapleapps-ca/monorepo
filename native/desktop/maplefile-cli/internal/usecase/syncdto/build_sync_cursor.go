// internal/usecase/syncdto/build_sync_cursor.go
package syncdto

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

// BuildSyncCursorInput represents the input for building a sync cursor
type BuildSyncCursorInput struct {
	LastModified time.Time
	LastID       gocql.UUID
}

// BuildSyncCursorUseCase defines the interface for building sync cursors
type BuildSyncCursorUseCase interface {
	Execute(ctx context.Context, input *BuildSyncCursorInput) (*syncdto.SyncCursorDTO, error)
	FromCollectionSyncItem(ctx context.Context, item *syncdto.CollectionSyncItem) (*syncdto.SyncCursorDTO, error)
	FromFileSyncItem(ctx context.Context, item *syncdto.FileSyncItem) (*syncdto.SyncCursorDTO, error)
}

// buildSyncCursorUseCase implements the BuildSyncCursorUseCase interface
type buildSyncCursorUseCase struct {
	logger *zap.Logger
}

// NewBuildSyncCursorUseCase creates a new use case for building sync cursors
func NewBuildSyncCursorUseCase(
	logger *zap.Logger,
) BuildSyncCursorUseCase {
	logger = logger.Named("BuildSyncCursorUseCase")
	return &buildSyncCursorUseCase{
		logger: logger,
	}
}

// Execute creates a sync cursor from the provided input
func (uc *buildSyncCursorUseCase) Execute(ctx context.Context, input *BuildSyncCursorInput) (*syncdto.SyncCursorDTO, error) {
	// Validate input
	if input == nil {
		return nil, errors.NewAppError("build sync cursor input is required", nil)
	}

	if input.LastModified.String() == "" {
		return nil, errors.NewAppError("last modified time is required", nil)
	}

	if input.LastID.String() == "" {
		return nil, errors.NewAppError("last ID is required", nil)
	}

	uc.logger.Debug("Building sync cursor",
		zap.Time("lastModified", input.LastModified),
		zap.String("lastID", input.LastID.String()))

	cursor := &syncdto.SyncCursorDTO{
		LastModified: input.LastModified,
		LastID:       input.LastID,
	}

	uc.logger.Debug("Successfully built sync cursor")
	return cursor, nil
}

// FromCollectionSyncItem creates a sync cursor from a collection sync item
func (uc *buildSyncCursorUseCase) FromCollectionSyncItem(ctx context.Context, item *syncdto.CollectionSyncItem) (*syncdto.SyncCursorDTO, error) {
	// Validate input
	if item == nil {
		return nil, errors.NewAppError("collection sync item is required", nil)
	}

	if item.ModifiedAt.String() == "" {
		return nil, errors.NewAppError("collection modified time is required", nil)
	}

	if item.ID.String() == "" {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	uc.logger.Debug("Building sync cursor from collection sync item",
		zap.String("collectionID", item.ID.String()),
		zap.Time("modifiedAt", item.ModifiedAt))

	input := &BuildSyncCursorInput{
		LastModified: item.ModifiedAt,
		LastID:       item.ID,
	}

	return uc.Execute(ctx, input)
}

// FromFileSyncItem creates a sync cursor from a file sync item
func (uc *buildSyncCursorUseCase) FromFileSyncItem(ctx context.Context, item *syncdto.FileSyncItem) (*syncdto.SyncCursorDTO, error) {
	// Validate input
	if item == nil {
		return nil, errors.NewAppError("file sync item is required", nil)
	}

	if item.ModifiedAt.String() == "" {
		return nil, errors.NewAppError("file modified time is required", nil)
	}

	if item.ID.String() == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	uc.logger.Debug("Building sync cursor from file sync item",
		zap.String("fileID", item.ID.String()),
		zap.Time("modifiedAt", item.ModifiedAt))

	input := &BuildSyncCursorInput{
		LastModified: item.ModifiedAt,
		LastID:       item.ID,
	}

	return uc.Execute(ctx, input)
}
