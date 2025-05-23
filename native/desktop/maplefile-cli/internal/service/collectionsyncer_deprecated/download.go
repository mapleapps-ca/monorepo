// internal/service/collectionsyncer/download.go
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

// DownloadOutput represents the result of downloading a collection
type DownloadOutput struct {
	Collection *collection.Collection `json:"collection"`
}

// DownloadService defines the interface for downloading collections
type DownloadService interface {
	Download(ctx context.Context, cloudID string) (*DownloadOutput, error)
	DownloadAll(ctx context.Context) (int, error)
}

// downloadService implements the DownloadService interface
type downloadService struct {
	logger               *zap.Logger
	localRepository      collection.CollectionRepository
	remoteRepository     remotecollection.RemoteCollectionRepository
	findByCloudIDService FindByCloudIDService
}

// NewDownloadService creates a new service for downloading collections
func NewDownloadService(
	logger *zap.Logger,
	localRepository collection.CollectionRepository,
	remoteRepository remotecollection.RemoteCollectionRepository,
	findByCloudIDService FindByCloudIDService,
) DownloadService {
	return &downloadService{
		logger:               logger,
		localRepository:      localRepository,
		remoteRepository:     remoteRepository,
		findByCloudIDService: findByCloudIDService,
	}
}

// Download downloads a cloud collection to local storage
func (s *downloadService) Download(ctx context.Context, cloudID string) (*DownloadOutput, error) {
	// Validate inputs
	if cloudID == "" {
		return nil, errors.NewAppError("cloud collection ID is required", nil)
	}

	// Convert cloud ID string to ObjectID
	remoteObjectID, err := primitive.ObjectIDFromHex(cloudID)
	if err != nil {
		s.logger.Error("Invalid cloud ID format", zap.String("cloudID", cloudID), zap.Error(err))
		return nil, errors.NewAppError("invalid cloud ID format", err)
	}

	// Check if we already have a local copy
	existingLocal, err := s.findByCloudIDService.FindByCloudID(ctx, remoteObjectID)
	if err != nil {
		return nil, errors.NewAppError("failed to check for existing local copy", err)
	}

	// Fetch the cloud collection
	remoteCollection, err := s.remoteRepository.Fetch(ctx, remoteObjectID)
	if err != nil {
		return nil, errors.NewAppError("failed to fetch cloud collection", err)
	}

	if remoteCollection == nil {
		return nil, errors.NewAppError("cloud collection not found", nil)
	}

	// If we have a local copy, update it
	if existingLocal != nil {
		s.logger.Debug("Updating existing local collection",
			zap.String("localID", existingLocal.ID.Hex()),
			zap.String("cloudID", cloudID))

		// Update fields from cloud
		existingLocal.EncryptedName = remoteCollection.EncryptedName
		existingLocal.Type = remoteCollection.Type
		existingLocal.EncryptedCollectionKey = remoteCollection.EncryptedCollectionKey
		existingLocal.EncryptedPathSegments = remoteCollection.EncryptedPathSegments
		existingLocal.LastSyncedAt = time.Now()
		existingLocal.IsModifiedLocally = false

		// Save the updated collection
		err = s.localRepository.Save(ctx, existingLocal)
		if err != nil {
			return nil, errors.NewAppError("failed to update local collection", err)
		}

		return &DownloadOutput{
			Collection: existingLocal,
		}, nil
	}

	// Create a new local collection
	collection := &collection.Collection{
		ID:                     primitive.NewObjectID(),
		CloudID:                remoteCollection.ID, // Store the cloud ID for future sync operations
		OwnerID:                remoteCollection.OwnerID,
		EncryptedName:          remoteCollection.EncryptedName,
		Type:                   remoteCollection.Type,
		ParentID:               remoteCollection.ParentID,
		AncestorIDs:            remoteCollection.AncestorIDs,
		EncryptedPathSegments:  remoteCollection.EncryptedPathSegments,
		EncryptedCollectionKey: remoteCollection.EncryptedCollectionKey,
		CreatedAt:              time.Now(),
		ModifiedAt:             time.Now(),
		LastSyncedAt:           time.Now(),
		IsModifiedLocally:      false,
	}

	// Create the local collection
	err = s.localRepository.Create(ctx, collection)
	if err != nil {
		return nil, errors.NewAppError("failed to create local collection", err)
	}

	return &DownloadOutput{
		Collection: collection,
	}, nil
}

// DownloadAll downloads all cloud collections to local storage
func (s *downloadService) DownloadAll(ctx context.Context) (int, error) {
	// Create an empty filter to get all cloud collections
	filter := remotecollection.CollectionFilter{}

	// Get all cloud collections
	remoteCollections, err := s.remoteRepository.List(ctx, filter)
	if err != nil {
		return 0, errors.NewAppError("failed to list cloud collections", err)
	}

	// Download each collection
	successCount := 0
	for _, collection := range remoteCollections {
		_, err := s.Download(ctx, collection.ID.Hex())
		if err != nil {
			// Log error but continue with other collections
			s.logger.Error("Failed to download collection",
				zap.String("collectionID", collection.ID.Hex()),
				zap.Error(err))
			continue
		}
		successCount++
	}

	return successCount, nil
}
