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
	GetByID(ctx context.Context, id primitive.ObjectID) (*GetOutput, error)
	GetWithData(ctx context.Context, id primitive.ObjectID) (*GetOutput, error)
	GetByRemoteID(ctx context.Context, remoteID primitive.ObjectID) (*GetOutput, error)
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
func (s *getService) GetByID(ctx context.Context, id primitive.ObjectID) (*GetOutput, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("ID is required", nil)
	}

	// Get file
	file, err := s.getFileUseCase.ByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &GetOutput{
		File: file,
	}, nil
}

// GetWithData retrieves a local file and its data by ID
func (s *getService) GetWithData(ctx context.Context, id primitive.ObjectID) (*GetOutput, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("ID is required", nil)
	}

	// Get file
	file, err := s.getFileUseCase.ByID(ctx, id)
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
func (s *getService) GetByRemoteID(ctx context.Context, remoteID primitive.ObjectID) (*GetOutput, error) {
	// Validate inputs
	if remoteID.IsZero() {
		return nil, errors.NewAppError("remote ID is required", nil)
	}

	// Get file
	file, err := s.getFileUseCase.ByRemoteID(ctx, remoteID)
	if err != nil {
		return nil, err
	}

	return &GetOutput{
		File: file,
	}, nil
}
