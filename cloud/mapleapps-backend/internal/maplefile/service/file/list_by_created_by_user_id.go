// monorepo/cloud/backend/internal/maplefile/service/file/list_by_created_by_user_id.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ListFilesByCreatedByUserIDRequestDTO struct {
	CreatedByUserID gocql.UUID `json:"created_by_user_id"`
}

type ListFilesByCreatedByUserIDService interface {
	Execute(ctx context.Context, req *ListFilesByCreatedByUserIDRequestDTO) (*FilesResponseDTO, error)
}

type listFilesByCreatedByUserIDServiceImpl struct {
	config                           *config.Configuration
	logger                           *zap.Logger
	getFilesByCreatedByUserIDUseCase uc_filemetadata.GetFileMetadataByCreatedByUserIDUseCase
}

func NewListFilesByCreatedByUserIDService(
	config *config.Configuration,
	logger *zap.Logger,
	getFilesByCreatedByUserIDUseCase uc_filemetadata.GetFileMetadataByCreatedByUserIDUseCase,
) ListFilesByCreatedByUserIDService {
	logger = logger.Named("ListFilesByCreatedByUserIDService")
	return &listFilesByCreatedByUserIDServiceImpl{
		config:                           config,
		logger:                           logger,
		getFilesByCreatedByUserIDUseCase: getFilesByCreatedByUserIDUseCase,
	}
}

func (svc *listFilesByCreatedByUserIDServiceImpl) Execute(ctx context.Context, req *ListFilesByCreatedByUserIDRequestDTO) (*FilesResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Created by user ID is required")
	}

	if req.CreatedByUserID.String() == "" {
		svc.logger.Warn("Empty created by user ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("created_by_user_id", "Created by user ID is required")
	}

	//
	// STEP 2: Get user ID from context (for authorization)
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Check if the requesting user can access files created by the specified user
	// Only allow users to see their own created files for privacy
	//
	if userID != req.CreatedByUserID {
		svc.logger.Warn("Unauthorized attempt to list files created by another user",
			zap.Any("requesting_user_id", userID),
			zap.Any("created_by_user_id", req.CreatedByUserID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You can only view files you have created")
	}

	//
	// STEP 4: Get files by created_by_user_id
	//
	files, err := svc.getFilesByCreatedByUserIDUseCase.Execute(req.CreatedByUserID)
	if err != nil {
		svc.logger.Error("Failed to get files by created_by_user_id",
			zap.Any("error", err),
			zap.Any("created_by_user_id", req.CreatedByUserID))
		return nil, err
	}

	//
	// STEP 5: Map domain models to response DTOs
	//
	response := &FilesResponseDTO{
		Files: make([]*FileResponseDTO, len(files)),
	}

	for i, file := range files {
		response.Files[i] = mapFileToDTO(file)
	}

	svc.logger.Debug("Found files by created_by_user_id",
		zap.Int("count", len(files)),
		zap.Any("created_by_user_id", req.CreatedByUserID))

	return response, nil
}
