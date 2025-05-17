// internal/service/collectionsyncer/upload.go
package collectionsyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectionsyncer"
)

// UploadOutput represents the result of uploading a collection
type UploadOutput struct {
	Collection *remotecollection.RemoteCollectionResponse `json:"collection"`
}

// UploadService defines the interface for uploading collections
type UploadService interface {
	Upload(ctx context.Context, localID string) (*UploadOutput, error)
	UploadAll(ctx context.Context) (int, error)
}

// uploadService implements the UploadService interface
type uploadService struct {
	logger           *zap.Logger
	uploadUseCase    collectionsyncer.UploadToRemoteUseCase
	uploadAllUseCase collectionsyncer.SyncAllLocalToRemoteUseCase
}

// NewUploadService creates a new service for uploading collections
func NewUploadService(
	logger *zap.Logger,
	uploadUseCase collectionsyncer.UploadToRemoteUseCase,
	uploadAllUseCase collectionsyncer.SyncAllLocalToRemoteUseCase,
) UploadService {
	return &uploadService{
		logger:           logger,
		uploadUseCase:    uploadUseCase,
		uploadAllUseCase: uploadAllUseCase,
	}
}

// Upload uploads a local collection to the remote server
func (s *uploadService) Upload(ctx context.Context, localID string) (*UploadOutput, error) {
	// Validate input
	if localID == "" {
		s.logger.Error("local collection ID is required")
		return nil, errors.NewAppError("local collection ID is required", nil)
	}

	// Convert local ID string to ObjectID
	localObjectID, err := primitive.ObjectIDFromHex(localID)
	if err != nil {
		s.logger.Error("invalid local ID format", zap.String("localID", localID), zap.Error(err))
		return nil, errors.NewAppError("invalid local ID format", err)
	}

	// Call the use case to upload the collection
	response, err := s.uploadUseCase.Execute(ctx, localObjectID)
	if err != nil {
		s.logger.Error("failed to upload local collection", zap.String("localID", localID), zap.Error(err))
		return nil, err
	}

	return &UploadOutput{
		Collection: response,
	}, nil
}

// UploadAll uploads all local collections to the remote server
func (s *uploadService) UploadAll(ctx context.Context) (int, error) {
	// Call the use case to upload all collections
	count, err := s.uploadAllUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("failed to upload all local collections", zap.Error(err))
		return 0, err
	}

	s.logger.Info("all local collections uploaded successfully", zap.Int("count", count))
	return count, nil
}
