// internal/usecase/localfile/get.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
)

// GetLocalFileUseCase defines the interface for getting a local file
type GetLocalFileUseCase interface {
	ByID(ctx context.Context, id primitive.ObjectID) (*localfile.LocalFile, error)
	ByRemoteID(ctx context.Context, remoteID primitive.ObjectID) (*localfile.LocalFile, error)
	GetEncryptedFileData(ctx context.Context, file *localfile.LocalFile) ([]byte, error)
	GetDecryptedFileData(ctx context.Context, file *localfile.LocalFile) ([]byte, error)
	GetThumbnail(ctx context.Context, file *localfile.LocalFile) ([]byte, error)
}

// getLocalFileUseCase implements the GetLocalFileUseCase interface
type getLocalFileUseCase struct {
	logger     *zap.Logger
	repository localfile.LocalFileRepository
}

// NewGetLocalFileUseCase creates a new use case for getting local files
func NewGetLocalFileUseCase(
	logger *zap.Logger,
	repository localfile.LocalFileRepository,
) GetLocalFileUseCase {
	return &getLocalFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// ByID retrieves a local file by ID
func (uc *getLocalFileUseCase) ByID(
	ctx context.Context,
	id primitive.ObjectID,
) (*localfile.LocalFile, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Get the file from the repository
	file, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewAppError("failed to get local file", err)
	}

	if file == nil {
		return nil, errors.NewAppError("local file not found", nil)
	}

	return file, nil
}

// ByRemoteID retrieves a local file by its remote ID
func (uc *getLocalFileUseCase) ByRemoteID(
	ctx context.Context,
	remoteID primitive.ObjectID,
) (*localfile.LocalFile, error) {
	// Validate inputs
	if remoteID.IsZero() {
		return nil, errors.NewAppError("remote file ID is required", nil)
	}

	// Get the file from the repository
	file, err := uc.repository.GetByRemoteID(ctx, remoteID)
	if err != nil {
		return nil, errors.NewAppError("failed to get local file by remote ID", err)
	}

	if file == nil {
		return nil, errors.NewAppError("local file with the specified remote ID not found", nil)
	}

	return file, nil
}

// GetDecryptedFileData retrieves the file data for a local file
func (uc *getLocalFileUseCase) GetDecryptedFileData(
	ctx context.Context,
	file *localfile.LocalFile,
) ([]byte, error) {
	// Validate inputs
	if file == nil {
		return nil, errors.NewAppError("file is required", nil)
	}

	// Load the file data
	data, err := uc.repository.LoadDecryptedFileDataAtFilePath(ctx, file.DecryptedFilePath)
	if err != nil {
		return nil, errors.NewAppError("failed to load file data", err)
	}

	return data, nil
}

// GetEn retrieves the file data for a local file
func (uc *getLocalFileUseCase) GetEncryptedFileData(
	ctx context.Context,
	file *localfile.LocalFile,
) ([]byte, error) {
	// Validate inputs
	if file == nil {
		return nil, errors.NewAppError("file is required", nil)
	}

	// Load the file data
	data, err := uc.repository.LoadEncryptedFileDataAtFilePath(ctx, file.EncryptedFilePath)
	if err != nil {
		return nil, errors.NewAppError("failed to load file data", err)
	}

	return data, nil
}

// GetDecryptedThumbnail retrieves the thumbnail data for a local file
func (uc *getLocalFileUseCase) GetThumbnail(
	ctx context.Context,
	file *localfile.LocalFile,
) ([]byte, error) {
	// Validate inputs
	if file == nil {
		return nil, errors.NewAppError("file is required", nil)
	}

	// Load the thumbnail data
	data, err := uc.repository.LoadDecryptedFileDataAtFilePath(ctx, file.LocalThumbnailPath)
	if err != nil {
		return nil, errors.NewAppError("failed to load thumbnail data", err)
	}

	return data, nil
}
