// cloud/backend/internal/maplefile/service/file/list_by_owner_id.go
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

type ListFilesByOwnerIDRequestDTO struct {
	OwnerID gocql.UUID `json:"owner_id"`
}

type ListFilesByOwnerIDService interface {
	Execute(ctx context.Context, req *ListFilesByOwnerIDRequestDTO) (*FilesResponseDTO, error)
}

type listFilesByOwnerIDServiceImpl struct {
	config                   *config.Configuration
	logger                   *zap.Logger
	getFilesByOwnerIDUseCase uc_filemetadata.GetFileMetadataByOwnerIDUseCase
}

func NewListFilesByOwnerIDService(
	config *config.Configuration,
	logger *zap.Logger,
	getFilesByOwnerIDUseCase uc_filemetadata.GetFileMetadataByOwnerIDUseCase,
) ListFilesByOwnerIDService {
	logger = logger.Named("ListFilesByOwnerIDService")
	return &listFilesByOwnerIDServiceImpl{
		config:                   config,
		logger:                   logger,
		getFilesByOwnerIDUseCase: getFilesByOwnerIDUseCase,
	}
}

func (svc *listFilesByOwnerIDServiceImpl) Execute(ctx context.Context, req *ListFilesByOwnerIDRequestDTO) (*FilesResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Owner ID is required")
	}

	if req.OwnerID.IsZero() {
		svc.logger.Warn("Empty owner ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("owner_id", "Owner ID is required")
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
	if userID != req.OwnerID {
		svc.logger.Warn("Unauthorized attempt to list files created by another user",
			zap.Any("requesting_user_id", userID),
			zap.Any("owner_id", req.OwnerID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You can only view files you have created")
	}

	//
	// STEP 4: Get files by owner_id
	//
	files, err := svc.getFilesByOwnerIDUseCase.Execute(req.OwnerID)
	if err != nil {
		svc.logger.Error("Failed to get files by owner_id",
			zap.Any("error", err),
			zap.Any("owner_id", req.OwnerID))
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

	svc.logger.Debug("Found files by owner_id",
		zap.Int("count", len(files)),
		zap.Any("owner_id", req.OwnerID))

	return response, nil
}
