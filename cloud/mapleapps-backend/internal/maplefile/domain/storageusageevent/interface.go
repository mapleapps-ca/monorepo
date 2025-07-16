// cloud/mapleapps-backend/internal/maplefile/domain/storageusageevent/interface.go
package storageusageevent

import (
	"context"
	"time"

	"github.com/gocql/gocql"
)

// StorageUsageEventRepository defines the interface for storage usage events
type StorageUsageEventRepository interface {
	Create(ctx context.Context, event *StorageUsageEvent) error
	CreateMany(ctx context.Context, events []*StorageUsageEvent) error
	GetByUserAndDay(ctx context.Context, userID gocql.UUID, eventDay time.Time) ([]*StorageUsageEvent, error)
	GetByUserDateRange(ctx context.Context, userID gocql.UUID, startDay, endDay time.Time) ([]*StorageUsageEvent, error)
	DeleteByUserAndDay(ctx context.Context, userID gocql.UUID, eventDay time.Time) error
}
