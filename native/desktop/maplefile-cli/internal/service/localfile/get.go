// internal/service/localfile/get.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// GetOutput represents the result of getting a local file
type GetOutput struct {
	File     *localfile.LocalFile `json:"file"`
	FileData []byte               `json:"file_data,omitempty"`
}

// GetService defines the interface for getting local files
type GetService interface {
	GetByID(ctx context.Context, id string) (*GetOutput, error)
	GetWithData(ctx context.Context, id string) (*GetOutput, error)
	GetByEncryptedFileID(ctx context.Context, encryptedFileID string) (*GetOutput, error)
}

// getService implements the GetService interface
type getService struct {
	logger         *zap.Logger
	getFileUseCase uc.GetLocalFileUseCase
}

// NewGetService creates a new service for getting local files
func NewGetService(
	logger *zap.Logger,
	getFileUseCase uc.GetLocalFileUseCase,
) GetService {
	return &getService{
		logger:         logger,
		getFileUseCase: getFileUseCase,
	}
}

// GetByID retrieves a local file by ID
func (s *getService) GetByID(ctx context.Context, id string) (*GetOutput, error) {
	// Validate inputs
	if id == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	// Get file
	file, err := s.getFileUseCase.ByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return &GetOutput{
		File: file,
	}, nil
}

// GetWithData retrieves a local file and its data by ID
func (s *getService) GetWithData(ctx context.Context, id string) (*GetOutput, error) {
	// Validate inputs
	if id == "" {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	// Get file
	file, err := s.getFileUseCase.ByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	// Get file data
	fileData, err := s.getFileUseCase.GetFileData(ctx, file)
	if err != nil {
		return nil, err
	}

	return &GetOutput{
		File:     file,
		FileData: fileData,
	}, nil
}

// GetByEncryptedFileID retrieves a local file by encrypted file ID
func (s *getService) GetByEncryptedFileID(ctx context.Context, encryptedFileID string) (*GetOutput, error) {
	// Validate inputs
	if encryptedFileID == "" {
		return nil, errors.NewAppError("encrypted file ID is required", nil)
	}

	// Get file
	file, err := s.getFileUseCase.ByEncryptedFileID(ctx, encryptedFileID)
	if err != nil {
		return nil, err
	}

	return &GetOutput{
		File: file,
	}, nil
}
