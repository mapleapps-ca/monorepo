// internal/usecase/remotefile/fetch.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
)

// FetchRemoteFileUseCase defines the interface for fetching a remote file
type FetchRemoteFileUseCase interface {
	ByID(ctx context.Context, id primitive.ObjectID) (*remotefile.RemoteFile, error)
	ByEncryptedFileID(ctx context.Context, encryptedFileID string) (*remotefile.RemoteFile, error)
	DownloadFileData(ctx context.Context, id primitive.ObjectID) ([]byte, error)
}

// fetchRemoteFileUseCase implements the FetchRemoteFileUseCase interface
type fetchRemoteFileUseCase struct {
	logger     *zap.Logger
	repository remotefile.RemoteFileRepository
}

// NewFetchRemoteFileUseCase creates a new use case for fetching remote files
func NewFetchRemoteFileUseCase(
	logger *zap.Logger,
	repository remotefile.RemoteFileRepository,
) FetchRemoteFileUseCase {
	return &fetchRemoteFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// ByID fetches a remote file by ID
func (uc *fetchRemoteFileUseCase) ByID(
	ctx context.Context,
	id primitive.ObjectID,
) (*remotefile.RemoteFile, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Fetch the file from the repository
	file, err := uc.repository.Fetch(ctx, id)
	if err != nil {
		return nil, errors.NewAppError("failed to fetch remote file", err)
	}

	if file == nil {
		return nil, errors.NewAppError("remote file not found", nil)
	}

	return file, nil
}

// ByEncryptedFileID fetches a remote file by its encrypted file ID
func (uc *fetchRemoteFileUseCase) ByEncryptedFileID(
	ctx context.Context,
	encryptedFileID string,
) (*remotefile.RemoteFile, error) {
	// Validate inputs
	if encryptedFileID == "" {
		return nil, errors.NewAppError("encrypted file ID is required", nil)
	}

	// Fetch the file from the repository
	file, err := uc.repository.GetByEncryptedFileID(ctx, encryptedFileID)
	if err != nil {
		return nil, errors.NewAppError("failed to fetch remote file by encrypted file ID", err)
	}

	if file == nil {
		return nil, errors.NewAppError("remote file with the specified encrypted file ID not found", nil)
	}

	return file, nil
}

// DownloadFileData downloads the file data for a remote file
func (uc *fetchRemoteFileUseCase) DownloadFileData(
	ctx context.Context,
	id primitive.ObjectID,
) ([]byte, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Download the file data
	data, err := uc.repository.DownloadFile(ctx, id)
	if err != nil {
		return nil, errors.NewAppError("failed to download file data", err)
	}

	return data, nil
}
