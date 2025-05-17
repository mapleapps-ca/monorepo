// internal/service/remotefile/download.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotefile"
)

// DownloadOutput represents the result of downloading file data
type DownloadOutput struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	URL      string `json:"url,omitempty"`
	FileData []byte `json:"file_data,omitempty"`
}

// DownloadService defines the interface for downloading file data
type DownloadService interface {
	Download(ctx context.Context, id string) (*DownloadOutput, error)
	GetDownloadURL(ctx context.Context, id string) (*DownloadOutput, error)
}

// downloadService implements the DownloadService interface
type downloadService struct {
	logger              *zap.Logger
	downloadFileUseCase uc.DownloadRemoteFileUseCase
}

// NewDownloadService creates a new service for downloading file data
func NewDownloadService(
	logger *zap.Logger,
	downloadFileUseCase uc.DownloadRemoteFileUseCase,
) DownloadService {
	return &downloadService{
		logger:              logger,
		downloadFileUseCase: downloadFileUseCase,
	}
}

// Download downloads data for a remote file
func (s *downloadService) Download(ctx context.Context, id string) (*DownloadOutput, error) {
	// Validate inputs
	if id == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	// Download the file data
	fileData, err := s.downloadFileUseCase.Execute(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return &DownloadOutput{
		Success:  true,
		Message:  "File data downloaded successfully",
		FileData: fileData,
	}, nil
}

// GetDownloadURL gets a pre-signed URL for downloading a file
func (s *downloadService) GetDownloadURL(ctx context.Context, id string) (*DownloadOutput, error) {
	// Validate inputs
	if id == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	// Get the download URL
	url, err := s.downloadFileUseCase.GetDownloadURL(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return &DownloadOutput{
		Success: true,
		Message: "Download URL generated successfully",
		URL:     url,
	}, nil
}
