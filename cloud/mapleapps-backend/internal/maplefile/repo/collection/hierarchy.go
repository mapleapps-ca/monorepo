// cloud/backend/internal/maplefile/repo/collection/hierarchy.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl *collectionRepositoryImpl) MoveCollection(
	ctx context.Context,
	collectionID,
	newParentID gocql.UUID,
	updatedAncestors []gocql.UUID,
	updatedPathSegments []string,
) error {
	// Get the collection to move
	collection, err := impl.Get(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	// Update collection hierarchy information
	oldParentID := collection.ParentID
	oldAncestorIDs := collection.AncestorIDs

	collection.ParentID = newParentID
	collection.AncestorIDs = updatedAncestors
	collection.ModifiedAt = time.Now()
	collection.Version++

	// Get all descendants that need to be updated
	descendants, err := impl.FindDescendants(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("failed to find descendants: %w", err)
	}

	batch := impl.Session.NewBatch(gocql.LoggedBatch)

	// 1. Update the moved collection
	if err := impl.updateCollectionInBatch(batch, collection, oldParentID, oldAncestorIDs); err != nil {
		return fmt.Errorf("failed to update moved collection: %w", err)
	}

	// 2. Update all descendants with new ancestor paths
	for _, descendant := range descendants {
		// Calculate new ancestor IDs for descendant
		newDescendantAncestors := append(updatedAncestors, collectionID)

		// Add intermediate ancestors between collection and descendant
		relativePath := impl.getRelativePath(collectionID, descendant.AncestorIDs)
		newDescendantAncestors = append(newDescendantAncestors, relativePath...)

		oldDescendantAncestors := descendant.AncestorIDs
		descendant.AncestorIDs = newDescendantAncestors
		descendant.ModifiedAt = time.Now()
		descendant.Version++

		if err := impl.updateCollectionInBatch(batch, descendant, descendant.ParentID, oldDescendantAncestors); err != nil {
			return fmt.Errorf("failed to update descendant %s: %w", descendant.ID.String(), err)
		}
	}

	// Execute the batch
	if err := impl.Session.ExecuteBatch(batch); err != nil {
		impl.Logger.Error("failed to move collection hierarchy",
			zap.String("collection_id", collectionID.String()),
			zap.String("new_parent_id", newParentID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to move collection hierarchy: %w", err)
	}

	impl.Logger.Info("collection moved successfully",
		zap.String("collection_id", collectionID.String()),
		zap.String("old_parent_id", oldParentID.String()),
		zap.String("new_parent_id", newParentID.String()))

	return nil
}

// Helper method to update collection in batch with proper index management
func (impl *collectionRepositoryImpl) updateCollectionInBatch(batch *gocql.Batch, collection *dom_collection.Collection, oldParentID gocql.UUID, oldAncestorIDs []gocql.UUID) error {
	// Serialize complex fields
	membersJSON, err := impl.serializeMembers(collection.Members)
	if err != nil {
		return fmt.Errorf("failed to serialize members: %w", err)
	}

	ancestorIDsJSON, err := impl.serializeAncestorIDs(collection.AncestorIDs)
	if err != nil {
		return fmt.Errorf("failed to serialize ancestor IDs: %w", err)
	}

	encryptedKeyJSON, err := impl.serializeEncryptedCollectionKey(collection.EncryptedCollectionKey)
	if err != nil {
		return fmt.Errorf("failed to serialize encrypted collection key: %w", err)
	}

	// Update main table
	batch.Query(`UPDATE maplefile_collections_by_id SET
		encrypted_name = ?, collection_type = ?, encrypted_collection_key = ?,
		members = ?, parent_id = ?, ancestor_ids = ?, modified_at = ?,
		modified_by_user_id = ?, version = ?, state = ?
		WHERE id = ?`,
		collection.EncryptedName, collection.CollectionType, encryptedKeyJSON,
		membersJSON, collection.ParentID, ancestorIDsJSON, collection.ModifiedAt,
		collection.ModifiedByUserID, collection.Version, collection.State, collection.ID)

	// Update parent index if parent changed
	if oldParentID != collection.ParentID {
		// Remove from old parent
		if impl.isValidUUID(oldParentID) {
			batch.Query(`DELETE FROM maplefile_collections_by_parent_id_with_asc_created_at_and_asc_id
				WHERE parent_id = ? AND created_at = ? AND collection_id = ?`,
				oldParentID, collection.CreatedAt, collection.ID)
		}

		// Add to new parent
		if impl.isValidUUID(collection.ParentID) {
			batch.Query(`INSERT INTO maplefile_collections_by_parent_id_with_asc_created_at_and_asc_id
				(parent_id, created_at, collection_id, state)
				VALUES (?, ?, ?, ?)`,
				collection.ParentID, collection.CreatedAt, collection.ID, collection.State)
		}
	}

	// Update ancestor depth index
	oldAncestorEntries := impl.buildAncestorDepthEntries(collection.ID, oldAncestorIDs)
	for _, entry := range oldAncestorEntries {
		batch.Query(`DELETE FROM maplefile_collections_by_ancestor_id_with_asc_depth_and_asc_collection_id
			WHERE ancestor_id = ? AND depth = ? AND collection_id = ?`,
			entry.AncestorID, entry.Depth, entry.CollectionID)
	}

	newAncestorEntries := impl.buildAncestorDepthEntries(collection.ID, collection.AncestorIDs)
	for _, entry := range newAncestorEntries {
		batch.Query(`INSERT INTO maplefile_collections_by_ancestor_id_with_asc_depth_and_asc_collection_id
			(ancestor_id, depth, collection_id, state)
			VALUES (?, ?, ?, ?)`,
			entry.AncestorID, entry.Depth, entry.CollectionID, collection.State)
	}

	return nil
}

// Helper method to get relative path between collection and descendant
func (impl *collectionRepositoryImpl) getRelativePath(collectionID gocql.UUID, descendantAncestors []gocql.UUID) []gocql.UUID {
	// Find where collectionID appears in descendant ancestors and return the path after it
	for i, ancestorID := range descendantAncestors {
		if ancestorID == collectionID {
			if i+1 < len(descendantAncestors) {
				return descendantAncestors[i+1:]
			}
			break
		}
	}
	return []gocql.UUID{}
}
