// cloud/mapleapps-backend/internal/maplefile/repo/storageusageevent/create.go
package storageusageevent

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storageusageevent"
)

func (impl *storageUsageEventRepositoryImpl) Create(ctx context.Context, event *storageusageevent.StorageUsageEvent) error {
	if event == nil {
		return fmt.Errorf("storage usage event cannot be nil")
	}

	// Ensure event day is truncated to date only
	event.EventDay = event.EventDay.Truncate(24 * time.Hour)

	// Set event time if not provided
	if event.EventTime.IsZero() {
		event.EventTime = time.Now()
	}

	query := `INSERT INTO mapleapps.maplefile_storage_usage_events_by_user_id_and_event_day_with_asc_event_time
		(user_id, event_day, event_time, file_size, operation)
		VALUES (?, ?, ?, ?, ?)`

	err := impl.Session.Query(query,
		event.UserID,
		event.EventDay,
		event.EventTime,
		event.FileSize,
		event.Operation).WithContext(ctx).Exec()

	if err != nil {
		impl.Logger.Error("failed to create storage usage event",
			zap.String("user_id", event.UserID.String()),
			zap.String("operation", event.Operation),
			zap.Int64("file_size", event.FileSize),
			zap.Error(err))
		return fmt.Errorf("failed to create storage usage event: %w", err)
	}

	return nil
}

func (impl *storageUsageEventRepositoryImpl) CreateMany(ctx context.Context, events []*storageusageevent.StorageUsageEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch := impl.Session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	for _, event := range events {
		if event == nil {
			continue
		}

		// Ensure event day is truncated to date only
		event.EventDay = event.EventDay.Truncate(24 * time.Hour)

		// Set event time if not provided
		if event.EventTime.IsZero() {
			event.EventTime = time.Now()
		}

		batch.Query(`INSERT INTO mapleapps.maplefile_storage_usage_events_by_user_id_and_event_day_with_asc_event_time
			(user_id, event_day, event_time, file_size, operation)
			VALUES (?, ?, ?, ?, ?)`,
			event.UserID,
			event.EventDay,
			event.EventTime,
			event.FileSize,
			event.Operation)
	}

	err := impl.Session.ExecuteBatch(batch)
	if err != nil {
		impl.Logger.Error("failed to create multiple storage usage events", zap.Error(err))
		return fmt.Errorf("failed to create multiple storage usage events: %w", err)
	}

	return nil
}
