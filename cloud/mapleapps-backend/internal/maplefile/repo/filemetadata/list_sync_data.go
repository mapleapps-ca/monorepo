// cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/list_sync_data.go
package filemetadata

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

func (impl *fileMetadataRepositoryImpl) ListSyncData(ctx context.Context, userID gocql.UUID, cursor *dom_file.FileSyncCursor, limit int64, accessibleCollectionIDs []gocql.UUID) (*dom_file.FileSyncResponse, error) {
	if len(accessibleCollectionIDs) == 0 {
		// No accessible collections, return empty response
		return &dom_file.FileSyncResponse{
			Files:   []dom_file.FileSyncItem{},
			HasMore: false,
		}, nil
	}

	// Build query based on cursor
	var query string
	var args []any

	if cursor == nil {
		// Initial sync - get all files for user
		query = `SELECT file_id, collection_id, version, modified_at, state, tombstone_version, tombstone_expiry, encrypted_file_size_in_bytes
			FROM mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
			WHERE user_id = ? LIMIT ?`
		args = []any{userID, limit}
	} else {
		// Incremental sync - get files modified after cursor
		query = `SELECT file_id, collection_id, version, modified_at, state, tombstone_version, tombstone_expiry, encrypted_file_size_in_bytes
			FROM mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
			WHERE user_id = ? AND (modified_at, file_id) > (?, ?) LIMIT ?`
		args = []any{userID, cursor.LastModified, cursor.LastID, limit}
	}

	iter := impl.Session.Query(query, args...).WithContext(ctx).Iter()

	var syncItems []dom_file.FileSyncItem
	var lastModified time.Time
	var lastID gocql.UUID

	var (
		fileID                      gocql.UUID
		collectionID                gocql.UUID
		version, tombstoneVersion   uint64
		modifiedAt, tombstoneExpiry time.Time
		state                       string
		encryptedFileSizeInBytes    int64
	)

	// Filter files by accessible collections
	accessibleCollections := make(map[gocql.UUID]bool)
	for _, cid := range accessibleCollectionIDs {
		accessibleCollections[cid] = true
	}

	for iter.Scan(&fileID, &collectionID, &version, &modifiedAt, &state, &tombstoneVersion, &tombstoneExpiry, &encryptedFileSizeInBytes) {
		// Only include files from accessible collections
		if !accessibleCollections[collectionID] {
			continue
		}

		syncItem := dom_file.FileSyncItem{
			ID:                       fileID,
			CollectionID:             collectionID,
			Version:                  version,
			ModifiedAt:               modifiedAt,
			State:                    state,
			TombstoneVersion:         tombstoneVersion,
			TombstoneExpiry:          tombstoneExpiry,
			EncryptedFileSizeInBytes: encryptedFileSizeInBytes,
		}

		syncItems = append(syncItems, syncItem)
		lastModified = modifiedAt
		lastID = fileID
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get file sync data: %w", err)
	}

	// Prepare response
	response := &dom_file.FileSyncResponse{
		Files:   syncItems,
		HasMore: len(syncItems) == int(limit),
	}

	// Set next cursor if there are more results
	if response.HasMore {
		response.NextCursor = &dom_file.FileSyncCursor{
			LastModified: lastModified,
			LastID:       lastID,
		}
	}

	impl.Logger.Debug("file sync data retrieved",
		zap.String("user_id", userID.String()),
		zap.Int("file_count", len(syncItems)),
		zap.Bool("has_more", response.HasMore))

	return response, nil
}
