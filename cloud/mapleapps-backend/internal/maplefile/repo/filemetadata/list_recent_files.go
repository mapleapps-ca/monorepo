// cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/list_recent_files.go
package filemetadata

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

// Using types from dom_file package (defined in model.go)

// ListRecentFiles retrieves recent files with pagination for the specified user and accessible collections
func (impl *fileMetadataRepositoryImpl) ListRecentFiles(ctx context.Context, userID gocql.UUID, cursor *dom_file.RecentFilesCursor, limit int64, accessibleCollectionIDs []gocql.UUID) (*dom_file.RecentFilesResponse, error) {
	if len(accessibleCollectionIDs) == 0 {
		// No accessible collections, return empty response
		return &dom_file.RecentFilesResponse{
			Files:   []dom_file.RecentFilesItem{},
			HasMore: false,
		}, nil
	}

	// Build query based on cursor
	var query string
	var args []any

	if cursor == nil {
		// Initial request - get most recent files for user
		query = `SELECT file_id, collection_id, owner_id, encrypted_metadata, encrypted_file_key,
			encryption_version, encrypted_hash, encrypted_file_size_in_bytes, encrypted_thumbnail_size_in_bytes,
			created_at, modified_at, version, state
			FROM mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
			WHERE user_id = ? LIMIT ?`
		args = []any{userID, limit}
	} else {
		// Paginated request - get files modified before cursor
		query = `SELECT file_id, collection_id, owner_id, encrypted_metadata, encrypted_file_key,
			encryption_version, encrypted_hash, encrypted_file_size_in_bytes, encrypted_thumbnail_size_in_bytes,
			created_at, modified_at, version, state
			FROM mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
			WHERE user_id = ? AND (modified_at, file_id) < (?, ?) LIMIT ?`
		args = []any{userID, cursor.LastModified, cursor.LastID, limit}
	}

	iter := impl.Session.Query(query, args...).WithContext(ctx).Iter()

	var recentItems []dom_file.RecentFilesItem
	var lastModified time.Time
	var lastID gocql.UUID

	var (
		fileID                                                                gocql.UUID
		collectionID, ownerID                                                 gocql.UUID
		encryptedMetadata, encryptedFileKey, encryptionVersion, encryptedHash string
		encryptedFileSizeInBytes, encryptedThumbnailSizeInBytes               int64
		createdAt, modifiedAt                                                 time.Time
		version                                                               uint64
		state                                                                 string
	)

	// Filter files by accessible collections and only include active files
	accessibleCollections := make(map[gocql.UUID]bool)
	for _, cid := range accessibleCollectionIDs {
		accessibleCollections[cid] = true
	}

	for iter.Scan(&fileID, &collectionID, &ownerID, &encryptedMetadata, &encryptedFileKey,
		&encryptionVersion, &encryptedHash, &encryptedFileSizeInBytes, &encryptedThumbnailSizeInBytes,
		&createdAt, &modifiedAt, &version, &state) {

		// Only include files from accessible collections
		if !accessibleCollections[collectionID] {
			continue
		}

		// Only include active files (exclude deleted, archived, pending)
		if state != dom_file.FileStateActive {
			continue
		}

		recentItem := dom_file.RecentFilesItem{
			ID:                            fileID,
			CollectionID:                  collectionID,
			OwnerID:                       ownerID,
			EncryptedMetadata:             encryptedMetadata,
			EncryptedFileKey:              encryptedFileKey,
			EncryptionVersion:             encryptionVersion,
			EncryptedHash:                 encryptedHash,
			EncryptedFileSizeInBytes:      encryptedFileSizeInBytes,
			EncryptedThumbnailSizeInBytes: encryptedThumbnailSizeInBytes,
			CreatedAt:                     createdAt,
			ModifiedAt:                    modifiedAt,
			Version:                       version,
			State:                         state,
		}

		recentItems = append(recentItems, recentItem)
		lastModified = modifiedAt
		lastID = fileID
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get recent files: %w", err)
	}

	// Prepare response
	response := &dom_file.RecentFilesResponse{
		Files:   recentItems,
		HasMore: len(recentItems) == int(limit),
	}

	// Set next cursor if there are more results
	if response.HasMore {
		response.NextCursor = &dom_file.RecentFilesCursor{
			LastModified: lastModified,
			LastID:       lastID,
		}
	}

	impl.Logger.Debug("recent files retrieved",
		zap.String("user_id", userID.String()),
		zap.Int("file_count", len(recentItems)),
		zap.Bool("has_more", response.HasMore))

	return response, nil
}
