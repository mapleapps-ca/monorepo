// cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage/interface.go
package storagedailyusage

import (
	"context"
	"time"

	"github.com/gocql/gocql"
)

// StorageDailyUsageRepository defines the interface for daily storage usage aggregates
type StorageDailyUsageRepository interface {
	Create(ctx context.Context, usage *StorageDailyUsage) error
	CreateMany(ctx context.Context, usages []*StorageDailyUsage) error
	GetByUserAndDay(ctx context.Context, userID gocql.UUID, usageDay time.Time) (*StorageDailyUsage, error)
	GetByUserDateRange(ctx context.Context, userID gocql.UUID, startDay, endDay time.Time) ([]*StorageDailyUsage, error)
	UpdateOrCreate(ctx context.Context, usage *StorageDailyUsage) error
	IncrementUsage(ctx context.Context, userID gocql.UUID, usageDay time.Time, totalBytes, addBytes, removeBytes int64) error
	DeleteByUserAndDay(ctx context.Context, userID gocql.UUID, usageDay time.Time) error
}
