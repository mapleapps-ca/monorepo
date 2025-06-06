// cloud/backend/internal/maplefile/repo/collection/check.go
package collection

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) CheckIfExistsByID(ctx context.Context, id gocql.UUID) (bool, error) {
	filter := bson.M{"_id": id}

	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		impl.Logger.Error("database check if exists by ID error", zap.Any("error", err))
		return false, err
	}
	return count >= 1, nil
}

func (impl collectionRepositoryImpl) IsCollectionOwner(ctx context.Context, collectionID, userID gocql.UUID) (bool, error) {
	filter := bson.M{
		"_id":      collectionID,
		"owner_id": userID,
	}

	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		impl.Logger.Error("database check if user is owner error", zap.Any("error", err))
		return false, err
	}

	return count >= 1, nil
}

func (impl collectionRepositoryImpl) CheckAccess(ctx context.Context, collectionID, userID gocql.UUID, requiredPermission string) (bool, error) {
	// First, check if the user is the owner
	isOwner, err := impl.IsCollectionOwner(ctx, collectionID, userID)
	if err != nil {
		return false, err
	}

	// Owners have all permissions
	if isOwner {
		return true, nil
	}

	// If not the owner, check if they're a member with sufficient permissions
	filter := bson.M{"_id": collectionID}

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
func (impl collectionRepositoryImpl) GetUserPermissionLevel(ctx context.Context, collectionID, userID gocql.UUID) (string, error) {
	// First check if user is the owner
	isOwner, err := impl.IsCollectionOwner(ctx, collectionID, userID)
	if err != nil {
		return "", err
	}

	if isOwner {
		// Owners have admin permissions
		return dom_collection.CollectionPermissionAdmin, nil
	}

	// Get the collection to check membership
	filter := bson.M{"_id": collectionID}

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
