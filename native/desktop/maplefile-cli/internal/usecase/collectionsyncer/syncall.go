// internal/usecase/collectionsyncer/syncall.go
package collectionsyncer

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
	uc_localcollection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localcollection"
)

// SyncAllLocalToRemoteUseCase defines the interface for syncing all local collections
type SyncAllLocalToRemoteUseCase interface {
	Execute(ctx context.Context) (int, error)
}

// syncAllLocalToRemoteUseCase implements the SyncAllLocalToRemoteUseCase interface
type syncAllLocalToRemoteUseCase struct {
	logger           *zap.Logger
	listLocalUseCase uc_localcollection.ListLocalCollectionsUseCase
	uploadUseCase    UploadToRemoteUseCase
}

// NewSyncAllLocalToRemoteUseCase creates a new use case for syncing all local collections
func NewSyncAllLocalToRemoteUseCase(
	logger *zap.Logger,
	listLocalUseCase uc_localcollection.ListLocalCollectionsUseCase,
	uploadUseCase UploadToRemoteUseCase,
) SyncAllLocalToRemoteUseCase {
	return &syncAllLocalToRemoteUseCase{
		logger:           logger,
		listLocalUseCase: listLocalUseCase,
		uploadUseCase:    uploadUseCase,
	}
}

// Execute syncs all local collections to the remote server
func (uc *syncAllLocalToRemoteUseCase) Execute(
	ctx context.Context,
) (int, error) {
	// Get all modified local collections
	modifiedCollections, err := uc.listLocalUseCase.ListModifiedLocally(ctx)
	if err != nil {
		return 0, errors.NewAppError("failed to list modified local collections", err)
	}

	// Also get local-only collections
	status := localcollection.SyncStatusLocalOnly
	filter := localcollection.LocalCollectionFilter{
		SyncStatus: &status,
	}
	localOnlyCollections, err := uc.listLocalUseCase.Execute(ctx, filter)
	if err != nil {
		return 0, errors.NewAppError("failed to list local-only collections", err)
	}

	// Combine both lists
	allCollectionsToSync := append(modifiedCollections, localOnlyCollections...)

	// Sync each collection
	successCount := 0
	for _, collection := range allCollectionsToSync {
		_, err := uc.uploadUseCase.Execute(ctx, collection.ID)
		if err != nil {
			// Log error but continue with other collections
			uc.logger.Error("Failed to sync collection",
				zap.String("collectionID", collection.ID.Hex()),
				zap.Error(err))
			continue
		}
		successCount++
	}

	return successCount, nil
}
