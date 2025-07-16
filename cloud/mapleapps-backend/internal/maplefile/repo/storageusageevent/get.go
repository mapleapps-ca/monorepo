// cloud/mapleapps-backend/internal/maplefile/repo/storageusageevent/get.go
package storageusageevent

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storageusageevent"
)

func (impl *storageUsageEventRepositoryImpl) GetByUserAndDay(ctx context.Context, userID gocql.UUID, eventDay time.Time) ([]*storageusageevent.StorageUsageEvent, error) {
	// Ensure event day is truncated to date only
	eventDay = eventDay.Truncate(24 * time.Hour)

	query := `SELECT user_id, event_day, event_time, file_size, operation
		FROM mapleapps.maplefile_storage_usage_events_by_user_id_and_event_day_with_asc_event_time
		WHERE user_id = ? AND event_day = ?`

	iter := impl.Session.Query(query, userID, eventDay).WithContext(ctx).Iter()

	var events []*storageusageevent.StorageUsageEvent
	var (
		resultUserID   gocql.UUID
		resultEventDay time.Time
		eventTime      time.Time
		fileSize       int64
		operation      string
	)

	for iter.Scan(&resultUserID, &resultEventDay, &eventTime, &fileSize, &operation) {
		event := &storageusageevent.StorageUsageEvent{
			UserID:    resultUserID,
			EventDay:  resultEventDay,
			EventTime: eventTime,
			FileSize:  fileSize,
			Operation: operation,
		}
		events = append(events, event)
	}

	if err := iter.Close(); err != nil {
		impl.Logger.Error("failed to get storage usage events by user and day", zap.Error(err))
		return nil, fmt.Errorf("failed to get storage usage events: %w", err)
	}

	return events, nil
}

func (impl *storageUsageEventRepositoryImpl) GetByUserDateRange(ctx context.Context, userID gocql.UUID, startDay, endDay time.Time) ([]*storageusageevent.StorageUsageEvent, error) {
	// Ensure dates are truncated to date only
	startDay = startDay.Truncate(24 * time.Hour)
	endDay = endDay.Truncate(24 * time.Hour)

	var allEvents []*storageusageevent.StorageUsageEvent

	// Iterate through each day in the range
	for day := startDay; !day.After(endDay); day = day.Add(24 * time.Hour) {
		events, err := impl.GetByUserAndDay(ctx, userID, day)
		if err != nil {
			impl.Logger.Warn("failed to get events for day, continuing", zap.Error(err))
			continue
		}
		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}
