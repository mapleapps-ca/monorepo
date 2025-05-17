// internal/service/localfile/delete.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// DeleteOutput represents the result of deleting files
type DeleteOutput struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	DeletedCount int    `json:"deleted_count,omitempty"`
}

// DeleteService defines the interface for deleting local files
type DeleteService interface {
	Delete(ctx context.Context, id string) (*DeleteOutput, error)
	DeleteMultiple(ctx context.Context, ids []string) (*DeleteOutput, error)
	DeleteByCollection(ctx context.Context, collectionID string) (*DeleteOutput, error)
}

// deleteService implements the DeleteService interface
type deleteService struct {
	logger            *zap.Logger
	deleteFileUseCase localfile.DeleteLocalFileUseCase
}

// NewDeleteService creates a new service for deleting local files
func NewDeleteService(
	logger *zap.Logger,
	deleteFileUseCase localfile.DeleteLocalFileUseCase,
) DeleteService {
	return &deleteService{
		logger:            logger,
		deleteFileUseCase: deleteFileUseCase,
	}
}

// Delete deletes a local file
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
		Message: "File deleted successfully",
	}, nil
}

// DeleteMultiple deletes multiple local files
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

	message := "Files deleted successfully"
	if count < len(ids) {
		message = "Some files could not be deleted"
	}

	return &DeleteOutput{
		Success:      true,
		Message:      message,
		DeletedCount: count,
	}, nil
}

// DeleteByCollection deletes all files in a collection
func (s *deleteService) DeleteByCollection(ctx context.Context, collectionID string) (*DeleteOutput, error) {
	// Validate inputs
	if collectionID == "" {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Convert collection ID to ObjectID
	colID, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Delete the files
	count, err := s.deleteFileUseCase.DeleteByCollection(ctx, colID)
	if err != nil {
		return nil, err
	}

	return &DeleteOutput{
		Success:      true,
		Message:      "All files in collection deleted successfully",
		DeletedCount: count,
	}, nil
}
