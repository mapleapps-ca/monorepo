// monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/storagedailyusage/get_trend.go
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

// GetStorageDailyUsageTrendRequest contains the trend parameters
type GetStorageDailyUsageTrendRequest struct {
	UserID      gocql.UUID  `json:"user_id"`
	TrendPeriod string      `json:"trend_period"` // "7days", "monthly", "yearly"
	Year        *int        `json:"year,omitempty"`
	Month       *time.Month `json:"month,omitempty"`
}

type GetStorageDailyUsageTrendUseCase interface {
	Execute(ctx context.Context, req *GetStorageDailyUsageTrendRequest) (*storagedailyusage.StorageUsageTrend, error)
}

type getStorageDailyUsageTrendUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   storagedailyusage.StorageDailyUsageRepository
}

func NewGetStorageDailyUsageTrendUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo storagedailyusage.StorageDailyUsageRepository,
) GetStorageDailyUsageTrendUseCase {
	logger = logger.Named("GetStorageDailyUsageTrendUseCase")
	return &getStorageDailyUsageTrendUseCaseImpl{config, logger, repo}
}

func (uc *getStorageDailyUsageTrendUseCaseImpl) Execute(ctx context.Context, req *GetStorageDailyUsageTrendRequest) (*storagedailyusage.StorageUsageTrend, error) {
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
		} else if req.TrendPeriod != "7days" && req.TrendPeriod != "monthly" && req.TrendPeriod != "yearly" {
			e["trend_period"] = "Trend period must be one of: 7days, monthly, yearly"
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
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get storage daily usage trend",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get trend based on period.
	//

	var trend *storagedailyusage.StorageUsageTrend
	var err error

	switch req.TrendPeriod {
	case "7days":
		trend, err = uc.repo.GetLast7DaysTrend(ctx, req.UserID)

	case "monthly":
		trend, err = uc.repo.GetMonthlyTrend(ctx, req.UserID, *req.Year, *req.Month)

	case "yearly":
		trend, err = uc.repo.GetYearlyTrend(ctx, req.UserID, *req.Year)

	default:
		return nil, httperror.NewForBadRequestWithSingleField("trend_period", "Invalid trend period")
	}

	if err != nil {
		uc.logger.Error("Failed to get storage daily usage trend",
			zap.String("user_id", req.UserID.String()),
			zap.String("trend_period", req.TrendPeriod),
			zap.Error(err))
		return nil, err
	}

	uc.logger.Debug("Successfully retrieved storage daily usage trend",
		zap.String("user_id", req.UserID.String()),
		zap.String("trend_period", req.TrendPeriod),
		zap.Int("daily_usages_count", len(trend.DailyUsages)),
		zap.Int64("total_added", trend.TotalAdded),
		zap.Int64("total_removed", trend.TotalRemoved),
		zap.Int64("net_change", trend.NetChange))

	return trend, nil
}
