// cloud/backend/internal/maplefile/service/file/delete_multiple.go
package file

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	uc_fileobjectstorage "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/fileobjectstorage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type DeleteMultipleFilesRequestDTO struct {
	FileIDs []gocql.UUID `json:"file_ids"`
}

type DeleteMultipleFilesResponseDTO struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	DeletedCount   int    `json:"deleted_count"`
	SkippedCount   int    `json:"skipped_count"`
	TotalRequested int    `json:"total_requested"`
}

type DeleteMultipleFilesService interface {
	Execute(ctx context.Context, req *DeleteMultipleFilesRequestDTO) (*DeleteMultipleFilesResponseDTO, error)
}

type deleteMultipleFilesServiceImpl struct {
	config                    *config.Configuration
	logger                    *zap.Logger
	collectionRepo            dom_collection.CollectionRepository
	getMetadataByIDsUseCase   uc_filemetadata.GetFileMetadataByIDsUseCase
	deleteMetadataManyUseCase uc_filemetadata.DeleteManyFileMetadataUseCase
	deleteMultipleDataUseCase uc_fileobjectstorage.DeleteMultipleEncryptedDataUseCase
}

func NewDeleteMultipleFilesService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataByIDsUseCase uc_filemetadata.GetFileMetadataByIDsUseCase,
	deleteMetadataManyUseCase uc_filemetadata.DeleteManyFileMetadataUseCase,
	deleteMultipleDataUseCase uc_fileobjectstorage.DeleteMultipleEncryptedDataUseCase,
) DeleteMultipleFilesService {
	logger = logger.Named("DeleteMultipleFilesService")
	return &deleteMultipleFilesServiceImpl{
		config:                    config,
		logger:                    logger,
		collectionRepo:            collectionRepo,
		getMetadataByIDsUseCase:   getMetadataByIDsUseCase,
		deleteMetadataManyUseCase: deleteMetadataManyUseCase,
		deleteMultipleDataUseCase: deleteMultipleDataUseCase,
	}
}

func (svc *deleteMultipleFilesServiceImpl) Execute(ctx context.Context, req *DeleteMultipleFilesRequestDTO) (*DeleteMultipleFilesResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File IDs are required")
	}

	if req.FileIDs == nil || len(req.FileIDs) == 0 {
		svc.logger.Warn("Empty file IDs provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_ids", "File IDs are required")
	}

	// Validate individual file IDs
	e := make(map[string]string)
	for i, fileID := range req.FileIDs {
		if fileID.IsZero() {
			e[fmt.Sprintf("file_ids[%d]", i)] = "File ID is required"
		}
	}
	if len(e) != 0 {
		svc.logger.Warn("Failed validation",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Get file metadata for all files
	//
	files, err := svc.getMetadataByIDsUseCase.Execute(req.FileIDs)
	if err != nil {
		svc.logger.Error("Failed to get file metadata",
			zap.Any("error", err),
			zap.Any("file_ids", req.FileIDs))
		return nil, err
	}

	//
	// STEP 4: Filter files that the user has permission to delete
	//
	var deletableFiles []*dom_file.File
	var storagePaths []string
	skippedCount := 0

	for _, file := range files {
		hasAccess, err := svc.collectionRepo.CheckAccess(ctx, file.CollectionID, userID, dom_collection.CollectionPermissionReadWrite)
		if err != nil {
			svc.logger.Warn("Failed to check access for file, skipping",
				zap.Any("error", err),
				zap.Any("file_id", file.ID),
				zap.Any("collection_id", file.CollectionID))
			skippedCount++
			continue
		}

		if !hasAccess {
			svc.logger.Warn("User doesn't have permission to delete file, skipping",
				zap.Any("user_id", userID),
				zap.Any("file_id", file.ID),
				zap.Any("collection_id", file.CollectionID))
			skippedCount++
			continue
		}

		// Check valid transitions.
		if err := dom_collection.IsValidStateTransition(file.State, dom_file.FileStateDeleted); err != nil {
			svc.logger.Warn("Invalid file state transition",
				zap.Any("user_id", userID),
				zap.Error(err))
			skippedCount++
			continue
		}

		deletableFiles = append(deletableFiles, file)
		storagePaths = append(storagePaths, file.EncryptedFileObjectKey)

		// Add thumbnail paths if they exist
		if file.EncryptedThumbnailObjectKey != "" {
			storagePaths = append(storagePaths, file.EncryptedThumbnailObjectKey)
		}
	}

	if len(deletableFiles) == 0 {
		return &DeleteMultipleFilesResponseDTO{
			Success:        true,
			Message:        "No files could be deleted due to permission restrictions",
			DeletedCount:   0,
			SkippedCount:   len(req.FileIDs),
			TotalRequested: len(req.FileIDs),
		}, nil
	}

	//
	// STEP 5: Delete encrypted data from object storage
	//
	if len(storagePaths) > 0 {
		err = svc.deleteMultipleDataUseCase.Execute(storagePaths)
		if err != nil {
			svc.logger.Error("Failed to delete some encrypted data, continuing with metadata deletion",
				zap.Any("error", err),
				zap.Int("storage_paths_count", len(storagePaths)))
		}
	}

	//TODO: FIX keeping track of version + modified at
	// file.Version++ // Mutation means we increment version.
	// file.ModifiedAt = time.Now()
	// file.ModifiedByUserID = userID

	//
	// STEP 6: Delete file metadata
	//
	deletableFileIDs := make([]gocql.UUID, len(deletableFiles))
	for i, file := range deletableFiles {
		deletableFileIDs[i] = file.ID
	}

	err = svc.deleteMetadataManyUseCase.Execute(deletableFileIDs)
	if err != nil {
		svc.logger.Error("Failed to delete file metadata",
			zap.Any("error", err),
			zap.Any("file_ids", deletableFileIDs))
		return nil, err
	}

	svc.logger.Info("Multiple files deleted successfully",
		zap.Int("deleted_count", len(deletableFiles)),
		zap.Int("skipped_count", skippedCount),
		zap.Int("total_requested", len(req.FileIDs)),
		zap.Any("user_id", userID))

	return &DeleteMultipleFilesResponseDTO{
		Success:        true,
		Message:        fmt.Sprintf("Successfully deleted %d files", len(deletableFiles)),
		DeletedCount:   len(deletableFiles),
		SkippedCount:   skippedCount,
		TotalRequested: len(req.FileIDs),
	}, nil
}
