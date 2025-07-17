// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/storagedailyusage/get.go
package storagedailyusage

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage"
)

func (impl *storageDailyUsageRepositoryImpl) GetByUserAndDay(ctx context.Context, userID gocql.UUID, usageDay time.Time) (*storagedailyusage.StorageDailyUsage, error) {
	// Ensure usage day is truncated to date only
	usageDay = usageDay.Truncate(24 * time.Hour)

	var (
		resultUserID     gocql.UUID
		resultUsageDay   time.Time
		totalBytes       int64
		totalAddBytes    int64
		totalRemoveBytes int64
	)

	query := `SELECT user_id, usage_day, total_bytes, total_add_bytes, total_remove_bytes
		FROM mapleapps.maplefile_storage_daily_usage_by_user_id_with_asc_usage_day
		WHERE user_id = ? AND usage_day = ?`

	err := impl.Session.Query(query, userID, usageDay).WithContext(ctx).Scan(
		&resultUserID, &resultUsageDay, &totalBytes, &totalAddBytes, &totalRemoveBytes)

	if err == gocql.ErrNotFound {
		return nil, nil
	}

	if err != nil {
		impl.Logger.Error("failed to get storage daily usage", zap.Error(err))
		return nil, fmt.Errorf("failed to get storage daily usage: %w", err)
	}

	usage := &storagedailyusage.StorageDailyUsage{
		UserID:           resultUserID,
		UsageDay:         resultUsageDay,
		TotalBytes:       totalBytes,
		TotalAddBytes:    totalAddBytes,
		TotalRemoveBytes: totalRemoveBytes,
	}

	return usage, nil
}

func (impl *storageDailyUsageRepositoryImpl) GetByUserDateRange(ctx context.Context, userID gocql.UUID, startDay, endDay time.Time) ([]*storagedailyusage.StorageDailyUsage, error) {
	// Ensure dates are truncated to date only
	startDay = startDay.Truncate(24 * time.Hour)
	endDay = endDay.Truncate(24 * time.Hour)

	query := `SELECT user_id, usage_day, total_bytes, total_add_bytes, total_remove_bytes
		FROM mapleapps.maplefile_storage_daily_usage_by_user_id_with_asc_usage_day
		WHERE user_id = ? AND usage_day >= ? AND usage_day <= ?`

	iter := impl.Session.Query(query, userID, startDay, endDay).WithContext(ctx).Iter()

	var usages []*storagedailyusage.StorageDailyUsage
	var (
		resultUserID     gocql.UUID
		resultUsageDay   time.Time
		totalBytes       int64
		totalAddBytes    int64
		totalRemoveBytes int64
	)

	for iter.Scan(&resultUserID, &resultUsageDay, &totalBytes, &totalAddBytes, &totalRemoveBytes) {
		usage := &storagedailyusage.StorageDailyUsage{
			UserID:           resultUserID,
			UsageDay:         resultUsageDay,
			TotalBytes:       totalBytes,
			TotalAddBytes:    totalAddBytes,
			TotalRemoveBytes: totalRemoveBytes,
		}
		usages = append(usages, usage)
	}

	if err := iter.Close(); err != nil {
		impl.Logger.Error("failed to get storage daily usage by date range", zap.Error(err))
		return nil, fmt.Errorf("failed to get storage daily usage: %w", err)
	}

	return usages, nil
}

// GetLast7DaysTrend retrieves the last 7 days of storage usage and calculates trends
func (impl *storageDailyUsageRepositoryImpl) GetLast7DaysTrend(ctx context.Context, userID gocql.UUID) (*storagedailyusage.StorageUsageTrend, error) {
	endDay := time.Now().Truncate(24 * time.Hour)
	startDay := endDay.Add(-6 * 24 * time.Hour) // 7 days including today

	usages, err := impl.GetByUserDateRange(ctx, userID, startDay, endDay)
	if err != nil {
		return nil, err
	}

	return impl.calculateTrend(userID, startDay, endDay, usages), nil
}

