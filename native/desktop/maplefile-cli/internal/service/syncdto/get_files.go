// internal/service/syncdto/get_files.go
package syncdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

// GetFilesInput represents the input for getting file sync data
type GetFilesInput struct {
	Cursor *syncdto.SyncCursorDTO `json:"cursor,omitempty"`
	Limit  int64                  `json:"limit,omitempty"`
}

// GetFilesOutput represents the result of getting file sync data
type GetFilesOutput struct {
	Response *syncdto.FileSyncResponseDTO `json:"response"`
	Message  string                       `json:"message"`
}

// GetFilesService defines the interface for getting file sync data
type GetFilesService interface {
	GetFileSyncData(ctx context.Context, input *GetFilesInput) (*GetFilesOutput, error)
}

// getFilesService implements the GetFilesService interface
type getFilesService struct {
	logger      *zap.Logger
	syncDTORepo syncdto.SyncDTORepository
}

// NewGetFilesService creates a new service for getting file sync data
func NewGetFilesService(
	logger *zap.Logger,
	syncDTORepo syncdto.SyncDTORepository,
) GetFilesService {
	logger = logger.Named("GetFilesService")
	return &getFilesService{
		logger:      logger,
		syncDTORepo: syncDTORepo,
	}
}

// GetFileSyncData retrieves file sync data from the cloud
func (s *getFilesService) GetFileSyncData(ctx context.Context, input *GetFilesInput) (*GetFilesOutput, error) {
	// Set default values
	if input == nil {
		input = &GetFilesInput{}
	}

	if input.Limit <= 0 {
		input.Limit = 100 // Default limit
	}

	s.logger.Debug("⬇️ Getting file sync data from cloud",
		zap.Any("cursor", input.Cursor),
		zap.Int64("limit", input.Limit))

	// Get file sync data from repository
	response, err := s.syncDTORepo.GetFileSyncDataFromCloud(ctx, input.Cursor, input.Limit)
	if err != nil {
		s.logger.Error("💥 failed to get file sync data from cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to get file sync data from cloud", err)
	}

	if response == nil {
		s.logger.Warn("⚠️ received nil response from file sync data")
		return nil, errors.NewAppError("received empty response from cloud", nil)
	}

	s.logger.Info("✅ Successfully retrieved file sync data",
		zap.Int("files_count", len(response.Files)),
		zap.Bool("has_more", response.HasMore))

	message := "File sync data retrieved successfully"
	if len(response.Files) == 0 {
		message = "No file changes found since last sync"
	}

	return &GetFilesOutput{
		Response: response,
		Message:  message,
	}, nil
}
