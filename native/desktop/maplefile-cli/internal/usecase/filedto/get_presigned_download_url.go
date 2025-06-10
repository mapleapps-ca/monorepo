// native/desktop/maplefile-cli/internal/usecase/filedto/get_presigned_download_url.go
package filedto

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
)

// GetPresignedDownloadURLUseCase defines the interface for getting presigned download URLs
type GetPresignedDownloadURLUseCase interface {
	Execute(ctx context.Context, fileID gocql.UUID, urlDuration time.Duration) (*filedto.GetPresignedDownloadURLResponse, error)
}

// getPresignedDownloadURLUseCase implements the GetPresignedDownloadURLUseCase interface
type getPresignedDownloadURLUseCase struct {
	logger      *zap.Logger
	fileDTORepo filedto.FileDTORepository
}

// NewGetPresignedDownloadURLUseCase creates a new use case for getting presigned download URLs
func NewGetPresignedDownloadURLUseCase(
	logger *zap.Logger,
	fileDTORepo filedto.FileDTORepository,
) GetPresignedDownloadURLUseCase {
	logger = logger.Named("GetPresignedDownloadURLUseCase")
	return &getPresignedDownloadURLUseCase{
		logger:      logger,
		fileDTORepo: fileDTORepo,
	}
}

// Execute gets presigned download URLs for a file
func (uc *getPresignedDownloadURLUseCase) Execute(
	ctx context.Context,
	fileID gocql.UUID,
	urlDuration time.Duration,
) (*filedto.GetPresignedDownloadURLResponse, error) {
	// Validate inputs
	if fileID.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Set default duration if not provided
	if urlDuration == 0 {
		urlDuration = 1 * time.Hour
	}

	// Validate duration is reasonable
	maxDuration := 24 * time.Hour
	if urlDuration > maxDuration {
		return nil, errors.NewAppError("URL duration cannot exceed 24 hours", nil)
	}

	// Create request
	request := &filedto.GetPresignedDownloadURLRequest{
		URLDuration: urlDuration,
	}

	// Get presigned URLs from cloud
	response, err := uc.fileDTORepo.GetPresignedDownloadURLFromCloud(ctx, fileID, request)
	if err != nil {
		return nil, errors.NewAppError("failed to get presigned download URLs", err)
	}

	return response, nil
}
