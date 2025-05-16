// github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/collection/share.go
package collection

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

// AddMember adds a new member to a collection, giving them access
func (impl collectionRepositoryImpl) AddMember(collectionID string, membership *dom_collection.CollectionMembership) error {
	ctx := context.Background()

	// Ensure the collection exists
	exists, err := impl.CheckIfExistsByID(collectionID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("collection not found")
	}

	// Set the collection ID and created time
	membership.CollectionID = collectionID
	if membership.CreatedAt.IsZero() {
		membership.CreatedAt = time.Now()
	}

	// Ensure the permission level is valid
	if membership.PermissionLevel == "" {
		membership.PermissionLevel = dom_collection.CollectionPermissionReadOnly
	}

	// Check if user is already a member of this collection
	filter := bson.M{
		"id":                   collectionID,
		"members.recipient_id": membership.RecipientID,
	}

	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		impl.Logger.Error("database check existing membership error", zap.Any("error", err))
		return err
	}

	if count > 0 {
		impl.Logger.Warn("user is already a member of this collection",
			zap.String("collection_id", collectionID),
			zap.String("recipient_id", membership.RecipientID))

		// Update the existing membership instead
		return impl.UpdateMemberPermission(collectionID, membership.RecipientID, membership.PermissionLevel)
	}

	// Add the new membership
	update := bson.M{
		"$push": bson.M{"members": membership},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err = impl.Collection.UpdateOne(ctx, bson.M{"id": collectionID}, update)
	if err != nil {
		impl.Logger.Error("database add member error",
			zap.Any("error", err),
			zap.String("collection_id", collectionID),
			zap.String("recipient_id", membership.RecipientID))
		return err
	}

	return nil
}

// RemoveMember removes a member from a collection, revoking their access
func (impl collectionRepositoryImpl) RemoveMember(collectionID string, recipientID string) error {
	ctx := context.Background()

	filter := bson.M{"id": collectionID}

	// Pull the member from the members array
	update := bson.M{
		"$pull": bson.M{
			"members": bson.M{
				"recipient_id": recipientID,
			},
		},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := impl.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database remove member error",
			zap.Any("error", err),
			zap.String("collection_id", collectionID),
			zap.String("recipient_id", recipientID))
		return err
	}

	if result.ModifiedCount == 0 {
		impl.Logger.Warn("no membership was removed, may not exist",
			zap.String("collection_id", collectionID),
			zap.String("recipient_id", recipientID))
	}

	return nil
}

// UpdateMemberPermission updates a member's permission level
func (impl collectionRepositoryImpl) UpdateMemberPermission(collectionID string, recipientID string, newPermission string) error {
	ctx := context.Background()

	// Ensure the permission level is valid
	if newPermission == "" {
		newPermission = dom_collection.CollectionPermissionReadOnly
	}

	// Find the collection and the specific membership to update
	filter := bson.M{
		"id":                   collectionID,
		"members.recipient_id": recipientID,
	}

	// Update the permission level in the specific array element
	update := bson.M{
		"$set": bson.M{
			"members.$.permission_level": newPermission,
			"updated_at":                 time.Now(),
		},
	}

	result, err := impl.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database update member permission error",
			zap.Any("error", err),
			zap.String("collection_id", collectionID),
			zap.String("recipient_id", recipientID),
			zap.String("new_permission", newPermission))
		return err
	}

	if result.ModifiedCount == 0 {
		impl.Logger.Warn("no membership was updated, may not exist",
			zap.String("collection_id", collectionID),
			zap.String("recipient_id", recipientID))
		return errors.New("membership not found")
	}

	return nil
}

// GetCollectionMembership retrieves a specific membership from a collection
func (impl collectionRepositoryImpl) GetCollectionMembership(collectionID string, recipientID string) (*dom_collection.CollectionMembership, error) {
	ctx := context.Background()

	// Find the collection
	filter := bson.M{"id": collectionID}

	var collection dom_collection.Collection
	err := impl.Collection.FindOne(ctx, filter).Decode(&collection)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("collection not found")
		}
		impl.Logger.Error("database get collection error", zap.Any("error", err))
		return nil, err
	}

	// Find the membership
	for _, membership := range collection.Members {
		if membership.RecipientID == recipientID {
			return &membership, nil
		}
	}

	return nil, errors.New("membership not found")
}
