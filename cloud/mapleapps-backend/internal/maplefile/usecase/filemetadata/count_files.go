// cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata/count_files.go
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

// CountFilesResponse contains the file count for a user
type CountFilesResponse struct {
	TotalFiles int `json:"total_files"`
}

type CountUserFilesUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID) (*CountFilesResponse, error)
}

type countUserFilesUseCaseImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileMetadataRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewCountUserFilesUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileMetadataRepository,
	collectionRepo dom_collection.CollectionRepository,
) CountUserFilesUseCase {
	logger = logger.Named("CountUserFilesUseCase")
	return &countUserFilesUseCaseImpl{config, logger, fileRepo, collectionRepo}
}

func (uc *countUserFilesUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID) (*CountFilesResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating count user files",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get accessible collections for the user.
	//

	// Get collections using the efficient filtered query
	filterOptions := dom_collection.CollectionFilterOptions{
		UserID:        userID,
		IncludeOwned:  true,
		IncludeShared: true,
	}

	collectionResult, err := uc.collectionRepo.GetCollectionsWithFilter(ctx, filterOptions)
	if err != nil {
		uc.logger.Error("Failed to get accessible collections for file count",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	// Extract collection IDs
	allCollections := collectionResult.GetAllCollections()
	accessibleCollectionIDs := make([]gocql.UUID, 0, len(allCollections))
	for _, collection := range allCollections {
		accessibleCollectionIDs = append(accessibleCollectionIDs, collection.ID)
	}

	uc.logger.Debug("Found accessible collections for file counting",
		zap.String("user_id", userID.String()),
		zap.Int("owned_collections", len(collectionResult.OwnedCollections)),
		zap.Int("shared_collections", len(collectionResult.SharedCollections)),
		zap.Int("total_accessible", len(accessibleCollectionIDs)))

	//
	// STEP 3: Count files in accessible collections.
	//

	fileCount, err := uc.fileRepo.CountFilesByUser(ctx, userID, accessibleCollectionIDs)
	if err != nil {
		uc.logger.Error("Failed to count files for user",
			zap.String("user_id", userID.String()),
			zap.Int("accessible_collections", len(accessibleCollectionIDs)),
			zap.Error(err))
		return nil, err
	}

	response := &CountFilesResponse{
		TotalFiles: fileCount,
	}

	uc.logger.Debug("Successfully counted user files",
		zap.String("user_id", userID.String()),
		zap.Int("accessible_collections", len(accessibleCollectionIDs)),
		zap.Int("total_files", fileCount))

	return response, nil
}
