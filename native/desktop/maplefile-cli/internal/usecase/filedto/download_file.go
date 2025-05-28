// native/desktop/maplefile-cli/internal/usecase/filedto/download_file.go
package filedto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
)

// DownloadRequest represents a file download request
type DownloadRequest struct {
	PresignedURL          string `json:"presigned_url"`
	PresignedThumbnailURL string `json:"presigned_thumbnail_url,omitempty"`
}

// DownloadResponse represents a file download response
type DownloadResponse struct {
	FileData      []byte `json:"file_data"`
	ThumbnailData []byte `json:"thumbnail_data,omitempty"`
	FileSize      int64  `json:"file_size"`
	ThumbnailSize int64  `json:"thumbnail_size"`
}

// DownloadFileUseCase defines the interface for downloading files
type DownloadFileUseCase interface {
	Execute(ctx context.Context, request *DownloadRequest) (*DownloadResponse, error)
}

// downloadFileUseCase implements the DownloadFileUseCase interface
type downloadFileUseCase struct {
	logger      *zap.Logger
	fileDTORepo filedto.FileDTORepository
}

// NewDownloadFileUseCase creates a new use case for downloading files
func NewDownloadFileUseCase(
	logger *zap.Logger,
	fileDTORepo filedto.FileDTORepository,
) DownloadFileUseCase {
	logger = logger.Named("DownloadFileUseCase")
	return &downloadFileUseCase{
		logger:      logger,
		fileDTORepo: fileDTORepo,
	}
}

// Execute downloads file content using presigned URLs
func (uc *downloadFileUseCase) Execute(
	ctx context.Context,
	request *DownloadRequest,
) (*DownloadResponse, error) {
	// Validate inputs
	if request == nil {
		return nil, errors.NewAppError("download request is required", nil)
	}
	if request.PresignedURL == "" {
		return nil, errors.NewAppError("presigned URL is required", nil)
	}

	// Download main file content
	fileData, err := uc.fileDTORepo.DownloadFileViaPresignedURLFromCloud(ctx, request.PresignedURL)
	if err != nil {
		return nil, errors.NewAppError("failed to download file content", err)
	}

	response := &DownloadResponse{
		FileData: fileData,
		FileSize: int64(len(fileData)),
	}

	// Download thumbnail if URL provided
	if request.PresignedThumbnailURL != "" {
		thumbnailData, err := uc.fileDTORepo.DownloadThumbnailViaPresignedURLFromCloud(ctx, request.PresignedThumbnailURL)
		if err != nil {
			uc.logger.Warn("Failed to download thumbnail, continuing without it",
				zap.Error(err))
		} else if thumbnailData != nil {
			response.ThumbnailData = thumbnailData
			response.ThumbnailSize = int64(len(thumbnailData))
		}
	}

	uc.logger.Info("Successfully downloaded file content",
		zap.Int64("fileSize", response.FileSize),
		zap.Int64("thumbnailSize", response.ThumbnailSize))

	return response, nil
}
