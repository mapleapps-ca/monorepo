// cloud/mapleapps-backend/internal/maplefile/repo/storageusageevent/delete.go
package storageusageevent

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

func (impl *storageUsageEventRepositoryImpl) DeleteByUserAndDay(ctx context.Context, userID gocql.UUID, eventDay time.Time) error {
	// Ensure event day is truncated to date only
	eventDay = eventDay.Truncate(24 * time.Hour)

	query := `DELETE FROM mapleapps.maplefile_storage_usage_events_by_user_id_and_event_day_with_asc_event_time
		WHERE user_id = ? AND event_day = ?`

	err := impl.Session.Query(query, userID, eventDay).WithContext(ctx).Exec()
	if err != nil {
		impl.Logger.Error("failed to delete storage usage events by user and day", zap.Error(err))
		return fmt.Errorf("failed to delete storage usage events: %w", err)
	}

	return nil
}
