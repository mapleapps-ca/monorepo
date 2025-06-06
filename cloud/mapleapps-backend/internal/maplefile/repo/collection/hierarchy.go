// cloud/backend/internal/maplefile/repo/collection/hierarchy.go
package collection

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) MoveCollection(
	ctx context.Context,
	collectionID,
	newParentID primitive.ObjectID,
	updatedAncestors []primitive.ObjectID,
	updatedPathSegments []string,
) error {
	// Verify the new parent exists
	exists, err := impl.CheckIfExistsByID(ctx, newParentID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("new parent collection not found")
	}

	// Verify this wouldn't create a cycle (collection can't be an ancestor of its new parent)
	// Get all ancestors of the new parent
	filter := bson.M{"_id": newParentID}
	var newParent dom_collection.Collection
	err = impl.Collection.FindOne(ctx, filter).Decode(&newParent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("new parent collection not found")
		}
		impl.Logger.Error("failed to retrieve new parent", zap.Any("error", err))
		return err
	}

	// Check if this collection is an ancestor of the new parent
	for _, ancestorID := range newParent.AncestorIDs {
		if ancestorID == collectionID {
			return errors.New("cannot move a collection to one of its descendants")
		}
	}

	// Update this collection with new parent and updated ancestry
	updateFilter := bson.M{"_id": collectionID}
	update := bson.M{
		"$set": bson.M{
			"parent_id":               newParentID,
			"ancestor_ids":            updatedAncestors,
			"encrypted_path_segments": updatedPathSegments,
			"modified_at":             time.Now(),
		},
	}

	_, err = impl.Collection.UpdateOne(ctx, updateFilter, update)
	if err != nil {
		impl.Logger.Error("failed to move collection", zap.Any("error", err))
		return err
	}

	// Update all descendants to have updated ancestors
	descendants, err := impl.FindDescendants(ctx, collectionID)
	if err != nil {
		impl.Logger.Error("failed to retrieve descendants for update", zap.Any("error", err))
		return err
	}

	for _, descendant := range descendants {
		// Calculate new ancestors for this descendant based on its existing relationship to the moved collection
		// We need to find which ancestors to keep (those after collectionID in the chain)
		var relativeAncestors []primitive.ObjectID
		foundMovedCollection := false
		for _, ancestor := range descendant.AncestorIDs {
			if ancestor == collectionID {
				foundMovedCollection = true
				continue // Skip the moved collection itself
			}

			if foundMovedCollection {
				// Keep only ancestors that come after the moved collection
				relativeAncestors = append(relativeAncestors, ancestor)
			}
		}

		// New ancestors = updated ancestors of moved collection + relative ancestors
		newAncestors := append(updatedAncestors, append([]primitive.ObjectID{collectionID}, relativeAncestors...)...)

		// Update this descendant
		descUpdate := bson.M{
			"$set": bson.M{
				"ancestor_ids": newAncestors,
				"modified_at":  time.Now(),
			},
		}

		_, err = impl.Collection.UpdateOne(ctx, bson.M{"_id": descendant.ID}, descUpdate)
		if err != nil {
			impl.Logger.Error("failed to update descendant ancestry",
				zap.Any("error", err),
				zap.Any("descendant_id", descendant.ID))
			// Continue updating other descendants even if one fails
		}
	}

	return nil
}
