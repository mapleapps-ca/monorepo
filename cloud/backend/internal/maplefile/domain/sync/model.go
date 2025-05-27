// cloud/backend/internal/maplefile/domain/sync/model.go
package sync

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SyncCursor represents cursor-based pagination for sync operations
type SyncCursor struct {
	LastModified time.Time          `json:"last_modified" bson:"last_modified"`
	LastID       primitive.ObjectID `json:"last_id" bson:"last_id"`
}

// CollectionSyncItem represents minimal collection data for sync operations
type CollectionSyncItem struct {
	ID         primitive.ObjectID  `json:"id" bson:"_id"`
	Version    uint64              `json:"version" bson:"version"`
	ModifiedAt time.Time           `json:"modified_at" bson:"modified_at"`
	State      string              `json:"state" bson:"state"`
	ParentID   *primitive.ObjectID `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
}

// CollectionSyncResponse represents the response for collection sync data
type CollectionSyncResponse struct {
	Collections []CollectionSyncItem `json:"collections"`
	NextCursor  *SyncCursor          `json:"next_cursor,omitempty"`
	HasMore     bool                 `json:"has_more"`
}

// FileSyncItem represents minimal file data for sync operations
type FileSyncItem struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	CollectionID primitive.ObjectID `json:"collection_id" bson:"collection_id"`
	Version      uint64             `json:"version" bson:"version"`
	ModifiedAt   time.Time          `json:"modified_at" bson:"modified_at"`
	State        string             `json:"state" bson:"state"`
}

// FileSyncResponse represents the response for file sync data
type FileSyncResponse struct {
	Files      []FileSyncItem `json:"files"`
	NextCursor *SyncCursor    `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}
