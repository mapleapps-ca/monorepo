// cloud/mapleapps-backend/internal/maplefile/service/storagedailyusage/get_usage_by_date_range.go
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

type GetStorageUsageByDateRangeRequestDTO struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type DateRangeSummaryResponseDTO struct {
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

type GetStorageUsageByDateRangeResponseDTO struct {
	UserID      gocql.UUID                      `json:"user_id"`
	StartDate   time.Time                       `json:"start_date"`
	EndDate     time.Time                       `json:"end_date"`
	DailyUsages []*StorageDailyUsageResponseDTO `json:"daily_usages"`
	Summary     *DateRangeSummaryResponseDTO    `json:"summary"`
	Success     bool                            `json:"success"`
	Message     string                          `json:"message"`
}

type GetStorageUsageByDateRangeService interface {
	Execute(ctx context.Context, req *GetStorageUsageByDateRangeRequestDTO) (*GetStorageUsageByDateRangeResponseDTO, error)
}

type getStorageUsageByDateRangeServiceImpl struct {
	config                            *config.Configuration
	logger                            *zap.Logger
	getStorageUsageByDateRangeUseCase uc_storagedailyusage.GetStorageUsageByDateRangeUseCase
}

func NewGetStorageUsageByDateRangeService(
	config *config.Configuration,
	logger *zap.Logger,
	getStorageUsageByDateRangeUseCase uc_storagedailyusage.GetStorageUsageByDateRangeUseCase,
) GetStorageUsageByDateRangeService {
	logger = logger.Named("GetStorageUsageByDateRangeService")
	return &getStorageUsageByDateRangeServiceImpl{
		config:                            config,
		logger:                            logger,
		getStorageUsageByDateRangeUseCase: getStorageUsageByDateRangeUseCase,
	}
}

func (svc *getStorageUsageByDateRangeServiceImpl) Execute(ctx context.Context, req *GetStorageUsageByDateRangeRequestDTO) (*GetStorageUsageByDateRangeResponseDTO, error) {
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
	useCaseReq := &uc_storagedailyusage.GetStorageUsageByDateRangeRequest{
		UserID:    userID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	//
	// STEP 4: Execute use case
	//
	useCaseResp, err := svc.getStorageUsageByDateRangeUseCase.Execute(ctx, useCaseReq)
	if err != nil {
		svc.logger.Error("Failed to get storage usage by date range",
			zap.String("user_id", userID.String()),
			zap.Time("start_date", req.StartDate),
			zap.Time("end_date", req.EndDate),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Map domain models to response DTOs
	//
	dailyUsages := make([]*StorageDailyUsageResponseDTO, len(useCaseResp.DailyUsages))
	for i, usage := range useCaseResp.DailyUsages {
		dailyUsages[i] = &StorageDailyUsageResponseDTO{
			UserID:           usage.UserID,
			UsageDay:         usage.UsageDay,
			TotalBytes:       usage.TotalBytes,
			TotalAddBytes:    usage.TotalAddBytes,
			TotalRemoveBytes: usage.TotalRemoveBytes,
		}
	}

	summaryResponse := &DateRangeSummaryResponseDTO{
		TotalDays:        useCaseResp.Summary.TotalDays,
		DaysWithData:     useCaseResp.Summary.DaysWithData,
		TotalAdded:       useCaseResp.Summary.TotalAdded,
		TotalRemoved:     useCaseResp.Summary.TotalRemoved,
		NetChange:        useCaseResp.Summary.NetChange,
		AverageDailyAdd:  useCaseResp.Summary.AverageDailyAdd,
		PeakUsageDay:     useCaseResp.Summary.PeakUsageDay,
		PeakUsageBytes:   useCaseResp.Summary.PeakUsageBytes,
		LowestUsageDay:   useCaseResp.Summary.LowestUsageDay,
		LowestUsageBytes: useCaseResp.Summary.LowestUsageBytes,
	}

	response := &GetStorageUsageByDateRangeResponseDTO{
		UserID:      useCaseResp.UserID,
		StartDate:   useCaseResp.StartDate,
		EndDate:     useCaseResp.EndDate,
		DailyUsages: dailyUsages,
		Summary:     summaryResponse,
		Success:     true,
		Message:     "Storage usage by date range retrieved successfully",
	}

	svc.logger.Debug("Storage usage by date range retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.Time("start_date", req.StartDate),
		zap.Time("end_date", req.EndDate),
		zap.Int("daily_usages_count", len(dailyUsages)),
		zap.Int64("net_change", useCaseResp.Summary.NetChange))

	return response, nil
}
