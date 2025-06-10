// cloud/mapleapps-backend/internal/maplefile/repo/collection/delete.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl *collectionRepositoryImpl) SoftDelete(ctx context.Context, id gocql.UUID) error {
	collection, err := impl.GetWithAnyState(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collection for soft delete: %w", err)
	}

	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	// Validate state transition
	if err := dom_collection.IsValidStateTransition(collection.State, dom_collection.CollectionStateDeleted); err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	// Update collection state
	collection.State = dom_collection.CollectionStateDeleted
	collection.ModifiedAt = time.Now()
	collection.Version++
	collection.TombstoneVersion = collection.Version
	collection.TombstoneExpiry = time.Now().Add(30 * 24 * time.Hour) // 30 days

	return impl.Update(ctx, collection)
}

func (impl *collectionRepositoryImpl) HardDelete(ctx context.Context, id gocql.UUID) error {
	collection, err := impl.GetWithAnyState(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collection for hard delete: %w", err)
	}

	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	batch := impl.Session.NewBatch(gocql.LoggedBatch)

	// 1. Delete from main table
	batch.Query(`DELETE FROM maplefile_collections_by_id WHERE id = ?`, id)

	// 2. Delete from user access table (owner)
	batch.Query(`DELETE FROM maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ? AND modified_at = ? AND collection_id = ?`,
		collection.OwnerID, collection.ModifiedAt, id)

	// 3. Delete member access entries
	for _, member := range collection.Members {
		batch.Query(`DELETE FROM maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND modified_at = ? AND collection_id = ?`,
			member.RecipientID, collection.ModifiedAt, id)
	}

	// 4. Delete from parent index
	parentID := collection.ParentID
	if !impl.isValidUUID(parentID) {
		parentID = impl.nullParentUUID()
	}
	batch.Query(`DELETE FROM maplefile_collections_by_parent_id_with_asc_created_at_and_asc_collection_id
		WHERE parent_id = ? AND created_at = ? AND collection_id = ?`,
		parentID, collection.CreatedAt, id)

	// 5. Delete from ancestor hierarchy
	ancestorEntries := impl.buildAncestorDepthEntries(id, collection.AncestorIDs)
	for _, entry := range ancestorEntries {
		batch.Query(`DELETE FROM maplefile_collections_by_ancestor_id_with_asc_depth_and_asc_collection_id
			WHERE ancestor_id = ? AND depth = ? AND collection_id = ?`,
			entry.AncestorID, entry.Depth, entry.CollectionID)
	}

	// 6. Delete from members table
	batch.Query(`DELETE FROM maplefile_collection_members_by_collection_id_and_recipient_id WHERE collection_id = ?`, id)

	// Execute batch
	if err := impl.Session.ExecuteBatch(batch.WithContext(ctx)); err != nil {
		impl.Logger.Error("failed to hard delete collection",
			zap.String("collection_id", id.String()),
			zap.Error(err))
		return fmt.Errorf("failed to hard delete collection: %w", err)
	}

	impl.Logger.Info("collection hard deleted successfully",
		zap.String("collection_id", id.String()))

	return nil
}
