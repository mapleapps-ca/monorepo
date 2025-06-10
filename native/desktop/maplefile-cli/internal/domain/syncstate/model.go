// native/desktop/maplefile-cli/internal/domain/syncstate/model.go
package syncstate

import (
	"time"

	"github.com/gocql/gocql"
)

// SyncState represents the local sync state for tracking last sync
type SyncState struct {
	LastCollectionSync time.Time  `json:"last_collection_sync"`
	LastFileSync       time.Time  `json:"last_file_sync"`
	LastCollectionID   gocql.UUID `json:"last_collection_id"`
	LastFileID         gocql.UUID `json:"last_file_id"`
}
