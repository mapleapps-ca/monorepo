// cloud/mapleapps-backend/internal/maplefile/usecase/storagedailyusage/get_usage_summary.go
package storagedailyusage

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// GetStorageUsageSummaryRequest contains the summary parameters
type GetStorageUsageSummaryRequest struct {
	UserID      gocql.UUID `json:"user_id"`
	SummaryType string     `json:"summary_type"` // "current_month", "current_year"
}

type GetStorageUsageSummaryUseCase interface {
	Execute(ctx context.Context, req *GetStorageUsageSummaryRequest) (*storagedailyusage.StorageUsageSummary, error)
}

type getStorageUsageSummaryUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   storagedailyusage.StorageDailyUsageRepository
}

func NewGetStorageUsageSummaryUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo storagedailyusage.StorageDailyUsageRepository,
) GetStorageUsageSummaryUseCase {
	logger = logger.Named("GetStorageUsageSummaryUseCase")
	return &getStorageUsageSummaryUseCaseImpl{config, logger, repo}
}

func (uc *getStorageUsageSummaryUseCaseImpl) Execute(ctx context.Context, req *GetStorageUsageSummaryRequest) (*storagedailyusage.StorageUsageSummary, error) {
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
		if req.SummaryType == "" {
			e["summary_type"] = "Summary type is required"
		} else if req.SummaryType != "current_month" && req.SummaryType != "current_year" {
			e["summary_type"] = "Summary type must be one of: current_month, current_year"
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get storage usage summary",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get summary based on type.
	//

	var summary *storagedailyusage.StorageUsageSummary
	var err error

	switch req.SummaryType {
	case "current_month":
		summary, err = uc.repo.GetCurrentMonthUsage(ctx, req.UserID)

	case "current_year":
		summary, err = uc.repo.GetCurrentYearUsage(ctx, req.UserID)

	default:
		return nil, httperror.NewForBadRequestWithSingleField("summary_type", "Invalid summary type")
	}

	if err != nil {
		uc.logger.Error("Failed to get storage usage summary",
			zap.String("user_id", req.UserID.String()),
			zap.String("summary_type", req.SummaryType),
			zap.Error(err))
		return nil, err
	}

	uc.logger.Debug("Successfully retrieved storage usage summary",
		zap.String("user_id", req.UserID.String()),
		zap.String("summary_type", req.SummaryType),
		zap.Int64("current_usage", summary.CurrentUsage),
		zap.Int64("total_added", summary.TotalAdded),
		zap.Int64("total_removed", summary.TotalRemoved),
		zap.Int64("net_change", summary.NetChange),
		zap.Int("days_with_data", summary.DaysWithData))

	return summary, nil
}
