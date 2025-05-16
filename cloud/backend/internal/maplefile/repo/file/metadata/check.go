// cloud/backend/internal/maplefile/repo/file/metadata/check.go
package metadata

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CheckIfExistsByID checks if a file exists by ID
func (impl fileMetadataRepositoryImpl) CheckIfExistsByID(id string) (bool, error) {
	ctx := context.Background()
	filter := bson.M{"id": id}

	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		impl.Logger.Error("database check if exists by ID error", zap.Any("error", err))
		return false, err
	}
	return count >= 1, nil
}

// CheckIfUserHasAccess checks if a user has access to a file
func (impl fileMetadataRepositoryImpl) CheckIfUserHasAccess(fileID string, userID string) (bool, error) {
	// First get the file to find its owner and collection ID
	file, err := impl.Get(fileID)
	if err != nil {
		impl.Logger.Error("database get file error", zap.Any("error", err))
		return false, err
	}
	if file == nil {
		return false, nil
	}

	// Direct access if user is the owner
	if file.OwnerID == userID {
		return true, nil
	}

	// We need to check collection sharing permissions
	// This would require a separate call to the collection repository
	// For now, we'll just return false, but in a real implementation
	// we would check if the collection is shared with the user

	return false, errors.New("collection access check not implemented")
}
