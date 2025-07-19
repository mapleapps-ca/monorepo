// monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/collection/count_collections.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// CountCollectionsResponse contains the collection counts for a user
type CountCollectionsResponse struct {
	OwnedCollections  int `json:"owned_collections"`
	SharedCollections int `json:"shared_collections"`
	TotalCollections  int `json:"total_collections"`
}

// CountFoldersResponse contains the folder counts for a user (folders only, not albums)
type CountFoldersResponse struct {
	OwnedFolders  int `json:"owned_folders"`
	SharedFolders int `json:"shared_folders"`
	TotalFolders  int `json:"total_folders"`
}

type CountUserCollectionsUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID) (*CountCollectionsResponse, error)
}

// NEW: Use case specifically for counting folders only
type CountUserFoldersUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID) (*CountFoldersResponse, error)
}

type countUserCollectionsUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

type countUserFoldersUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewCountUserCollectionsUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) CountUserCollectionsUseCase {
	logger = logger.Named("CountUserCollectionsUseCase")
	return &countUserCollectionsUseCaseImpl{config, logger, repo}
}

func NewCountUserFoldersUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) CountUserFoldersUseCase {
	logger = logger.Named("CountUserFoldersUseCase")
	return &countUserFoldersUseCaseImpl{config, logger, repo}
}

func (uc *countUserCollectionsUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID) (*CountCollectionsResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating count user collections",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Count collections.
	//

	ownedCollections, err := uc.repo.CountOwnedCollections(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to count owned collections",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	sharedCollections, err := uc.repo.CountSharedCollections(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to count shared collections",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	response := &CountCollectionsResponse{
		OwnedCollections:  ownedCollections,
		SharedCollections: sharedCollections,
		TotalCollections:  ownedCollections + sharedCollections,
	}

	uc.logger.Debug("Successfully counted user collections",
		zap.String("user_id", userID.String()),
		zap.Int("owned_collections", ownedCollections),
		zap.Int("shared_collections", sharedCollections),
		zap.Int("total_collections", response.TotalCollections))

	return response, nil
}

func (uc *countUserFoldersUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID) (*CountFoldersResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating count user folders",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: DEBUG - Check what's actually in the database
	//

	// ADD DEBUG LOGGING - Cast to concrete type to access debug method
	if debugRepo, ok := uc.repo.(interface {
		DebugCollectionRecords(context.Context, gocql.UUID) error
	}); ok {
		if debugErr := debugRepo.DebugCollectionRecords(ctx, userID); debugErr != nil {
			uc.logger.Warn("Failed to debug collection records", zap.Error(debugErr))
		}
	}

	//
	// STEP 3: Count folders with separate owned/shared counts AND total unique count
	//

	ownedFolders, err := uc.repo.CountOwnedFolders(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to count owned folders",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	sharedFolders, err := uc.repo.CountSharedFolders(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to count shared folders",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	// NEW: Get the deduplicated total count
	var totalUniqueFolders int
	if uniqueRepo, ok := uc.repo.(interface {
		CountTotalUniqueFolders(context.Context, gocql.UUID) (int, error)
	}); ok {
		totalUniqueFolders, err = uniqueRepo.CountTotalUniqueFolders(ctx, userID)
		if err != nil {
			uc.logger.Error("Failed to count unique total folders",
				zap.String("user_id", userID.String()),
				zap.Error(err))
			return nil, err
		}
	} else {
		// Fallback to simple addition if the method is not available
		uc.logger.Warn("CountTotalUniqueFolders method not available, using simple addition")
		totalUniqueFolders = ownedFolders + sharedFolders
	}

	response := &CountFoldersResponse{
		OwnedFolders:  ownedFolders,
		SharedFolders: sharedFolders,
		TotalFolders:  totalUniqueFolders, // Use deduplicated count
	}

	uc.logger.Info("Successfully counted user folders with deduplication",
		zap.String("user_id", userID.String()),
		zap.Int("owned_folders", ownedFolders),
		zap.Int("shared_folders", sharedFolders),
		zap.Int("total_unique_folders", totalUniqueFolders),
		zap.Int("would_be_simple_sum", ownedFolders+sharedFolders))

	return response, nil
}
