// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/storagedailyusage/update.go
package storagedailyusage

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage"
)

func (impl *storageDailyUsageRepositoryImpl) UpdateOrCreate(ctx context.Context, usage *storagedailyusage.StorageDailyUsage) error {
	if usage == nil {
		return fmt.Errorf("storage daily usage cannot be nil")
	}

	// Ensure usage day is truncated to date only
	usage.UsageDay = usage.UsageDay.Truncate(24 * time.Hour)

	// Use UPSERT (INSERT with no IF NOT EXISTS) to update or create
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
		impl.Logger.Error("failed to upsert storage daily usage", zap.Error(err))
		return fmt.Errorf("failed to upsert storage daily usage: %w", err)
	}

	return nil
}
