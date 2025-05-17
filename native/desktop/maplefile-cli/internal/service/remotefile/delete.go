// internal/service/remotefile/delete.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotefile"
)

// DeleteOutput represents the result of deleting remote files
type DeleteOutput struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	DeletedCount int    `json:"deleted_count,omitempty"`
}

// DeleteService defines the interface for deleting remote files
type DeleteService interface {
	Delete(ctx context.Context, id string) (*DeleteOutput, error)
	DeleteMultiple(ctx context.Context, ids []string) (*DeleteOutput, error)
}

// deleteService implements the DeleteService interface
type deleteService struct {
	logger            *zap.Logger
	deleteFileUseCase uc.DeleteRemoteFileUseCase
}

// NewDeleteService creates a new service for deleting remote files
func NewDeleteService(
	logger *zap.Logger,
	deleteFileUseCase uc.DeleteRemoteFileUseCase,
) DeleteService {
	return &deleteService{
		logger:            logger,
		deleteFileUseCase: deleteFileUseCase,
	}
}

// Delete deletes a remote file
func (s *deleteService) Delete(ctx context.Context, id string) (*DeleteOutput, error) {
	// Validate inputs
	if id == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	// Delete the file
	if err := s.deleteFileUseCase.Execute(ctx, fileID); err != nil {
		return nil, err
	}

	return &DeleteOutput{
		Success: true,
		Message: "Remote file deleted successfully",
	}, nil
}

// DeleteMultiple deletes multiple remote files
func (s *deleteService) DeleteMultiple(ctx context.Context, ids []string) (*DeleteOutput, error) {
	// Validate inputs
	if len(ids) == 0 {
		return nil, errors.NewAppError("at least one file ID is required", nil)
	}

	// Convert IDs to ObjectIDs
	fileIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		fileID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, errors.NewAppError("invalid file ID format: "+id, err)
		}
		fileIDs = append(fileIDs, fileID)
	}

	// Delete the files
	count, err := s.deleteFileUseCase.DeleteMultiple(ctx, fileIDs)
	if err != nil {
		return nil, err
	}

	message := "Remote files deleted successfully"
	if count < len(ids) {
		message = "Some files could not be deleted"
	}

	return &DeleteOutput{
		Success:      true,
		Message:      message,
		DeletedCount: count,
	}, nil
}
