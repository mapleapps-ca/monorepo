// monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storageusageevent/model.go
package storageusageevent

import (
	"time"

	"github.com/gocql/gocql"
)

type StorageUsageEvent struct {
	UserID    gocql.UUID `json:"user_id"`    // Partition key
	EventDay  time.Time  `json:"event_day"`  // Partition key (date only)
	EventTime time.Time  `json:"event_time"` // Clustering key
	FileSize  int64      `json:"file_size"`  // Bytes
	Operation string     `json:"operation"`  // "add" or "remove"
}
