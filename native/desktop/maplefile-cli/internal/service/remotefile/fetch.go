// internal/service/remotefile/fetch.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotefile"
)

// FetchOutput represents the result of fetching a remote file
type FetchOutput struct {
	File     *remotefile.RemoteFile `json:"file"`
	FileData []byte                 `json:"file_data,omitempty"`
}

// FetchService defines the interface for fetching remote files
type FetchService interface {
	FetchByID(ctx context.Context, id string) (*FetchOutput, error)
	FetchWithData(ctx context.Context, id string) (*FetchOutput, error)
	FetchByEncryptedFileID(ctx context.Context, encryptedFileID string) (*FetchOutput, error)
}

// fetchService implements the FetchService interface
type fetchService struct {
	logger              *zap.Logger
	fetchFileUseCase    uc.FetchRemoteFileUseCase
	downloadFileUseCase uc.DownloadRemoteFileUseCase
}

// NewFetchService creates a new service for fetching remote files
func NewFetchService(
	logger *zap.Logger,
	fetchFileUseCase uc.FetchRemoteFileUseCase,
	downloadFileUseCase uc.DownloadRemoteFileUseCase,
) FetchService {
	return &fetchService{
		logger:              logger,
		fetchFileUseCase:    fetchFileUseCase,
		downloadFileUseCase: downloadFileUseCase,
	}
}

// FetchByID fetches a remote file by ID
func (s *fetchService) FetchByID(ctx context.Context, id string) (*FetchOutput, error) {
	// Validate inputs
	if id == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	// Fetch the file
	file, err := s.fetchFileUseCase.ByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return &FetchOutput{
		File: file,
	}, nil
}

// FetchWithData fetches a remote file and its data by ID
func (s *fetchService) FetchWithData(ctx context.Context, id string) (*FetchOutput, error) {
	// Validate inputs
	if id == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	// Fetch the file
	file, err := s.fetchFileUseCase.ByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	// Download the file data
	fileData, err := s.fetchFileUseCase.DownloadFileData(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return &FetchOutput{
		File:     file,
		FileData: fileData,
	}, nil
}

// FetchByEncryptedFileID fetches a remote file by encrypted file ID
func (s *fetchService) FetchByEncryptedFileID(ctx context.Context, encryptedFileID string) (*FetchOutput, error) {
	// Validate inputs
	if encryptedFileID == "" {
		return nil, errors.NewAppError("encrypted file ID is required", nil)
	}

	// Fetch the file
	file, err := s.fetchFileUseCase.ByEncryptedFileID(ctx, encryptedFileID)
	if err != nil {
		return nil, err
	}

	return &FetchOutput{
		File: file,
	}, nil
}
