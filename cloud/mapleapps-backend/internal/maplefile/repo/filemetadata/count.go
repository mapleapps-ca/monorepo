// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/count.go
package filemetadata

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

// CountFilesByUser counts all active files accessible to the user
// accessibleCollectionIDs should include all collections the user owns or has access to
func (impl *fileMetadataRepositoryImpl) CountFilesByUser(ctx context.Context, userID gocql.UUID, accessibleCollectionIDs []gocql.UUID) (int, error) {
	if len(accessibleCollectionIDs) == 0 {
		// No accessible collections, return 0
		impl.Logger.Debug("no accessible collections provided for file count",
			zap.String("user_id", userID.String()))
		return 0, nil
	}

	// Create a map for efficient collection access checking
	accessibleCollections := make(map[gocql.UUID]bool)
	for _, cid := range accessibleCollectionIDs {
		accessibleCollections[cid] = true
	}

	// Query files for the user using the user sync table
	query := `SELECT file_id, collection_id, state FROM mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
		WHERE user_id = ?`

	iter := impl.Session.Query(query, userID).WithContext(ctx).Iter()

	count := 0
	var fileID, collectionID gocql.UUID
	var state string

	for iter.Scan(&fileID, &collectionID, &state) {
		// Only count files from accessible collections
		if !accessibleCollections[collectionID] {
			continue
		}

		// Only count active files
		if state != dom_file.FileStateActive {
			continue
		}

		count++
	}

	if err := iter.Close(); err != nil {
		impl.Logger.Error("failed to count files by user",
			zap.String("user_id", userID.String()),
			zap.Int("accessible_collections_count", len(accessibleCollectionIDs)),
			zap.Error(err))
		return 0, fmt.Errorf("failed to count files by user: %w", err)
	}

	impl.Logger.Debug("counted files by user successfully",
		zap.String("user_id", userID.String()),
		zap.Int("accessible_collections_count", len(accessibleCollectionIDs)),
		zap.Int("file_count", count))

	return count, nil
}

// CountFilesByOwner counts all active files owned by the user (alternative approach)
func (impl *fileMetadataRepositoryImpl) CountFilesByOwner(ctx context.Context, ownerID gocql.UUID) (int, error) {
	// Query files owned by the user using the owner table
	query := `SELECT file_id, state FROM mapleapps.maplefile_files_by_owner_id_with_desc_modified_at_and_asc_file_id
		WHERE owner_id = ?`

	iter := impl.Session.Query(query, ownerID).WithContext(ctx).Iter()

	count := 0
	var fileID gocql.UUID
	var state string

	for iter.Scan(&fileID, &state) {
		// Only count active files
		if state != dom_file.FileStateActive {
			continue
		}

		count++
	}

	if err := iter.Close(); err != nil {
		impl.Logger.Error("failed to count files by owner",
			zap.String("owner_id", ownerID.String()),
			zap.Error(err))
		return 0, fmt.Errorf("failed to count files by owner: %w", err)
	}

	impl.Logger.Debug("counted files by owner successfully",
		zap.String("owner_id", ownerID.String()),
		zap.Int("file_count", count))

	return count, nil
}

// CountFilesByCollection counts active files in a specific collection
func (impl *fileMetadataRepositoryImpl) CountFilesByCollection(ctx context.Context, collectionID gocql.UUID) (int, error) {
	// Query files in the collection using the collection table
	query := `SELECT file_id, state FROM mapleapps.maplefile_files_by_collection_id_with_desc_modified_at_and_asc_file_id
		WHERE collection_id = ?`

	iter := impl.Session.Query(query, collectionID).WithContext(ctx).Iter()

	count := 0
	var fileID gocql.UUID
	var state string

	for iter.Scan(&fileID, &state) {
		// Only count active files
		if state != dom_file.FileStateActive {
			continue
		}

		count++
	}

	if err := iter.Close(); err != nil {
		impl.Logger.Error("failed to count files by collection",
			zap.String("collection_id", collectionID.String()),
			zap.Error(err))
		return 0, fmt.Errorf("failed to count files by collection: %w", err)
	}

	impl.Logger.Debug("counted files by collection successfully",
		zap.String("collection_id", collectionID.String()),
		zap.Int("file_count", count))

	return count, nil
}
