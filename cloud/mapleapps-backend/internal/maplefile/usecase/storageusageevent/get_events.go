// monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/storageusageevent/get_events.go
package storageusageevent

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storageusageevent"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// GetStorageUsageEventsRequest contains the filtering parameters
type GetStorageUsageEventsRequest struct {
	UserID      gocql.UUID  `json:"user_id"`
	TrendPeriod string      `json:"trend_period"` // "7days", "monthly", "yearly"
	Year        *int        `json:"year,omitempty"`
	Month       *time.Month `json:"month,omitempty"`
	Days        *int        `json:"days,omitempty"` // For custom day ranges
}

// GetStorageUsageEventsResponse contains the filtered events
type GetStorageUsageEventsResponse struct {
	UserID      gocql.UUID                             `json:"user_id"`
	TrendPeriod string                                 `json:"trend_period"`
	StartDate   time.Time                              `json:"start_date"`
	EndDate     time.Time                              `json:"end_date"`
	Events      []*storageusageevent.StorageUsageEvent `json:"events"`
	EventCount  int                                    `json:"event_count"`
}

type GetStorageUsageEventsUseCase interface {
	Execute(ctx context.Context, req *GetStorageUsageEventsRequest) (*GetStorageUsageEventsResponse, error)
}

type getStorageUsageEventsUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   storageusageevent.StorageUsageEventRepository
}

func NewGetStorageUsageEventsUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo storageusageevent.StorageUsageEventRepository,
) GetStorageUsageEventsUseCase {
	logger = logger.Named("GetStorageUsageEventsUseCase")
	return &getStorageUsageEventsUseCaseImpl{config, logger, repo}
}

func (uc *getStorageUsageEventsUseCaseImpl) Execute(ctx context.Context, req *GetStorageUsageEventsRequest) (*GetStorageUsageEventsResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if req == nil {
		e["request"] = "Request is required"
	} else {
		if req.UserID.String() == "" {
			e["user_id"] = "User ID is required"
		}
		if req.TrendPeriod == "" {
			e["trend_period"] = "Trend period is required"
		} else if req.TrendPeriod != "7days" && req.TrendPeriod != "monthly" && req.TrendPeriod != "yearly" && req.TrendPeriod != "custom" {
			e["trend_period"] = "Trend period must be one of: 7days, monthly, yearly, custom"
		}

		// Validate period-specific parameters
		switch req.TrendPeriod {
		case "monthly":
			if req.Year == nil {
				e["year"] = "Year is required for monthly trend"
			}
			if req.Month == nil {
				e["month"] = "Month is required for monthly trend"
			}
		case "yearly":
			if req.Year == nil {
				e["year"] = "Year is required for yearly trend"
			}
		case "custom":
			if req.Days == nil || *req.Days <= 0 {
				e["days"] = "Days must be greater than 0 for custom trend"
			}
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get storage usage events",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get events based on trend period.
	//

	var events []*storageusageevent.StorageUsageEvent
	var err error
	var startDate, endDate time.Time

	switch req.TrendPeriod {
	case "7days":
		events, err = uc.repo.GetLast7DaysEvents(ctx, req.UserID)
		endDate = time.Now().Truncate(24 * time.Hour)
		startDate = endDate.Add(-6 * 24 * time.Hour)

	case "monthly":
		events, err = uc.repo.GetMonthlyEvents(ctx, req.UserID, *req.Year, *req.Month)
		startDate = time.Date(*req.Year, *req.Month, 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 1, -1) // Last day of the month

	case "yearly":
		events, err = uc.repo.GetYearlyEvents(ctx, req.UserID, *req.Year)
		startDate = time.Date(*req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = time.Date(*req.Year, 12, 31, 0, 0, 0, 0, time.UTC)

	case "custom":
		events, err = uc.repo.GetLastNDaysEvents(ctx, req.UserID, *req.Days)
		endDate = time.Now().Truncate(24 * time.Hour)
		startDate = endDate.Add(-time.Duration(*req.Days-1) * 24 * time.Hour)

	default:
		return nil, httperror.NewForBadRequestWithSingleField("trend_period", "Invalid trend period")
	}

	if err != nil {
		uc.logger.Error("Failed to get storage usage events",
			zap.String("user_id", req.UserID.String()),
			zap.String("trend_period", req.TrendPeriod),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 3: Build response.
	//

	response := &GetStorageUsageEventsResponse{
		UserID:      req.UserID,
		TrendPeriod: req.TrendPeriod,
		StartDate:   startDate,
		EndDate:     endDate,
		Events:      events,
		EventCount:  len(events),
	}

	uc.logger.Debug("Successfully retrieved storage usage events",
		zap.String("user_id", req.UserID.String()),
		zap.String("trend_period", req.TrendPeriod),
		zap.Int("event_count", len(events)),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate))

	return response, nil
}
