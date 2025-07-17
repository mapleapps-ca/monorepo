// cloud/mapleapps-backend/internal/maplefile/service/storagedailyusage/get_trend.go
package storagedailyusage

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	uc_storagedailyusage "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/storagedailyusage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetStorageDailyUsageTrendRequestDTO struct {
	TrendPeriod string      `json:"trend_period"` // "7days", "monthly", "yearly"
	Year        *int        `json:"year,omitempty"`
	Month       *time.Month `json:"month,omitempty"`
}

type StorageDailyUsageResponseDTO struct {
	UserID           gocql.UUID `json:"user_id"`
	UsageDay         time.Time  `json:"usage_day"`
	TotalBytes       int64      `json:"total_bytes"`
	TotalAddBytes    int64      `json:"total_add_bytes"`
	TotalRemoveBytes int64      `json:"total_remove_bytes"`
}

type StorageUsageTrendResponseDTO struct {
	UserID          gocql.UUID                      `json:"user_id"`
	StartDate       time.Time                       `json:"start_date"`
	EndDate         time.Time                       `json:"end_date"`
	DailyUsages     []*StorageDailyUsageResponseDTO `json:"daily_usages"`
	TotalAdded      int64                           `json:"total_added"`
	TotalRemoved    int64                           `json:"total_removed"`
	NetChange       int64                           `json:"net_change"`
	AverageDailyAdd int64                           `json:"average_daily_add"`
	PeakUsageDay    *time.Time                      `json:"peak_usage_day,omitempty"`
	PeakUsageBytes  int64                           `json:"peak_usage_bytes"`
}

type GetStorageDailyUsageTrendResponseDTO struct {
	TrendPeriod string                        `json:"trend_period"`
	Trend       *StorageUsageTrendResponseDTO `json:"trend"`
	Success     bool                          `json:"success"`
	Message     string                        `json:"message"`
}

type GetStorageDailyUsageTrendService interface {
	Execute(ctx context.Context, req *GetStorageDailyUsageTrendRequestDTO) (*GetStorageDailyUsageTrendResponseDTO, error)
}

type getStorageDailyUsageTrendServiceImpl struct {
	config                           *config.Configuration
	logger                           *zap.Logger
	getStorageDailyUsageTrendUseCase uc_storagedailyusage.GetStorageDailyUsageTrendUseCase
}

func NewGetStorageDailyUsageTrendService(
	config *config.Configuration,
	logger *zap.Logger,
	getStorageDailyUsageTrendUseCase uc_storagedailyusage.GetStorageDailyUsageTrendUseCase,
) GetStorageDailyUsageTrendService {
	logger = logger.Named("GetStorageDailyUsageTrendService")
	return &getStorageDailyUsageTrendServiceImpl{
		config:                           config,
		logger:                           logger,
		getStorageDailyUsageTrendUseCase: getStorageDailyUsageTrendUseCase,
	}
}

func (svc *getStorageDailyUsageTrendServiceImpl) Execute(ctx context.Context, req *GetStorageDailyUsageTrendRequestDTO) (*GetStorageDailyUsageTrendResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Request details are required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Build use case request
	//
	useCaseReq := &uc_storagedailyusage.GetStorageDailyUsageTrendRequest{
		UserID:      userID,
		TrendPeriod: req.TrendPeriod,
		Year:        req.Year,
		Month:       req.Month,
	}

	//
	// STEP 4: Execute use case
	//
	trend, err := svc.getStorageDailyUsageTrendUseCase.Execute(ctx, useCaseReq)
	if err != nil {
		svc.logger.Error("Failed to get storage daily usage trend",
			zap.String("user_id", userID.String()),
			zap.String("trend_period", req.TrendPeriod),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Map domain models to response DTOs
	//
	dailyUsages := make([]*StorageDailyUsageResponseDTO, len(trend.DailyUsages))
	for i, usage := range trend.DailyUsages {
		dailyUsages[i] = &StorageDailyUsageResponseDTO{
			UserID:           usage.UserID,
			UsageDay:         usage.UsageDay,
			TotalBytes:       usage.TotalBytes,
			TotalAddBytes:    usage.TotalAddBytes,
			TotalRemoveBytes: usage.TotalRemoveBytes,
		}
	}

	trendResponse := &StorageUsageTrendResponseDTO{
		UserID:          trend.UserID,
		StartDate:       trend.StartDate,
		EndDate:         trend.EndDate,
		DailyUsages:     dailyUsages,
		TotalAdded:      trend.TotalAdded,
		TotalRemoved:    trend.TotalRemoved,
		NetChange:       trend.NetChange,
		AverageDailyAdd: trend.AverageDailyAdd,
		PeakUsageDay:    trend.PeakUsageDay,
		PeakUsageBytes:  trend.PeakUsageBytes,
	}

	response := &GetStorageDailyUsageTrendResponseDTO{
		TrendPeriod: req.TrendPeriod,
		Trend:       trendResponse,
		Success:     true,
		Message:     "Storage daily usage trend retrieved successfully",
	}

	svc.logger.Debug("Storage daily usage trend retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.String("trend_period", req.TrendPeriod),
		zap.Int("daily_usages_count", len(dailyUsages)),
		zap.Int64("net_change", trend.NetChange))

	return response, nil
}
