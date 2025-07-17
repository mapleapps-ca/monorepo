// monorepo/cloud/backend/internal/maplefile/service/file/restore.go
package file

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type RestoreFileRequestDTO struct {
	FileID gocql.UUID `json:"file_id"`
}

type RestoreFileResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RestoreFileService interface {
	Execute(ctx context.Context, req *RestoreFileRequestDTO) (*RestoreFileResponseDTO, error)
}

type restoreFileServiceImpl struct {
	config                *config.Configuration
	logger                *zap.Logger
	collectionRepo        dom_collection.CollectionRepository
	getMetadataUseCase    uc_filemetadata.GetFileMetadataUseCase
	updateMetadataUseCase uc_filemetadata.UpdateFileMetadataUseCase
}

func NewRestoreFileService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase,
	updateMetadataUseCase uc_filemetadata.UpdateFileMetadataUseCase,
) RestoreFileService {
	logger = logger.Named("RestoreFileService")
	return &restoreFileServiceImpl{
		config:                config,
		logger:                logger,
		collectionRepo:        collectionRepo,
		getMetadataUseCase:    getMetadataUseCase,
		updateMetadataUseCase: updateMetadataUseCase,
	}
}

func (svc *restoreFileServiceImpl) Execute(ctx context.Context, req *RestoreFileRequestDTO) (*RestoreFileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File ID is required")
	}

	if req.FileID.String() == "" {
		svc.logger.Warn("Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
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
	// STEP 3: Get file metadata (including any state for restoration)
	//
	file, err := svc.getMetadataUseCase.Execute(req.FileID)
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
		svc.logger.Warn("Unauthorized file restore attempt",
			zap.Any("user_id", userID),
			zap.Any("file_id", req.FileID),
			zap.Any("collection_id", file.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to restore this file")
	}

	//
	// STEP 5: Validate state transition
	//
	err = dom_file.IsValidStateTransition(file.State, dom_file.FileStateActive)
	if err != nil {
		svc.logger.Warn("Invalid state transition for file restore",
			zap.Any("file_id", req.FileID),
			zap.String("current_state", file.State),
			zap.String("target_state", dom_file.FileStateActive),
			zap.Error(err))
		return nil, httperror.NewForBadRequestWithSingleField("state", err.Error())
	}

	//
	// STEP 6: Restore the file
	//
	file.State = dom_file.FileStateActive
	file.Version++ // Mutation means we increment version.
	file.ModifiedAt = time.Now()
	file.ModifiedByUserID = userID
	err = svc.updateMetadataUseCase.Execute(ctx, file)
	if err != nil {
		svc.logger.Error("Failed to restore file",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	svc.logger.Info("File restored successfully",
		zap.Any("file_id", req.FileID),
		zap.Any("collection_id", file.CollectionID),
		zap.Any("user_id", userID))

	return &RestoreFileResponseDTO{
		Success: true,
		Message: "File restored successfully",
	}, nil
}
