// native/desktop/maplefile-cli/internal/domain/syncdto/model.go
package syncdto

import (
	"time"

	"github.com/gocql/gocql"
)

// SyncCursorDTO represents cursor-based pagination for sync operations
type SyncCursorDTO struct {
	LastModified time.Time  `json:"last_modified"`
	LastID       gocql.UUID `json:"last_id"`
}

// CollectionSyncItem represents minimal collection data for sync operations
type CollectionSyncItem struct {
	ID               gocql.UUID  `json:"id"`
	Version          uint64      `json:"version"`
	ModifiedAt       time.Time   `json:"modified_at"`
	State            string      `json:"state"`
	ParentID         *gocql.UUID `json:"parent_id,omitempty"`
	TombstoneVersion uint64      `bson:"tombstone_version" json:"tombstone_version"`
	TombstoneExpiry  time.Time   `bson:"tombstone_expiry" json:"tombstone_expiry"`
}

// CollectionSyncResponseDTO represents the response for collection sync data
type CollectionSyncResponseDTO struct {
	Collections []CollectionSyncItem `json:"collections"`
	NextCursor  *SyncCursorDTO       `json:"next_cursor,omitempty"`
	HasMore     bool                 `json:"has_more"`
}

// FileSyncItem represents minimal file data for sync operations
type FileSyncItem struct {
	ID               gocql.UUID `json:"id"`
	CollectionID     gocql.UUID `json:"collection_id"`
	Version          uint64     `json:"version"`
	ModifiedAt       time.Time  `json:"modified_at"`
	State            string     `json:"state"`
	TombstoneVersion uint64     `bson:"tombstone_version" json:"tombstone_version"`
	TombstoneExpiry  time.Time  `bson:"tombstone_expiry" json:"tombstone_expiry"`
}

// FileSyncResponseDTO represents the response for file sync data
type FileSyncResponseDTO struct {
	Files      []FileSyncItem `json:"files"`
	NextCursor *SyncCursorDTO `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	CollectionsProcessed int      `json:"collections_processed"`
	FilesProcessed       int      `json:"files_processed"`
	CollectionsAdded     int      `json:"collections_added"`
	CollectionsUpdated   int      `json:"collections_updated"`
	CollectionsDeleted   int      `json:"collections_deleted"`
	FilesAdded           int      `json:"files_added"`
	FilesUpdated         int      `json:"files_updated"`
	FilesDeleted         int      `json:"files_deleted"`
	Errors               []string `json:"errors,omitempty"`
}
