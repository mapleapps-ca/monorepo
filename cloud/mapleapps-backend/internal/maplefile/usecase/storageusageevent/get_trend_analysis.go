// monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/storageusageevent/get_trend_analysis.go
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

// StorageEventTrendAnalysis contains aggregated trend data
type StorageEventTrendAnalysis struct {
	UserID                gocql.UUID   `json:"user_id"`
	TrendPeriod           string       `json:"trend_period"`
	StartDate             time.Time    `json:"start_date"`
	EndDate               time.Time    `json:"end_date"`
	TotalEvents           int          `json:"total_events"`
	AddEvents             int          `json:"add_events"`
	RemoveEvents          int          `json:"remove_events"`
	TotalBytesAdded       int64        `json:"total_bytes_added"`
	TotalBytesRemoved     int64        `json:"total_bytes_removed"`
	NetBytesChange        int64        `json:"net_bytes_change"`
	AverageBytesPerAdd    float64      `json:"average_bytes_per_add"`
	AverageBytesPerRemove float64      `json:"average_bytes_per_remove"`
	LargestAddEvent       int64        `json:"largest_add_event"`
	LargestRemoveEvent    int64        `json:"largest_remove_event"`
	DailyBreakdown        []DailyStats `json:"daily_breakdown,omitempty"`
}

// DailyStats represents daily aggregated statistics
type DailyStats struct {
	Date         time.Time `json:"date"`
	AddEvents    int       `json:"add_events"`
	RemoveEvents int       `json:"remove_events"`
	BytesAdded   int64     `json:"bytes_added"`
	BytesRemoved int64     `json:"bytes_removed"`
	NetChange    int64     `json:"net_change"`
}

type GetStorageUsageEventsTrendAnalysisUseCase interface {
	Execute(ctx context.Context, req *GetStorageUsageEventsRequest) (*StorageEventTrendAnalysis, error)
}

type getStorageUsageEventsTrendAnalysisUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   storageusageevent.StorageUsageEventRepository
}

func NewGetStorageUsageEventsTrendAnalysisUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo storageusageevent.StorageUsageEventRepository,
) GetStorageUsageEventsTrendAnalysisUseCase {
	logger = logger.Named("GetStorageUsageEventsTrendAnalysisUseCase")
	return &getStorageUsageEventsTrendAnalysisUseCaseImpl{config, logger, repo}
}

func (uc *getStorageUsageEventsTrendAnalysisUseCaseImpl) Execute(ctx context.Context, req *GetStorageUsageEventsRequest) (*StorageEventTrendAnalysis, error) {
	//
	// STEP 1: Validation (reuse from GetStorageUsageEventsUseCase).
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
		uc.logger.Warn("Failed validating get storage usage events trend analysis",
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
		endDate = startDate.AddDate(0, 1, -1)

	case "yearly":
		events, err = uc.repo.GetYearlyEvents(ctx, req.UserID, *req.Year)
		startDate = time.Date(*req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = time.Date(*req.Year, 12, 31, 0, 0, 0, 0, time.UTC)

	case "custom":
		events, err = uc.repo.GetLastNDaysEvents(ctx, req.UserID, *req.Days)
		endDate = time.Now().Truncate(24 * time.Hour)
		startDate = endDate.Add(-time.Duration(*req.Days-1) * 24 * time.Hour)
	}

	if err != nil {
		uc.logger.Error("Failed to get storage usage events for trend analysis",
			zap.String("user_id", req.UserID.String()),
			zap.String("trend_period", req.TrendPeriod),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 3: Analyze events and build trend analysis.
	//

	analysis := uc.analyzeEvents(req.UserID, req.TrendPeriod, startDate, endDate, events)

	uc.logger.Debug("Successfully analyzed storage usage events trend",
		zap.String("user_id", req.UserID.String()),
		zap.String("trend_period", req.TrendPeriod),
		zap.Int("total_events", analysis.TotalEvents),
		zap.Int64("net_bytes_change", analysis.NetBytesChange))

	return analysis, nil
}

// analyzeEvents processes the events and generates trend analysis
func (uc *getStorageUsageEventsTrendAnalysisUseCaseImpl) analyzeEvents(userID gocql.UUID, trendPeriod string, startDate, endDate time.Time, events []*storageusageevent.StorageUsageEvent) *StorageEventTrendAnalysis {
	analysis := &StorageEventTrendAnalysis{
		UserID:      userID,
		TrendPeriod: trendPeriod,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if len(events) == 0 {
		return analysis
	}

	// Daily breakdown map
	dailyMap := make(map[string]*DailyStats)

	// Process each event
	for _, event := range events {
		analysis.TotalEvents++

		if event.Operation == "add" {
			analysis.AddEvents++
			analysis.TotalBytesAdded += event.FileSize
			if event.FileSize > analysis.LargestAddEvent {
				analysis.LargestAddEvent = event.FileSize
			}
		} else if event.Operation == "remove" {
			analysis.RemoveEvents++
			analysis.TotalBytesRemoved += event.FileSize
			if event.FileSize > analysis.LargestRemoveEvent {
				analysis.LargestRemoveEvent = event.FileSize
			}
		}

		// Daily breakdown
		dayKey := event.EventDay.Format("2006-01-02")
		if dailyMap[dayKey] == nil {
			dailyMap[dayKey] = &DailyStats{
				Date: event.EventDay,
			}
		}

		daily := dailyMap[dayKey]
		if event.Operation == "add" {
			daily.AddEvents++
			daily.BytesAdded += event.FileSize
		} else if event.Operation == "remove" {
			daily.RemoveEvents++
			daily.BytesRemoved += event.FileSize
		}
		daily.NetChange = daily.BytesAdded - daily.BytesRemoved
	}

	// Calculate derived metrics
	analysis.NetBytesChange = analysis.TotalBytesAdded - analysis.TotalBytesRemoved

	if analysis.AddEvents > 0 {
		analysis.AverageBytesPerAdd = float64(analysis.TotalBytesAdded) / float64(analysis.AddEvents)
	}

	if analysis.RemoveEvents > 0 {
		analysis.AverageBytesPerRemove = float64(analysis.TotalBytesRemoved) / float64(analysis.RemoveEvents)
	}

	// Convert daily map to slice and sort by date
	for _, daily := range dailyMap {
		analysis.DailyBreakdown = append(analysis.DailyBreakdown, *daily)
	}

	// Sort daily breakdown by date
	for i := 0; i < len(analysis.DailyBreakdown)-1; i++ {
		for j := i + 1; j < len(analysis.DailyBreakdown); j++ {
			if analysis.DailyBreakdown[i].Date.After(analysis.DailyBreakdown[j].Date) {
				analysis.DailyBreakdown[i], analysis.DailyBreakdown[j] = analysis.DailyBreakdown[j], analysis.DailyBreakdown[i]
			}
		}
	}

	return analysis
}
