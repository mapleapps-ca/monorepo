// internal/service/collectionsyncer/download.go
package collectionsyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectionsyncer"
)

// DownloadOutput represents the result of downloading a collection
type DownloadOutput struct {
	Collection *localcollection.LocalCollection `json:"collection"`
}

// DownloadService defines the interface for downloading collections
type DownloadService interface {
	Download(ctx context.Context, remoteID string) (*DownloadOutput, error)
	DownloadAll(ctx context.Context) (int, error)
}

// downloadService implements the DownloadService interface
type downloadService struct {
	logger             *zap.Logger
	downloadUseCase    collectionsyncer.DownloadToLocalUseCase
	downloadAllUseCase collectionsyncer.DownloadAllRemoteUseCase
}

// NewDownloadService creates a new service for downloading collections
func NewDownloadService(
	logger *zap.Logger,
	downloadUseCase collectionsyncer.DownloadToLocalUseCase,
	downloadAllUseCase collectionsyncer.DownloadAllRemoteUseCase,
) DownloadService {
	return &downloadService{
		logger:             logger,
		downloadUseCase:    downloadUseCase,
		downloadAllUseCase: downloadAllUseCase,
	}
}

// Download downloads a remote collection to local storage
func (s *downloadService) Download(ctx context.Context, remoteID string) (*DownloadOutput, error) {
	// Validate input
	if remoteID == "" {
		s.logger.Error("remote collection ID is required")
		return nil, errors.NewAppError("remote collection ID is required", nil)
	}

	// Convert remote ID string to ObjectID
	remoteObjectID, err := primitive.ObjectIDFromHex(remoteID)
	if err != nil {
		s.logger.Error("invalid remote ID format", zap.String("remoteID", remoteID), zap.Error(err))
		return nil, errors.NewAppError("invalid remote ID format", err)
	}

	// Call the use case to download the collection
	collection, err := s.downloadUseCase.Execute(ctx, remoteObjectID)
	if err != nil {
		s.logger.Error("failed to download remote collection", zap.String("remoteID", remoteID), zap.Error(err))
		return nil, err
	}

	return &DownloadOutput{
		Collection: collection,
	}, nil
}

// DownloadAll downloads all remote collections to local storage
func (s *downloadService) DownloadAll(ctx context.Context) (int, error) {
	// Call the use case to download all collections
	count, err := s.downloadAllUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("failed to download all remote collections", zap.Error(err))
		return 0, err
	}

	s.logger.Info("all remote collections downloaded successfully", zap.Int("count", count))
	return count, nil
}
