// native/desktop/maplefile-cli/internal/service/filedownload/download.go
package filedownload

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	uc_filedto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/filedto"
)

// DownloadService handles file download operations
type DownloadService interface {
	DownloadFile(ctx context.Context, fileID primitive.ObjectID, urlDuration time.Duration) (*uc_filedto.DownloadResponse, error)
}

type downloadService struct {
	logger                         *zap.Logger
	getPresignedDownloadURLUseCase uc_filedto.GetPresignedDownloadURLUseCase
	downloadFileUseCase            uc_filedto.DownloadFileUseCase
}

func NewDownloadService(
	logger *zap.Logger,
	getPresignedDownloadURLUseCase uc_filedto.GetPresignedDownloadURLUseCase,
	downloadFileUseCase uc_filedto.DownloadFileUseCase,
) DownloadService {
	return &downloadService{
		logger:                         logger,
		getPresignedDownloadURLUseCase: getPresignedDownloadURLUseCase,
		downloadFileUseCase:            downloadFileUseCase,
	}
}

func (s *downloadService) DownloadFile(ctx context.Context, fileID primitive.ObjectID, urlDuration time.Duration) (*uc_filedto.DownloadResponse, error) {
	s.logger.Info("Starting file download", zap.String("fileID", fileID.Hex()))

	//
	// Step 1: Get presigned download URLs
	//
	urlResponse, err := s.getPresignedDownloadURLUseCase.Execute(ctx, fileID, urlDuration)
	if err != nil {
		return nil, errors.NewAppError("failed to get presigned download URLs", err)
	}

	if !urlResponse.Success {
		return nil, errors.NewAppError("server failed to generate presigned URLs: "+urlResponse.Message, nil)
	}

	//
	// Step 2: Download file content using presigned URLs
	//
	downloadRequest := &uc_filedto.DownloadRequest{
		PresignedURL:          urlResponse.PresignedDownloadURL,
		PresignedThumbnailURL: urlResponse.PresignedThumbnailURL,
	}

	downloadResponse, err := s.downloadFileUseCase.Execute(ctx, downloadRequest)
	if err != nil {
		return nil, errors.NewAppError("failed to download file content", err)
	}

	s.logger.Info("Successfully completed file download",
		zap.String("fileID", fileID.Hex()),
		zap.Int64("fileSize", downloadResponse.FileSize))

	return downloadResponse, nil
}
