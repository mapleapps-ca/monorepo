// cloud/backend/internal/maplefile/usecase/filemetadata/get_sync_data.go
package filemetadata

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetFileMetadataSyncDataUseCase interface {
	Execute(ctx context.Context, userID primitive.ObjectID, cursor *dom_file.FileSyncCursor, limit int64) (*dom_file.FileSyncResponse, error)
}

type getFileMetadataSyncDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewGetFileMetadataSyncDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) GetFileMetadataSyncDataUseCase {
	return &getFileMetadataSyncDataUseCaseImpl{config, logger, repo}
}

func (uc *getFileMetadataSyncDataUseCaseImpl) Execute(ctx context.Context, userID primitive.ObjectID, cursor *dom_file.FileSyncCursor, limit int64) (*dom_file.FileSyncResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.IsZero() {
		e["user_id"] = "User ID is required"
	}
	if cursor == nil {
		e["cursor"] = "Cursor is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get filtered collections",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get filtered collections from repository.
	//

	result, err := uc.repo.GetSyncData(ctx, userID, cursor, limit)
	if err != nil {
		uc.logger.Error("Failed to get filtered collections from repository",
			zap.Any("error", err),
			zap.Any("user_id", userID))
		return nil, err
	}

	return result, nil
}
