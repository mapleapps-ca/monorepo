// cloud/mapleapps-backend/internal/maplefile/repo/storagedailyusage/get.go
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