// GetMonthlyTrend retrieves usage trend for a specific month
func (impl *storageDailyUsageRepositoryImpl) GetMonthlyTrend(ctx context.Context, userID gocql.UUID, year int, month time.Month) (*storagedailyusage.StorageUsageTrend, error) {
	startDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDay := startDay.AddDate(0, 1, -1) // Last day of the month

	usages, err := impl.GetByUserDateRange(ctx, userID, startDay, endDay)
	if err != nil {
		return nil, err
	}

	return impl.calculateTrend(userID, startDay, endDay, usages), nil
}

// GetYearlyTrend retrieves usage trend for a specific year
func (impl *storageDailyUsageRepositoryImpl) GetYearlyTrend(ctx context.Context, userID gocql.UUID, year int) (*storagedailyusage.StorageUsageTrend, error) {
	startDay := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDay := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)

	usages, err := impl.GetByUserDateRange(ctx, userID, startDay, endDay)
	if err != nil {
		return nil, err
	}

	return impl.calculateTrend(userID, startDay, endDay, usages), nil
}

// GetCurrentMonthUsage gets the current month's usage summary
func (impl *storageDailyUsageRepositoryImpl) GetCurrentMonthUsage(ctx context.Context, userID gocql.UUID) (*storagedailyusage.StorageUsageSummary, error) {
	now := time.Now()
	startDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDay := now.Truncate(24 * time.Hour)

	usages, err := impl.GetByUserDateRange(ctx, userID, startDay, endDay)
	if err != nil {
		return nil, err
	}

	return impl.calculateSummary(userID, "month", startDay, endDay, usages), nil
}

// GetCurrentYearUsage gets the current year's usage summary
func (impl *storageDailyUsageRepositoryImpl) GetCurrentYearUsage(ctx context.Context, userID gocql.UUID) (*storagedailyusage.StorageUsageSummary, error) {
	now := time.Now()
	startDay := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	endDay := now.Truncate(24 * time.Hour)

	usages, err := impl.GetByUserDateRange(ctx, userID, startDay, endDay)
	if err != nil {
		return nil, err
	}

	return impl.calculateSummary(userID, "year", startDay, endDay, usages), nil
}

// Helper methods

func (impl *storageDailyUsageRepositoryImpl) calculateTrend(userID gocql.UUID, startDay, endDay time.Time, usages []*storagedailyusage.StorageDailyUsage) *storagedailyusage.StorageUsageTrend {
	trend := &storagedailyusage.StorageUsageTrend{
		UserID:      userID,
		StartDate:   startDay,
		EndDate:     endDay,
		DailyUsages: usages,
	}

	if len(usages) == 0 {
		return trend
	}

	var peakDay time.Time
	var peakBytes int64

	for _, usage := range usages {
		trend.TotalAdded += usage.TotalAddBytes
		trend.TotalRemoved += usage.TotalRemoveBytes

		if usage.TotalBytes > peakBytes {
			peakBytes = usage.TotalBytes
			peakDay = usage.UsageDay
		}
	}

	trend.NetChange = trend.TotalAdded - trend.TotalRemoved
	if len(usages) > 0 {
		trend.AverageDailyAdd = trend.TotalAdded / int64(len(usages))
		trend.PeakUsageDay = &peakDay
		trend.PeakUsageBytes = peakBytes
	}

	return trend
}

func (impl *storageDailyUsageRepositoryImpl) calculateSummary(userID gocql.UUID, period string, startDay, endDay time.Time, usages []*storagedailyusage.StorageDailyUsage) *storagedailyusage.StorageUsageSummary {
	summary := &storagedailyusage.StorageUsageSummary{
		UserID:       userID,
		Period:       period,
		StartDate:    startDay,
		EndDate:      endDay,
		DaysWithData: len(usages),
	}

	if len(usages) == 0 {
		return summary
	}

	// Get the most recent usage as current
	summary.CurrentUsage = usages[len(usages)-1].TotalBytes

	for _, usage := range usages {
		summary.TotalAdded += usage.TotalAddBytes
		summary.TotalRemoved += usage.TotalRemoveBytes
	}

	summary.NetChange = summary.TotalAdded - summary.TotalRemoved

	return summary
}
