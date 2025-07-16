// cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage/model.go
package storagedailyusage

import (
	"time"

	"github.com/gocql/gocql"
)

type StorageDailyUsage struct {
	UserID           gocql.UUID `json:"user_id"`   // Partition key
	UsageDay         time.Time  `json:"usage_day"` // Clustering key (date only)
	TotalBytes       int64      `json:"total_bytes"`
	TotalAddBytes    int64      `json:"total_add_bytes"`
	TotalRemoveBytes int64      `json:"total_remove_bytes"`
}

//
// Use gocql.UUID from the github.com/gocql/gocql driver.
//
// For consistency, always store and retrieve DATE fields (like event_day and usage_day) as time.Time, but truncate to date only before inserting:
//
// ```go
// usageDay := time.Now().Truncate(24 * time.Hour)
// ```
//
