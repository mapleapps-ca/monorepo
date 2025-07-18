// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/storagedailyusage/create.go
package storagedailyusage

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage"
)

func (impl *storageDailyUsageRepositoryImpl) Create(ctx context.Context, usage *storagedailyusage.StorageDailyUsage) error {
	if usage == nil {
		return fmt.Errorf("storage daily usage cannot be nil")
	}

	// Ensure usage day is truncated to date only
	usage.UsageDay = usage.UsageDay.Truncate(24 * time.Hour)

	query := `INSERT INTO maplefile_storage_daily_usage_by_user_id_with_asc_usage_day
			(user_id, usage_day, total_bytes, total_add_bytes, total_remove_bytes)
			VALUES (?, ?, ?, ?, ?)`

	err := impl.Session.Query(query,
		usage.UserID,
		usage.UsageDay,
		usage.TotalBytes,
		usage.TotalAddBytes,
		usage.TotalRemoveBytes,
	).WithContext(ctx).Exec()

	if err != nil {
		impl.Logger.Error("failed to create storage daily usage",
			zap.String("user_id", usage.UserID.String()),
			zap.Time("usage_day", usage.UsageDay),
			zap.Error(err))
		return fmt.Errorf("failed to create storage daily usage: %w", err)
	}

	return nil
}

func (impl *storageDailyUsageRepositoryImpl) CreateMany(ctx context.Context, usages []*storagedailyusage.StorageDailyUsage) error {
	if len(usages) == 0 {
		return nil
	}

	batch := impl.Session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	for _, usage := range usages {
		if usage == nil {
			continue
		}

		// Ensure usage day is truncated to date only
		usage.UsageDay = usage.UsageDay.Truncate(24 * time.Hour)

		batch.Query(`INSERT INTO maplefile_storage_daily_usage_by_user_id_with_asc_usage_day
				(user_id, usage_day, total_bytes, total_add_bytes, total_remove_bytes)
				VALUES (?, ?, ?, ?, ?)`,
			usage.UserID,
			usage.UsageDay,
			usage.TotalBytes,
			usage.TotalAddBytes,
			usage.TotalRemoveBytes,
		)
	}

	err := impl.Session.ExecuteBatch(batch)
	if err != nil {
		impl.Logger.Error("failed to create multiple storage daily usages", zap.Error(err))
		return fmt.Errorf("failed to create multiple storage daily usages: %w", err)
	}

	return nil
}

func (impl *storageDailyUsageRepositoryImpl) IncrementUsage(ctx context.Context, userID gocql.UUID, usageDay time.Time, totalBytes, addBytes, removeBytes int64) error {
	// Ensure usage day is truncated to date only
	usageDay = usageDay.Truncate(24 * time.Hour)

	// First, get the current values
	existing, err := impl.GetByUserAndDay(ctx, userID, usageDay)
	if err != nil {
		impl.Logger.Error("failed to get existing usage for increment",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.Time("usage_day", usageDay))
		return fmt.Errorf("failed to get existing usage: %w", err)
	}

	// Calculate new values
	var newTotalBytes, newAddBytes, newRemoveBytes int64
	if existing != nil {
		// Add to existing values
		newTotalBytes = existing.TotalBytes + totalBytes
		newAddBytes = existing.TotalAddBytes + addBytes
		newRemoveBytes = existing.TotalRemoveBytes + removeBytes
	} else {
		// First record for this day
		newTotalBytes = totalBytes
		newAddBytes = addBytes
		newRemoveBytes = removeBytes
	}
	// Insert/Update with the new values
	query := `
			INSERT INTO maplefile_storage_daily_usage_by_user_id_with_asc_usage_day
			(user_id, usage_day, total_bytes, total_add_bytes, total_remove_bytes)
			VALUES (?, ?, ?, ?, ?)`

	if err := impl.Session.Query(query,
		userID,
		usageDay,
		newTotalBytes,
		newAddBytes,
		newRemoveBytes,
	).WithContext(ctx).Exec(); err != nil {
		impl.Logger.Error("failed to increment storage daily usage",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.Time("usage_day", usageDay))
		return fmt.Errorf("failed to increment storage daily usage: %w", err)
	}

	impl.Logger.Debug("storage daily usage incremented",
		zap.String("user_id", userID.String()),
		zap.Time("usage_day", usageDay),
		zap.Int64("total_bytes_delta", totalBytes),
		zap.Int64("add_bytes_delta", addBytes),
		zap.Int64("remove_bytes_delta", removeBytes),
		zap.Int64("new_total_bytes", newTotalBytes),
		zap.Int64("new_add_bytes", newAddBytes),
		zap.Int64("new_remove_bytes", newRemoveBytes))
	return nil
}
