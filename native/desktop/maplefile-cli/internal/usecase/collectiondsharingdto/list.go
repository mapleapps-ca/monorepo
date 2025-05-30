// native/desktop/maplefile-cli/internal/usecase/collectiondsharingdto/list.go
package collectiondsharingdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
)

// ListSharedCollectionsUseCase defines the interface for listing shared collections
type ListSharedCollectionsUseCase interface {
	Execute(ctx context.Context) ([]*collectiondto.CollectionDTO, error)
}

// listSharedCollectionsUseCase implements the ListSharedCollectionsUseCase interface
type listSharedCollectionsUseCase struct {
	logger      *zap.Logger
	sharingRepo collectionsharingdto.CollectionSharingDTORepository
}

// NewListSharedCollectionsUseCase creates a new use case for listing shared collections
func NewListSharedCollectionsUseCase(
	logger *zap.Logger,
	sharingRepo collectionsharingdto.CollectionSharingDTORepository,
) ListSharedCollectionsUseCase {
	logger = logger.Named("ListSharedCollectionsUseCase")
	return &listSharedCollectionsUseCase{
		logger:      logger,
		sharingRepo: sharingRepo,
	}
}

// Execute lists collections shared with the current user
func (uc *listSharedCollectionsUseCase) Execute(ctx context.Context) ([]*collectiondto.CollectionDTO, error) {
	// Get shared collections from repository
	collections, err := uc.sharingRepo.ListSharedCollectionsFromCloud(ctx)
	if err != nil {
		uc.logger.Error("❌ Failed to list shared collections", zap.Error(err))
		return nil, err
	}

	uc.logger.Info("✅ Successfully listed shared collections",
		zap.Int("count", len(collections)))

	return collections, nil
}
