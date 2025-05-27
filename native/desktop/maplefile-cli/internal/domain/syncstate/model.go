// native/desktop/maplefile-cli/internal/domain/syncstate/model.go
package sync

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SyncState represents the local sync state for tracking last sync
type SyncState struct {
	LastCollectionSync time.Time          `json:"last_collection_sync"`
	LastFileSync       time.Time          `json:"last_file_sync"`
	LastCollectionID   primitive.ObjectID `json:"last_collection_id"`
	LastFileID         primitive.ObjectID `json:"last_file_id"`
}
