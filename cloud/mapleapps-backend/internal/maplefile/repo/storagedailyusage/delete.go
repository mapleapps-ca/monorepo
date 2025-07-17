// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/storagedailyusage/delete.go
package storagedailyusage

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

func (impl *storageDailyUsageRepositoryImpl) DeleteByUserAndDay(ctx context.Context, userID gocql.UUID, usageDay time.Time) error {
	// Ensure usage day is truncated to date only
	usageDay = usageDay.Truncate(24 * time.Hour)

	query := `DELETE FROM mapleapps.maplefile_storage_daily_usage_by_user_id_with_asc_usage_day
		WHERE user_id = ? AND usage_day = ?`

	err := impl.Session.Query(query, userID, usageDay).WithContext(ctx).Exec()
	if err != nil {
		impl.Logger.Error("failed to delete storage daily usage", zap.Error(err))
		return fmt.Errorf("failed to delete storage daily usage: %w", err)
	}

	return nil
}
