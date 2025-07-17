// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/storage_size.go
package filemetadata

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

// GetTotalStorageSizeByOwner calculates total storage size for all active files owned by the user
func (impl *fileMetadataRepositoryImpl) GetTotalStorageSizeByOwner(ctx context.Context, ownerID gocql.UUID) (int64, error) {
	// Query files owned by the user using the owner table
	query := `SELECT file_id, state, encrypted_file_size_in_bytes, encrypted_thumbnail_size_in_bytes
		FROM mapleapps.maplefile_files_by_owner_id_with_desc_modified_at_and_asc_file_id
		WHERE owner_id = ?`

	iter := impl.Session.Query(query, ownerID).WithContext(ctx).Iter()

	var totalSize int64
	var fileID gocql.UUID
	var state string
	var fileSize, thumbnailSize int64

	for iter.Scan(&fileID, &state, &fileSize, &thumbnailSize) {
		// Only include active files in size calculation
		if state != dom_file.FileStateActive {
			continue
		}

		// Add both file and thumbnail sizes
		totalSize += fileSize + thumbnailSize
	}

	if err := iter.Close(); err != nil {
		impl.Logger.Error("failed to calculate total storage size by owner",
			zap.String("owner_id", ownerID.String()),
			zap.Error(err))
		return 0, fmt.Errorf("failed to calculate total storage size by owner: %w", err)
	}

	impl.Logger.Debug("calculated total storage size by owner successfully",
		zap.String("owner_id", ownerID.String()),
		zap.Int64("total_size_bytes", totalSize))

	return totalSize, nil
}

// GetTotalStorageSizeByUser calculates total storage size for all active files accessible to the user
// accessibleCollectionIDs should include all collections the user owns or has access to
func (impl *fileMetadataRepositoryImpl) GetTotalStorageSizeByUser(ctx context.Context, userID gocql.UUID, accessibleCollectionIDs []gocql.UUID) (int64, error) {
	if len(accessibleCollectionIDs) == 0 {
		// No accessible collections, return 0
		impl.Logger.Debug("no accessible collections provided for storage size calculation",
			zap.String("user_id", userID.String()))
		return 0, nil
	}

	// Create a map for efficient collection access checking
	accessibleCollections := make(map[gocql.UUID]bool)
	for _, cid := range accessibleCollectionIDs {
		accessibleCollections[cid] = true
	}

	// Query files for the user using the user sync table
	query := `SELECT file_id, collection_id, state, encrypted_file_size_in_bytes, encrypted_thumbnail_size_in_bytes
		FROM mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
		WHERE user_id = ?`

	iter := impl.Session.Query(query, userID).WithContext(ctx).Iter()

	var totalSize int64
	var fileID, collectionID gocql.UUID
	var state string
	var fileSize, thumbnailSize int64

	for iter.Scan(&fileID, &collectionID, &state, &fileSize, &thumbnailSize) {
		// Only include files from accessible collections
		if !accessibleCollections[collectionID] {
			continue
		}

		// Only include active files in size calculation
		if state != dom_file.FileStateActive {
			continue
		}

		// Add both file and thumbnail sizes
		totalSize += fileSize + thumbnailSize
	}

	if err := iter.Close(); err != nil {
		impl.Logger.Error("failed to calculate total storage size by user",
			zap.String("user_id", userID.String()),
			zap.Int("accessible_collections_count", len(accessibleCollectionIDs)),
			zap.Error(err))
		return 0, fmt.Errorf("failed to calculate total storage size by user: %w", err)
	}

	impl.Logger.Debug("calculated total storage size by user successfully",
		zap.String("user_id", userID.String()),
		zap.Int("accessible_collections_count", len(accessibleCollectionIDs)),
		zap.Int64("total_size_bytes", totalSize))

	return totalSize, nil
}

