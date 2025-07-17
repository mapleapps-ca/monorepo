// monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/storagedailyusage/get_usage_by_date_range.go
package storagedailyusage

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// GetStorageUsageByDateRangeRequest contains the date range parameters
type GetStorageUsageByDateRangeRequest struct {
	UserID    gocql.UUID `json:"user_id"`
	StartDate time.Time  `json:"start_date"`
	EndDate   time.Time  `json:"end_date"`
}

// GetStorageUsageByDateRangeResponse contains the usage data for the date range
type GetStorageUsageByDateRangeResponse struct {
	UserID      gocql.UUID                             `json:"user_id"`
	StartDate   time.Time                              `json:"start_date"`
	EndDate     time.Time                              `json:"end_date"`
	DailyUsages []*storagedailyusage.StorageDailyUsage `json:"daily_usages"`
	Summary     *DateRangeSummary                      `json:"summary"`
}

// DateRangeSummary contains aggregated statistics for the date range
type DateRangeSummary struct {
	TotalDays        int        `json:"total_days"`
	DaysWithData     int        `json:"days_with_data"`
	TotalAdded       int64      `json:"total_added"`
	TotalRemoved     int64      `json:"total_removed"`
	NetChange        int64      `json:"net_change"`
	AverageDailyAdd  float64    `json:"average_daily_add"`
	PeakUsageDay     *time.Time `json:"peak_usage_day,omitempty"`
	PeakUsageBytes   int64      `json:"peak_usage_bytes"`
	LowestUsageDay   *time.Time `json:"lowest_usage_day,omitempty"`
	LowestUsageBytes int64      `json:"lowest_usage_bytes"`
}

type GetStorageUsageByDateRangeUseCase interface {
	Execute(ctx context.Context, req *GetStorageUsageByDateRangeRequest) (*GetStorageUsageByDateRangeResponse, error)
}

type getStorageUsageByDateRangeUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   storagedailyusage.StorageDailyUsageRepository
}

func NewGetStorageUsageByDateRangeUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo storagedailyusage.StorageDailyUsageRepository,
) GetStorageUsageByDateRangeUseCase {
	logger = logger.Named("GetStorageUsageByDateRangeUseCase")
	return &getStorageUsageByDateRangeUseCaseImpl{config, logger, repo}
}

func (uc *getStorageUsageByDateRangeUseCaseImpl) Execute(ctx context.Context, req *GetStorageUsageByDateRangeRequest) (*GetStorageUsageByDateRangeResponse, error) {
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
		if req.StartDate.IsZero() {
			e["start_date"] = "Start date is required"
		}
		if req.EndDate.IsZero() {
			e["end_date"] = "End date is required"
		}
		if !req.StartDate.IsZero() && !req.EndDate.IsZero() && req.StartDate.After(req.EndDate) {
			e["date_range"] = "Start date must be before or equal to end date"
		}
		// Check for reasonable date range (max 1 year)
		if !req.StartDate.IsZero() && !req.EndDate.IsZero() {
			daysDiff := int(req.EndDate.Sub(req.StartDate).Hours() / 24)
			if daysDiff > 365 {
				e["date_range"] = "Date range cannot exceed 365 days"
			}
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get storage usage by date range",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get usage data from repository.
	//

	// Truncate dates to ensure we're working with date-only values
	startDate := req.StartDate.Truncate(24 * time.Hour)
	endDate := req.EndDate.Truncate(24 * time.Hour)

	dailyUsages, err := uc.repo.GetByUserDateRange(ctx, req.UserID, startDate, endDate)
	if err != nil {
		uc.logger.Error("Failed to get storage usage by date range",
			zap.String("user_id", req.UserID.String()),
			zap.Time("start_date", startDate),
			zap.Time("end_date", endDate),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 3: Generate summary statistics.
	//

	summary := uc.generateDateRangeSummary(startDate, endDate, dailyUsages)

	response := &GetStorageUsageByDateRangeResponse{
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
		DailyUsages: dailyUsages,
		Summary:     summary,
	}

	uc.logger.Debug("Successfully retrieved storage usage by date range",
		zap.String("user_id", req.UserID.String()),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
		zap.Int("daily_usages_count", len(dailyUsages)),
		zap.Int("days_with_data", summary.DaysWithData),
		zap.Int64("net_change", summary.NetChange))

	return response, nil
}

// generateDateRangeSummary creates summary statistics for the date range
func (uc *getStorageUsageByDateRangeUseCaseImpl) generateDateRangeSummary(startDate, endDate time.Time, dailyUsages []*storagedailyusage.StorageDailyUsage) *DateRangeSummary {
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1

	summary := &DateRangeSummary{
		TotalDays:        totalDays,
		DaysWithData:     len(dailyUsages),
		LowestUsageBytes: int64(^uint64(0) >> 1), // Max int64 value as initial
	}

	if len(dailyUsages) == 0 {
		summary.LowestUsageBytes = 0
		return summary
	}

	for _, usage := range dailyUsages {
		summary.TotalAdded += usage.TotalAddBytes
		summary.TotalRemoved += usage.TotalRemoveBytes

		// Track peak usage
		if usage.TotalBytes > summary.PeakUsageBytes {
			summary.PeakUsageBytes = usage.TotalBytes
			peakDay := usage.UsageDay
			summary.PeakUsageDay = &peakDay
		}

		// Track lowest usage
		if usage.TotalBytes < summary.LowestUsageBytes {
			summary.LowestUsageBytes = usage.TotalBytes
			lowestDay := usage.UsageDay
			summary.LowestUsageDay = &lowestDay
		}
	}

	summary.NetChange = summary.TotalAdded - summary.TotalRemoved

	// Calculate average daily add (only for days with data)
	if summary.DaysWithData > 0 {
		summary.AverageDailyAdd = float64(summary.TotalAdded) / float64(summary.DaysWithData)
	}

	return summary
}
