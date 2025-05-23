// internal/service/collectionsyncer/upload.go
package collectionsyncer

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
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
	localRepository  collection.CollectionRepository
	remoteRepository remotecollection.RemoteCollectionRepository
}

// NewUploadService creates a new service for uploading collections
func NewUploadService(
	logger *zap.Logger,
	localRepository collection.CollectionRepository,
	remoteRepository remotecollection.RemoteCollectionRepository,
) UploadService {
	return &uploadService{
		logger:           logger,
		localRepository:  localRepository,
		remoteRepository: remoteRepository,
	}
}

// Upload uploads a local collection to the cloud server
func (s *uploadService) Upload(ctx context.Context, localID string) (*UploadOutput, error) {
	// Validate inputs
	if localID == "" {
		return nil, errors.NewAppError("local collection ID is required", nil)
	}

	// Convert local ID string to ObjectID
	localObjectID, err := primitive.ObjectIDFromHex(localID)
	if err != nil {
		s.logger.Error("Invalid local ID format", zap.String("localID", localID), zap.Error(err))
		return nil, errors.NewAppError("invalid local ID format", err)
	}

	// Get the local collection
	collection, err := s.localRepository.GetByID(ctx, localObjectID)
	if err != nil {
		return nil, errors.NewAppError("failed to get local collection", err)
	}

	if collection == nil {
		return nil, errors.NewAppError("local collection not found", nil)
	}

	// Create input for the cloud collection creation
	input := &remotecollection.RemoteCreateCollectionRequest{
		EncryptedName:          collection.EncryptedName,
		Type:                   collection.Type,
		EncryptedCollectionKey: collection.EncryptedCollectionKey,
		EncryptedPathSegments:  collection.EncryptedPathSegments,
	}

	// Set parent ID if it's not zero
	if !collection.ParentID.IsZero() {
		input.ParentID = collection.ParentID
	}

	// Create or update the cloud collection
	var response *remotecollection.RemoteCollectionResponse

	// If we already have a cloud ID, update the existing cloud collection
	if !collection.CloudID.IsZero() {
		// In a real implementation, you'd have an update method on the cloud repository
		// For now, we'll just create a new one to complete the pattern
		response, err = s.remoteRepository.Create(ctx, input)
	} else {
		// Create a new cloud collection
		response, err = s.remoteRepository.Create(ctx, input)
	}

	if err != nil {
		return nil, errors.NewAppError("failed to create/update cloud collection", err)
	}

	// Update the local collection with sync info
	collection.CloudID = response.ID // Set the cloud ID reference
	collection.LastSyncedAt = time.Now()
	collection.IsModifiedLocally = false

	// Save the updated local collection
	err = s.localRepository.Save(ctx, collection)
	if err != nil {
		return nil, errors.NewAppError("failed to update local collection sync status", err)
	}

	return &UploadOutput{
		Collection: response,
	}, nil
}

// UploadAll uploads all locally modified collections to the cloud server
func (s *uploadService) UploadAll(ctx context.Context) (int, error) {
	// Create a filter for locally modified collections
	status := collection.SyncStatusModifiedLocally
	filter := collection.CollectionFilter{
		SyncStatus: &status,
	}

	// Get all modified local collections
	modifiedCollections, err := s.localRepository.List(ctx, filter)
	if err != nil {
		return 0, errors.NewAppError("failed to list modified local collections", err)
	}

	// Upload each collection
	successCount := 0
	for _, collection := range modifiedCollections {
		_, err := s.Upload(ctx, collection.ID.Hex())
		if err != nil {
			// Log error but continue with other collections
			s.logger.Error("Failed to upload collection",
				zap.String("collectionID", collection.ID.Hex()),
				zap.Error(err))
			continue
		}
		successCount++
	}

	return successCount, nil
}
