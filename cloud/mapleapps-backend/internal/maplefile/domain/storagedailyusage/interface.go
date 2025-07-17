// monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage/interface.go
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
	GetLast7DaysTrend(ctx context.Context, userID gocql.UUID) (*StorageUsageTrend, error)
	GetMonthlyTrend(ctx context.Context, userID gocql.UUID, year int, month time.Month) (*StorageUsageTrend, error)
	GetYearlyTrend(ctx context.Context, userID gocql.UUID, year int) (*StorageUsageTrend, error)
	GetCurrentMonthUsage(ctx context.Context, userID gocql.UUID) (*StorageUsageSummary, error)
	GetCurrentYearUsage(ctx context.Context, userID gocql.UUID) (*StorageUsageSummary, error)
}

// StorageUsageTrend represents usage trend over a period
type StorageUsageTrend struct {
	UserID          gocql.UUID           `json:"user_id"`
	StartDate       time.Time            `json:"start_date"`
	EndDate         time.Time            `json:"end_date"`
	DailyUsages     []*StorageDailyUsage `json:"daily_usages"`
	TotalAdded      int64                `json:"total_added"`
	TotalRemoved    int64                `json:"total_removed"`
	NetChange       int64                `json:"net_change"`
	AverageDailyAdd int64                `json:"average_daily_add"`
	PeakUsageDay    *time.Time           `json:"peak_usage_day,omitempty"`
	PeakUsageBytes  int64                `json:"peak_usage_bytes"`
}

// StorageUsageSummary represents a summary of storage usage
type StorageUsageSummary struct {
	UserID       gocql.UUID `json:"user_id"`
	Period       string     `json:"period"` // "month" or "year"
	StartDate    time.Time  `json:"start_date"`
	EndDate      time.Time  `json:"end_date"`
	CurrentUsage int64      `json:"current_usage_bytes"`
	TotalAdded   int64      `json:"total_added_bytes"`
	TotalRemoved int64      `json:"total_removed_bytes"`
	NetChange    int64      `json:"net_change_bytes"`
	DaysWithData int        `json:"days_with_data"`
}
