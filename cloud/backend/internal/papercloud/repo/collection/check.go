// github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/repo/collection/check.go
package collection

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/collection"
)

func (impl collectionRepositoryImpl) CheckIfExistsByID(id string) (bool, error) {
	ctx := context.Background()
	filter := bson.M{"id": id}

	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		impl.Logger.Error("database check if exists by ID error", zap.Any("error", err))
		return false, err
	}
	return count >= 1, nil
}

func (impl collectionRepositoryImpl) IsCollectionOwner(collectionID string, userID string) (bool, error) {
	ctx := context.Background()

	filter := bson.M{
		"id":       collectionID,
		"owner_id": userID,
	}

	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		impl.Logger.Error("database check if user is owner error", zap.Any("error", err))
		return false, err
	}

	return count >= 1, nil
}

func (impl collectionRepositoryImpl) CheckAccess(collectionID string, userID string, requiredPermission string) (bool, error) {
	ctx := context.Background()

	// First, check if the user is the owner
	isOwner, err := impl.IsCollectionOwner(collectionID, userID)
	if err != nil {
		return false, err
	}

	// Owners have all permissions
	if isOwner {
		return true, nil
	}

	// If not the owner, check if they're a member with sufficient permissions
	filter := bson.M{"id": collectionID}

	var collection dom_collection.Collection
	err = impl.Collection.FindOne(ctx, filter).Decode(&collection)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, errors.New("collection not found")
		}
		impl.Logger.Error("database get collection error", zap.Any("error", err))
		return false, err
	}

	// Check if user is a member and has sufficient permission
	for _, membership := range collection.Members {
		if membership.RecipientID == userID {
			// Found the membership, now check permission level
			switch requiredPermission {
			case dom_collection.CollectionPermissionReadOnly:
				// Any permission level is sufficient for read-only access
				return true, nil

			case dom_collection.CollectionPermissionReadWrite:
				// Need read-write or admin permission
				return membership.PermissionLevel == dom_collection.CollectionPermissionReadWrite ||
					membership.PermissionLevel == dom_collection.CollectionPermissionAdmin, nil

			case dom_collection.CollectionPermissionAdmin:
				// Need admin permission
				return membership.PermissionLevel == dom_collection.CollectionPermissionAdmin, nil

			default:
				// Unknown permission level requested
				impl.Logger.Warn("unknown permission level requested",
					zap.String("required_permission", requiredPermission))
				return false, errors.New("invalid permission level requested")
			}
		}
	}

	// User is not a member
	return false, nil
}

// Helper function to get a user's permission level for a collection
func (impl collectionRepositoryImpl) GetUserPermissionLevel(collectionID string, userID string) (string, error) {
	ctx := context.Background()

	// First check if user is the owner
	isOwner, err := impl.IsCollectionOwner(collectionID, userID)
	if err != nil {
		return "", err
	}

	if isOwner {
		// Owners have admin permissions
		return dom_collection.CollectionPermissionAdmin, nil
	}

	// Get the collection to check membership
	filter := bson.M{"id": collectionID}

	var collection dom_collection.Collection
	err = impl.Collection.FindOne(ctx, filter).Decode(&collection)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New("collection not found")
		}
		impl.Logger.Error("database get collection error", zap.Any("error", err))
		return "", err
	}

	// Look for the user's membership
	for _, membership := range collection.Members {
		if membership.RecipientID == userID {
			return membership.PermissionLevel, nil
		}
	}

	// User is not a member
	return "", errors.New("user is not a member of this collection")
}
