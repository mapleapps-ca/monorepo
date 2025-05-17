// internal/service/localfile/update.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// UpdateInput represents the input for updating a local file
type UpdateInput struct {
	ID                string  `json:"id"`
	EncryptedMetadata *string `json:"encrypted_metadata,omitempty"`
	DecryptedName     *string `json:"decrypted_name,omitempty"`
	DecryptedMimeType *string `json:"decrypted_mime_type,omitempty"`
	FileData          []byte  `json:"file_data,omitempty"`
	ThumbnailData     []byte  `json:"thumbnail_data,omitempty"`
}

// UpdateOutput represents the result of updating a local file
type UpdateOutput struct {
	File *localfile.LocalFile `json:"file"`
}

// UpdateService defines the interface for updating local files
type UpdateService interface {
	Update(ctx context.Context, input UpdateInput) (*UpdateOutput, error)
	UpdateSyncStatus(ctx context.Context, id string, remoteID string, status localfile.SyncStatus) (*UpdateOutput, error)
}

// updateService implements the UpdateService interface
type updateService struct {
	logger            *zap.Logger
	updateFileUseCase uc.UpdateLocalFileUseCase
}

// NewUpdateService creates a new service for updating local files
func NewUpdateService(
	logger *zap.Logger,
	updateFileUseCase uc.UpdateLocalFileUseCase,
) UpdateService {
	return &updateService{
		logger:            logger,
		updateFileUseCase: updateFileUseCase,
	}
}

// Update updates a local file
func (s *updateService) Update(ctx context.Context, input UpdateInput) (*UpdateOutput, error) {
	// Validate inputs
	if input.ID == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	// Prepare use case input
	useCaseInput := uc.UpdateLocalFileInput{
		ID:                    fileID,
		EncryptedMetadata:     input.EncryptedMetadata,
		DecryptedName:         input.DecryptedName,
		DecryptedMimeType:     input.DecryptedMimeType,
		FileData:              input.FileData,
		ThumbnailData:         input.ThumbnailData,
		MarkAsModifiedLocally: true,
	}

	// Call the use case
	file, err := s.updateFileUseCase.Execute(ctx, useCaseInput)
	if err != nil {
		return nil, err
	}

	return &UpdateOutput{
		File: file,
	}, nil
}

// UpdateSyncStatus updates the sync status of a local file
func (s *updateService) UpdateSyncStatus(
	ctx context.Context,
	id string,
	remoteID string,
	status localfile.SyncStatus,
) (*UpdateOutput, error) {
	// Validate inputs
	if id == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	// Convert remote ID to ObjectID if provided
	var remoteObjectID primitive.ObjectID
	if remoteID != "" {
		remoteObjectID, err = primitive.ObjectIDFromHex(remoteID)
		if err != nil {
			return nil, errors.NewAppError("invalid remote file ID format", err)
		}
	}

	// Call the use case
	file, err := s.updateFileUseCase.UpdateSyncStatus(ctx, fileID, remoteObjectID, status)
	if err != nil {
		return nil, err
	}

	return &UpdateOutput{
		File: file,
	}, nil
}
