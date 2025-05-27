// cloud/backend/internal/maplefile/service/file/archive.go
package file

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type ArchiveFileRequestDTO struct {
	FileID primitive.ObjectID `json:"file_id"`
}

type ArchiveFileResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ArchiveFileService interface {
	Execute(ctx context.Context, req *ArchiveFileRequestDTO) (*ArchiveFileResponseDTO, error)
}

type archiveFileServiceImpl struct {
	config                *config.Configuration
	logger                *zap.Logger
	collectionRepo        dom_collection.CollectionRepository
	getMetadataUseCase    uc_filemetadata.GetFileMetadataUseCase
	updateMetadataUseCase uc_filemetadata.UpdateFileMetadataUseCase
}

func NewArchiveFileService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase,
	updateMetadataUseCase uc_filemetadata.UpdateFileMetadataUseCase,
) ArchiveFileService {
	return &archiveFileServiceImpl{
		config:                config,
		logger:                logger,
		collectionRepo:        collectionRepo,
		getMetadataUseCase:    getMetadataUseCase,
		updateMetadataUseCase: updateMetadataUseCase,
	}
}

func (svc *archiveFileServiceImpl) Execute(ctx context.Context, req *ArchiveFileRequestDTO) (*ArchiveFileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File ID is required")
	}

	if req.FileID.IsZero() {
		svc.logger.Warn("Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Get file metadata (including any state for archiving)
	//
	file, err := svc.getMetadataUseCase.ExecuteWithAnyState(req.FileID)
	if err != nil {
		svc.logger.Error("Failed to get file metadata",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	//
	// STEP 4: Check if user has write access to the file's collection
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(ctx, file.CollectionID, userID, dom_collection.CollectionPermissionReadWrite)
	if err != nil {
		svc.logger.Error("Failed to check collection access",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file archive attempt",
			zap.Any("user_id", userID),
			zap.Any("file_id", req.FileID),
			zap.Any("collection_id", file.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to archive this file")
	}

	//
	// STEP 5: Validate state transition
	//
	err = dom_file.IsValidStateTransition(file.State, dom_file.FileStateArchived)
	if err != nil {
		svc.logger.Warn("Invalid state transition for file archive",
			zap.Any("file_id", req.FileID),
			zap.String("current_state", file.State),
			zap.String("target_state", dom_file.FileStateArchived),
			zap.Error(err))
		return nil, httperror.NewForBadRequestWithSingleField("state", err.Error())
	}

	//
	// STEP 6: Archive the file
	//
	file.State = dom_file.FileStateArchived
	err = svc.updateMetadataUseCase.Execute(file)
	if err != nil {
		svc.logger.Error("Failed to archive file",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	svc.logger.Info("File archived successfully",
		zap.Any("file_id", req.FileID),
		zap.Any("collection_id", file.CollectionID),
		zap.Any("user_id", userID))

	return &ArchiveFileResponseDTO{
		Success: true,
		Message: "File archived successfully",
	}, nil
}
