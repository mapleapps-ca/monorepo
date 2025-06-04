// cloud/backend/internal/maplefile/repo/collection/share.go
package collection

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

// AddMember adds a new member to a collection, giving them access
func (impl collectionRepositoryImpl) AddMember(ctx context.Context, collectionID primitive.ObjectID, membership *dom_collection.CollectionMembership) error {
	// Ensure the collection exists
	exists, err := impl.CheckIfExistsByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("collection not found")
	}

	// Generate ID if not provided
	if membership.ID.IsZero() {
		membership.ID = primitive.NewObjectID()
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
		"_id":                  collectionID,
		"members.recipient_id": membership.RecipientID,
	}

	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		impl.Logger.Error("database check existing membership error", zap.Any("error", err))
		return err
	}

	if count > 0 {
		impl.Logger.Warn("user is already a member of this collection",
			zap.Any("collection_id", collectionID),
			zap.Any("recipient_id", membership.RecipientID))

		// Update the existing membership instead
		return impl.UpdateMemberPermission(ctx, collectionID, membership.RecipientID, membership.PermissionLevel)
	}

	// Add the new membership
	update := bson.M{
		"$push": bson.M{"members": membership},
		"$set":  bson.M{"modified_at": time.Now()},
	}

	_, err = impl.Collection.UpdateOne(ctx, bson.M{"_id": collectionID}, update)
	if err != nil {
		impl.Logger.Error("database add member error",
			zap.Any("error", err),
			zap.Any("collection_id", collectionID),
			zap.Any("recipient_id", membership.RecipientID))
		return err
	}

	return nil
}

// RemoveMember removes a member from a collection, revoking their access
func (impl collectionRepositoryImpl) RemoveMember(ctx context.Context, collectionID, recipientID primitive.ObjectID) error {
	filter := bson.M{"_id": collectionID}

	// Pull the member from the members array
	update := bson.M{
		"$pull": bson.M{
			"members": bson.M{
				"recipient_id": recipientID,
			},
		},
		"$set": bson.M{"modified_at": time.Now()},
	}

	result, err := impl.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database remove member error",
			zap.Any("error", err),
			zap.Any("collection_id", collectionID),
			zap.Any("recipient_id", recipientID))
		return err
	}

	if result.ModifiedCount == 0 {
		impl.Logger.Warn("no membership was removed, may not exist",
			zap.Any("collection_id", collectionID),
			zap.Any("recipient_id", recipientID))
	}

	return nil
}

// UpdateMemberPermission updates a member's permission level
func (impl collectionRepositoryImpl) UpdateMemberPermission(ctx context.Context, collectionID, recipientID primitive.ObjectID, newPermission string) error {
	// Ensure the permission level is valid
	if newPermission == "" {
		newPermission = dom_collection.CollectionPermissionReadOnly
	}

	// Find the collection and the specific membership to update
	filter := bson.M{
		"_id":                  collectionID,
		"members.recipient_id": recipientID,
	}

	// Update the permission level in the specific array element
	update := bson.M{
		"$set": bson.M{
			"members.$.permission_level": newPermission,
			"modified_at":                time.Now(),
		},
	}

	result, err := impl.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database update member permission error",
			zap.Any("error", err),
			zap.Any("collection_id", collectionID),
			zap.Any("recipient_id", recipientID),
			zap.String("new_permission", newPermission))
		return err
	}

	if result.ModifiedCount == 0 {
		impl.Logger.Warn("no membership was updated, may not exist",
			zap.Any("collection_id", collectionID),
			zap.Any("recipient_id", recipientID))
		return errors.New("membership not found")
	}

	return nil
}

