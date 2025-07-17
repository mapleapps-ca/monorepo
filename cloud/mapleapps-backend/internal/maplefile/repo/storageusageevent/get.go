// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/storageusageevent/get.go
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

	// For better performance with large date ranges, we'll query in parallel
	var allEvents []*storageusageevent.StorageUsageEvent
	eventsChan := make(chan []*storageusageevent.StorageUsageEvent)
	errorsChan := make(chan error)

	// Calculate number of days
	days := int(endDay.Sub(startDay).Hours()/24) + 1

	// Query each day in parallel (limit concurrency to avoid overwhelming Cassandra)
	concurrency := 10
	if days < concurrency {
		concurrency = days
	}

	semaphore := make(chan struct{}, concurrency)
	daysProcessed := 0

	for day := startDay; !day.After(endDay); day = day.Add(24 * time.Hour) {
		semaphore <- struct{}{}
		daysProcessed++

		go func(queryDay time.Time) {
			defer func() { <-semaphore }()

			events, err := impl.GetByUserAndDay(ctx, userID, queryDay)
			if err != nil {
				errorsChan <- err
				return
			}
			eventsChan <- events
		}(day)
	}

	// Collect results
	var firstError error
	for i := 0; i < daysProcessed; i++ {
		select {
		case events := <-eventsChan:
			allEvents = append(allEvents, events...)
		case err := <-errorsChan:
			if firstError == nil {
				firstError = err
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if firstError != nil {
		impl.Logger.Error("failed to get events for date range",
			zap.Error(firstError),
			zap.Int("days_requested", days))
		return allEvents, firstError // Return partial results
	}

	return allEvents, nil
}

// Convenience methods for trend analysis

func (impl *storageUsageEventRepositoryImpl) GetLast7DaysEvents(ctx context.Context, userID gocql.UUID) ([]*storageusageevent.StorageUsageEvent, error) {
	endDay := time.Now().Truncate(24 * time.Hour)
	startDay := endDay.Add(-6 * 24 * time.Hour) // 7 days including today

	return impl.GetByUserDateRange(ctx, userID, startDay, endDay)
}

func (impl *storageUsageEventRepositoryImpl) GetLastNDaysEvents(ctx context.Context, userID gocql.UUID, days int) ([]*storageusageevent.StorageUsageEvent, error) {
	if days <= 0 {
		return nil, fmt.Errorf("days must be positive")
	}

	endDay := time.Now().Truncate(24 * time.Hour)
	startDay := endDay.Add(-time.Duration(days-1) * 24 * time.Hour)

	return impl.GetByUserDateRange(ctx, userID, startDay, endDay)
}

func (impl *storageUsageEventRepositoryImpl) GetMonthlyEvents(ctx context.Context, userID gocql.UUID, year int, month time.Month) ([]*storageusageevent.StorageUsageEvent, error) {
	startDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDay := startDay.AddDate(0, 1, -1) // Last day of the month

	return impl.GetByUserDateRange(ctx, userID, startDay, endDay)
}

func (impl *storageUsageEventRepositoryImpl) GetYearlyEvents(ctx context.Context, userID gocql.UUID, year int) ([]*storageusageevent.StorageUsageEvent, error) {
	startDay := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDay := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)

	return impl.GetByUserDateRange(ctx, userID, startDay, endDay)
}
