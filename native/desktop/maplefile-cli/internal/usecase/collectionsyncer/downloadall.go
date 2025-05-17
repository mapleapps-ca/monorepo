// internal/usecase/collectionsyncer/downloadall.go
package collectionsyncer

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
	uc_remotecollection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotecollection"
)

// DownloadAllRemoteUseCase defines the interface for downloading all remote collections
type DownloadAllRemoteUseCase interface {
	Execute(ctx context.Context) (int, error)
}

// downloadAllRemoteUseCase implements the DownloadAllRemoteUseCase interface
type downloadAllRemoteUseCase struct {
	logger            *zap.Logger
	listRemoteUseCase uc_remotecollection.ListRemoteCollectionsUseCase
	downloadUseCase   DownloadToLocalUseCase
}

// NewDownloadAllRemoteUseCase creates a new use case for downloading all remote collections
func NewDownloadAllRemoteUseCase(
	logger *zap.Logger,
	listRemoteUseCase uc_remotecollection.ListRemoteCollectionsUseCase,
	downloadUseCase DownloadToLocalUseCase,
) DownloadAllRemoteUseCase {
	return &downloadAllRemoteUseCase{
		logger:            logger,
		listRemoteUseCase: listRemoteUseCase,
		downloadUseCase:   downloadUseCase,
	}
}

// Execute downloads all remote collections to local storage
func (uc *downloadAllRemoteUseCase) Execute(
	ctx context.Context,
) (int, error) {
	// Get all remote collections
	filter := remotecollection.CollectionFilter{}
	remoteCollections, err := uc.listRemoteUseCase.Execute(ctx, filter)
	if err != nil {
		return 0, errors.NewAppError("failed to list remote collections", err)
	}

	// Download each collection
	successCount := 0
	for _, collection := range remoteCollections {
		_, err := uc.downloadUseCase.Execute(ctx, collection.ID)
		if err != nil {
			// Log error but continue with other collections
			uc.logger.Error("Failed to download collection",
				zap.String("collectionID", collection.ID.Hex()),
				zap.Error(err))
			continue
		}
		successCount++
	}

	return successCount, nil
}