// GetCollectionMembership retrieves a specific membership from a collection
func (impl collectionRepositoryImpl) GetCollectionMembership(ctx context.Context, collectionID, recipientID primitive.ObjectID) (*dom_collection.CollectionMembership, error) {
	// Find the collection
	filter := bson.M{"_id": collectionID}

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

// AddMemberToHierarchy adds a user to a collection and all of its descendants
func (impl collectionRepositoryImpl) AddMemberToHierarchy(ctx context.Context, rootID primitive.ObjectID, membership *dom_collection.CollectionMembership) error {
	// First add the member to the root collection
	rootMembership := *membership
	rootMembership.IsInherited = false
	rootMembership.ID = primitive.NewObjectID()

	err := impl.AddMember(ctx, rootID, &rootMembership)
	if err != nil {
		return err
	}

	// Then find all descendant collections
	descendants, err := impl.FindDescendants(ctx, rootID)
	if err != nil {
		return err
	}

	// For each descendant, we need to get its collection key and encrypt it for the recipient
	for _, descendant := range descendants {
		// Get the descendant collection to access its encrypted key
		descendantCollection, err := impl.Get(ctx, descendant.ID)
		if err != nil {
			impl.Logger.Error("failed to get descendant collection for key encryption",
				zap.Any("error", err),
				zap.Any("descendant_id", descendant.ID))
			continue
		}

		if descendantCollection == nil {
			impl.Logger.Warn("descendant collection not found, skipping",
				zap.Any("descendant_id", descendant.ID))
			continue
		}

		// Create membership for descendant with the descendant's encrypted key
		// Note: We cannot re-encrypt the key here because we don't have access to:
		// 1. The recipient's public key
		// 2. The decrypted collection key
		// 3. The owner's master key
		//
		// Therefore, we need to pass the recipient's encrypted key from the service layer
		childMembership := dom_collection.CollectionMembership{
			ID:              primitive.NewObjectID(),
			CollectionID:    descendant.ID,
			RecipientID:     membership.RecipientID,
			RecipientEmail:  membership.RecipientEmail,
			GrantedByID:     membership.GrantedByID,
			PermissionLevel: membership.PermissionLevel,
			CreatedAt:       membership.CreatedAt,
			IsInherited:     true,
			InheritedFromID: rootID,
			// ❌ PROBLEM: We still need the correct encrypted key for this specific collection
			// This requires refactoring the service layer to encrypt each collection's key separately
			EncryptedCollectionKey: membership.EncryptedCollectionKey, // This is wrong but we'll fix it in Solution 2
		}

		if err := impl.AddMember(ctx, descendant.ID, &childMembership); err != nil {
			impl.Logger.Error("failed to add inherited membership to descendant",
				zap.Any("error", err),
				zap.Any("root_id", rootID),
				zap.Any("descendant_id", descendant.ID))
			// Continue with other descendants even if one fails
			continue
		}
	}

	return nil
}

// RemoveMemberFromHierarchy removes a user from a collection and all of its descendants
func (impl collectionRepositoryImpl) RemoveMemberFromHierarchy(ctx context.Context, rootID, recipientID primitive.ObjectID) error {
	// First remove the member from the root collection
	err := impl.RemoveMember(ctx, rootID, recipientID)
	if err != nil {
		return err
	}

	// Then find all descendant collections
	descendants, err := impl.FindDescendants(ctx, rootID)
	if err != nil {
		return err
	}

	// For each descendant, remove the inherited membership
	for _, descendant := range descendants {
		// Only remove memberships that were inherited from this root
		filter := bson.M{
			"_id": descendant.ID,
			"members": bson.M{
				"$elemMatch": bson.M{
					"recipient_id":      recipientID,
					"is_inherited":      true,
					"inherited_from_id": rootID,
				},
			},
		}

		update := bson.M{
			"$pull": bson.M{
				"members": bson.M{
					"recipient_id":      recipientID,
					"is_inherited":      true,
					"inherited_from_id": rootID,
				},
			},
			"$set": bson.M{"modified_at": time.Now()},
		}

		_, err := impl.Collection.UpdateOne(ctx, filter, update)
		if err != nil {
			impl.Logger.Error("failed to remove inherited membership from descendant",
				zap.Any("error", err),
				zap.Any("root_id", rootID),
				zap.Any("descendant_id", descendant.ID))
			// Continue with other descendants even if one fails
			continue
		}
	}

	return nil
}
