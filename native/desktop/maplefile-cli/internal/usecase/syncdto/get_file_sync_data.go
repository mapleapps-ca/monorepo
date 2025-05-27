// internal/usecase/syncdto/get_file_sync_data.go
package syncdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

// GetFileSyncDataInput represents the input for getting file sync data
type GetFileSyncDataInput struct {
	Cursor *syncdto.SyncCursorDTO
	Limit  int64
}

// GetFileSyncDataUseCase defines the interface for getting file sync data from cloud
type GetFileSyncDataUseCase interface {
	Execute(ctx context.Context, input *GetFileSyncDataInput) (*syncdto.FileSyncResponseDTO, error)
}

// getFileSyncDataUseCase implements the GetFileSyncDataUseCase interface
type getFileSyncDataUseCase struct {
	logger     *zap.Logger
	repository syncdto.SyncDTORepository
}

// NewGetFileSyncDataUseCase creates a new use case for getting file sync data
func NewGetFileSyncDataUseCase(
	logger *zap.Logger,
	repository syncdto.SyncDTORepository,
) GetFileSyncDataUseCase {
	return &getFileSyncDataUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute retrieves file sync data from the cloud
func (uc *getFileSyncDataUseCase) Execute(ctx context.Context, input *GetFileSyncDataInput) (*syncdto.FileSyncResponseDTO, error) {
	// Validate input
	if input == nil {
		return nil, errors.NewAppError("get file sync data input is required", nil)
	}

	// Set default limit if not provided
	limit := input.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}

	// Validate limit is reasonable
	if limit > 1000 {
		return nil, errors.NewAppError("limit cannot exceed 1000", nil)
	}

	uc.logger.Debug("Getting file sync data from cloud",
		zap.Any("cursor", input.Cursor),
		zap.Int64("limit", limit))

	// Get file sync data from repository
	response, err := uc.repository.GetFileSyncDataFromCloud(ctx, input.Cursor, limit)
	if err != nil {
		uc.logger.Error("Failed to get file sync data from cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to get file sync data from cloud", err)
	}

	if response == nil {
		uc.logger.Warn("Received nil response from file sync data repository")
		return &syncdto.FileSyncResponseDTO{
			Files:   []syncdto.FileSyncItem{},
			HasMore: false,
		}, nil
	}

	uc.logger.Debug("Successfully retrieved file sync data",
		zap.Int("filesCount", len(response.Files)),
		zap.Bool("hasMore", response.HasMore))

	return response, nil
}
