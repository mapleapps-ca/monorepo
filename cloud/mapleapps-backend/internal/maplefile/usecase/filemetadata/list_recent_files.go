// cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata/list_recent_files.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ListRecentFilesUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID, cursor *dom_file.RecentFilesCursor, limit int64) (*dom_file.RecentFilesResponse, error)
}

type listRecentFilesUseCaseImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileMetadataRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewListRecentFilesUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileMetadataRepository,
	collectionRepo dom_collection.CollectionRepository,
) ListRecentFilesUseCase {
	logger = logger.Named("ListRecentFilesUseCase")
	return &listRecentFilesUseCaseImpl{config, logger, fileRepo, collectionRepo}
}

func (uc *listRecentFilesUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID, cursor *dom_file.RecentFilesCursor, limit int64) (*dom_file.RecentFilesResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if limit <= 0 {
		e["limit"] = "Limit must be greater than 0"
	}
	if limit > 100 {
		e["limit"] = "Limit cannot exceed 100"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating list recent files",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get accessible collections for the user.
	//

	uc.logger.Debug("Getting accessible collections for recent files",
		zap.String("user_id", userID.String()))

	// Get collections using the efficient filtered query
	filterOptions := dom_collection.CollectionFilterOptions{
		UserID:        userID,
		IncludeOwned:  true,
		IncludeShared: true,
	}

	collectionResult, err := uc.collectionRepo.GetCollectionsWithFilter(ctx, filterOptions)
	if err != nil {
		uc.logger.Error("Failed to get accessible collections for recent files",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	// Extract collection IDs
	allCollections := collectionResult.GetAllCollections()
	accessibleCollectionIDs := make([]gocql.UUID, 0, len(allCollections))
	for _, collection := range allCollections {
		// Only include active collections
		if collection.State == "active" {
			accessibleCollectionIDs = append(accessibleCollectionIDs, collection.ID)
		}
	}

	uc.logger.Debug("Found accessible collections for recent files",
		zap.String("user_id", userID.String()),
		zap.Int("owned_collections", len(collectionResult.OwnedCollections)),
		zap.Int("shared_collections", len(collectionResult.SharedCollections)),
		zap.Int("total_accessible", len(accessibleCollectionIDs)))

	// If no accessible collections, return empty response
	if len(accessibleCollectionIDs) == 0 {
		uc.logger.Info("User has no accessible collections for recent files",
			zap.String("user_id", userID.String()))
		return &dom_file.RecentFilesResponse{
			Files:      []dom_file.RecentFilesItem{},
			NextCursor: nil,
			HasMore:    false,
		}, nil
	}

	//
	// STEP 3: List recent files for accessible collections.
	//

	recentFiles, err := uc.fileRepo.ListRecentFiles(ctx, userID, cursor, limit, accessibleCollectionIDs)
	if err != nil {
		uc.logger.Error("Failed to list recent files",
			zap.Any("error", err),
			zap.String("user_id", userID.String()))
		return nil, err
	}

	if recentFiles == nil {
		uc.logger.Debug("Recent files not found",
			zap.String("user_id", userID.String()))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Recent files not found")
	}

	uc.logger.Debug("Recent files successfully retrieved",
		zap.String("user_id", userID.String()),
		zap.Any("next_cursor", recentFiles.NextCursor),
		zap.Int("files_count", len(recentFiles.Files)))

	return recentFiles, nil
}
