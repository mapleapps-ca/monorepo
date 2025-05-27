// native/desktop/maplefile-cli/internal/domain/sync/model.go
package sync

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SyncCursor represents cursor-based pagination for sync operations
type SyncCursor struct {
	LastModified time.Time          `json:"last_modified"`
	LastID       primitive.ObjectID `json:"last_id"`
}

// CollectionSyncItem represents minimal collection data for sync operations
type CollectionSyncItem struct {
	ID         primitive.ObjectID  `json:"id"`
	Version    uint64              `json:"version"`
	ModifiedAt time.Time           `json:"modified_at"`
	State      string              `json:"state"`
	ParentID   *primitive.ObjectID `json:"parent_id,omitempty"`
}

// CollectionSyncResponse represents the response for collection sync data
type CollectionSyncResponse struct {
	Collections []CollectionSyncItem `json:"collections"`
	NextCursor  *SyncCursor          `json:"next_cursor,omitempty"`
	HasMore     bool                 `json:"has_more"`
}

// FileSyncItem represents minimal file data for sync operations
type FileSyncItem struct {
	ID           primitive.ObjectID `json:"id"`
	CollectionID primitive.ObjectID `json:"collection_id"`
	Version      uint64             `json:"version"`
	ModifiedAt   time.Time          `json:"modified_at"`
	State        string             `json:"state"`
}

// FileSyncResponse represents the response for file sync data
type FileSyncResponse struct {
	Files      []FileSyncItem `json:"files"`
	NextCursor *SyncCursor    `json:"next_cursor,omitempty"`
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

// SyncState represents the local sync state for tracking last sync
type SyncState struct {
	LastCollectionSync time.Time          `json:"last_collection_sync"`
	LastFileSync       time.Time          `json:"last_file_sync"`
	LastCollectionID   primitive.ObjectID `json:"last_collection_id"`
	LastFileID         primitive.ObjectID `json:"last_file_id"`
}
