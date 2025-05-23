// internal/service/collectionsyncer/find_by_cloud_id.go
package collectionsyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

// FindByCloudIDService defines the interface for finding a local collection by its cloud ID
type FindByCloudIDService interface {
	FindByCloudID(ctx context.Context, cloudID primitive.ObjectID) (*collection.Collection, error)
}

// findByCloudIDService implements the FindByCloudIDService interface
type findByCloudIDService struct {
	logger          *zap.Logger
	localRepository collection.CollectionRepository
}

// NewFindByCloudIDService creates a new service for finding local collections by cloud ID
func NewFindByCloudIDService(
	logger *zap.Logger,
	localRepository collection.CollectionRepository,
) FindByCloudIDService {
	return &findByCloudIDService{
		logger:          logger,
		localRepository: localRepository,
	}
}

// FindByCloudID finds a local collection by its cloud ID
func (s *findByCloudIDService) FindByCloudID(
	ctx context.Context,
	cloudID primitive.ObjectID,
) (*collection.Collection, error) {
	// Validate inputs
	if cloudID.IsZero() {
		return nil, errors.NewAppError("cloud ID is required", nil)
	}

	// Get all local collections
	// In a real implementation, this would likely use a more efficient filter
	// that directly queries by cloudID instead of getting all collections
	collections, err := s.localRepository.List(ctx, collection.CollectionFilter{})
	if err != nil {
		return nil, errors.NewAppError("failed to list local collections", err)
	}

	// Find the one matching the cloud ID
	// Assuming that Collection has a CloudID field to store the ID of the
	// corresponding cloud collection
	for _, collection := range collections {
		// Note: This would need the Collection struct to have a CloudID field
		// This field might not exist yet in your current model
		if collection.CloudID == cloudID {
			return collection, nil
		}
	}

	// Not found
	return nil, nil
}
