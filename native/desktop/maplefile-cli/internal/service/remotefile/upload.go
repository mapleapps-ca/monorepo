// internal/service/remotefile/upload.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotefile"
)

// UploadOutput represents the result of uploading file data
type UploadOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
}

// UploadService defines the interface for uploading file data
type UploadService interface {
	Upload(ctx context.Context, id string, data []byte) (*UploadOutput, error)
}

// uploadService implements the UploadService interface
type uploadService struct {
	logger            *zap.Logger
	uploadFileUseCase uc.UploadRemoteFileUseCase
}

// NewUploadService creates a new service for uploading file data
func NewUploadService(
	logger *zap.Logger,
	uploadFileUseCase uc.UploadRemoteFileUseCase,
) UploadService {
	return &uploadService{
		logger:            logger,
		uploadFileUseCase: uploadFileUseCase,
	}
}

// Upload uploads data for a remote file
func (s *uploadService) Upload(ctx context.Context, id string, data []byte) (*UploadOutput, error) {
	// Validate inputs
	if id == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	if data == nil || len(data) == 0 {
		return nil, errors.NewAppError("file data is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	// Upload the file data
	if err := s.uploadFileUseCase.Execute(ctx, fileID, data); err != nil {
		return nil, err
	}

	return &UploadOutput{
		Success: true,
		Message: "File data uploaded successfully",
	}, nil
}
