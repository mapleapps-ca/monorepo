// internal/service/collectionsyncer/find_by_remote_id.go
package collectionsyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

// FindByRemoteIDService defines the interface for finding a local collection by its remote ID
type FindByRemoteIDService interface {
	FindByRemoteID(ctx context.Context, remoteID primitive.ObjectID) (*localcollection.LocalCollection, error)
}

// findByRemoteIDService implements the FindByRemoteIDService interface
type findByRemoteIDService struct {
	logger          *zap.Logger
	localRepository localcollection.LocalCollectionRepository
}

// NewFindByRemoteIDService creates a new service for finding local collections by remote ID
func NewFindByRemoteIDService(
	logger *zap.Logger,
	localRepository localcollection.LocalCollectionRepository,
) FindByRemoteIDService {
	return &findByRemoteIDService{
		logger:          logger,
		localRepository: localRepository,
	}
}

// FindByRemoteID finds a local collection by its remote ID
func (s *findByRemoteIDService) FindByRemoteID(
	ctx context.Context,
	remoteID primitive.ObjectID,
) (*localcollection.LocalCollection, error) {
	// Validate inputs
	if remoteID.IsZero() {
		return nil, errors.NewAppError("remote ID is required", nil)
	}

	// Get all local collections
	// In a real implementation, this would likely use a more efficient filter
	// that directly queries by remoteID instead of getting all collections
	collections, err := s.localRepository.List(ctx, localcollection.LocalCollectionFilter{})
	if err != nil {
		return nil, errors.NewAppError("failed to list local collections", err)
	}

	// Find the one matching the remote ID
	// Assuming that LocalCollection has a RemoteID field to store the ID of the
	// corresponding remote collection
	for _, collection := range collections {
		// Note: This would need the LocalCollection struct to have a RemoteID field
		// This field might not exist yet in your current model
		if collection.RemoteID == remoteID {
			return collection, nil
		}
	}

	// Not found
	return nil, nil
}