// GetTotalStorageSizeByCollection calculates total storage size for all active files in a specific collection
func (impl *fileMetadataRepositoryImpl) GetTotalStorageSizeByCollection(ctx context.Context, collectionID gocql.UUID) (int64, error) {
	// Query files in the collection using the collection table
	query := `SELECT file_id, state, encrypted_file_size_in_bytes, encrypted_thumbnail_size_in_bytes
		FROM mapleapps.maplefile_files_by_collection_id_with_desc_modified_at_and_asc_file_id
		WHERE collection_id = ?`

	iter := impl.Session.Query(query, collectionID).WithContext(ctx).Iter()

	var totalSize int64
	var fileID gocql.UUID
	var state string
	var fileSize, thumbnailSize int64

	for iter.Scan(&fileID, &state, &fileSize, &thumbnailSize) {
		// Only include active files in size calculation
		if state != dom_file.FileStateActive {
			continue
		}

		// Add both file and thumbnail sizes
		totalSize += fileSize + thumbnailSize
	}

	if err := iter.Close(); err != nil {
		impl.Logger.Error("failed to calculate total storage size by collection",
			zap.String("collection_id", collectionID.String()),
			zap.Error(err))
		return 0, fmt.Errorf("failed to calculate total storage size by collection: %w", err)
	}

	impl.Logger.Debug("calculated total storage size by collection successfully",
		zap.String("collection_id", collectionID.String()),
		zap.Int64("total_size_bytes", totalSize))

	return totalSize, nil
}

// GetStorageSizeBreakdownByUser provides detailed breakdown of storage usage
// Returns owned size, shared size, and detailed collection breakdown
func (impl *fileMetadataRepositoryImpl) GetStorageSizeBreakdownByUser(ctx context.Context, userID gocql.UUID, ownedCollectionIDs, sharedCollectionIDs []gocql.UUID) (ownedSize, sharedSize int64, collectionBreakdown map[gocql.UUID]int64, err error) {
	collectionBreakdown = make(map[gocql.UUID]int64)

	// Calculate owned files storage size
	if len(ownedCollectionIDs) > 0 {
		ownedSize, err = impl.GetTotalStorageSizeByUser(ctx, userID, ownedCollectionIDs)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("failed to calculate owned storage size: %w", err)
		}

		// Get breakdown by owned collections
		for _, collectionID := range ownedCollectionIDs {
			size, sizeErr := impl.GetTotalStorageSizeByCollection(ctx, collectionID)
			if sizeErr != nil {
				impl.Logger.Warn("failed to get storage size for owned collection",
					zap.String("collection_id", collectionID.String()),
					zap.Error(sizeErr))
				continue
			}
			collectionBreakdown[collectionID] = size
		}
	}

	// Calculate shared files storage size
	if len(sharedCollectionIDs) > 0 {
		sharedSize, err = impl.GetTotalStorageSizeByUser(ctx, userID, sharedCollectionIDs)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("failed to calculate shared storage size: %w", err)
		}

		// Get breakdown by shared collections
		for _, collectionID := range sharedCollectionIDs {
			size, sizeErr := impl.GetTotalStorageSizeByCollection(ctx, collectionID)
			if sizeErr != nil {
				impl.Logger.Warn("failed to get storage size for shared collection",
					zap.String("collection_id", collectionID.String()),
					zap.Error(sizeErr))
				continue
			}
			// Note: For shared collections, this shows the total size of the collection,
			// not just the user's contribution to it
			collectionBreakdown[collectionID] = size
		}
	}

	impl.Logger.Debug("calculated storage size breakdown successfully",
		zap.String("user_id", userID.String()),
		zap.Int64("owned_size_bytes", ownedSize),
		zap.Int64("shared_size_bytes", sharedSize),
		zap.Int("owned_collections_count", len(ownedCollectionIDs)),
		zap.Int("shared_collections_count", len(sharedCollectionIDs)))

	return ownedSize, sharedSize, collectionBreakdown, nil
}
