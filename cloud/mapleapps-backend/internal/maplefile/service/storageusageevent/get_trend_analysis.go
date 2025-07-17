// monorepo/cloud/mapleapps-backend/internal/maplefile/service/storageusageevent/get_trend_analysis.go
package storageusageevent

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	uc_storageusageevent "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/storageusageevent"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetStorageUsageEventsTrendAnalysisRequestDTO struct {
	TrendPeriod string      `json:"trend_period"` // "7days", "monthly", "yearly", "custom"
	Year        *int        `json:"year,omitempty"`
	Month       *time.Month `json:"month,omitempty"`
	Days        *int        `json:"days,omitempty"` // For custom day ranges
}

type DailyStatsResponseDTO struct {
	Date         time.Time `json:"date"`
	AddEvents    int       `json:"add_events"`
	RemoveEvents int       `json:"remove_events"`
	BytesAdded   int64     `json:"bytes_added"`
	BytesRemoved int64     `json:"bytes_removed"`
	NetChange    int64     `json:"net_change"`
}

type GetStorageUsageEventsTrendAnalysisResponseDTO struct {
	UserID                gocql.UUID               `json:"user_id"`
	TrendPeriod           string                   `json:"trend_period"`
	StartDate             time.Time                `json:"start_date"`
	EndDate               time.Time                `json:"end_date"`
	TotalEvents           int                      `json:"total_events"`
	AddEvents             int                      `json:"add_events"`
	RemoveEvents          int                      `json:"remove_events"`
	TotalBytesAdded       int64                    `json:"total_bytes_added"`
	TotalBytesRemoved     int64                    `json:"total_bytes_removed"`
	NetBytesChange        int64                    `json:"net_bytes_change"`
	AverageBytesPerAdd    float64                  `json:"average_bytes_per_add"`
	AverageBytesPerRemove float64                  `json:"average_bytes_per_remove"`
	LargestAddEvent       int64                    `json:"largest_add_event"`
	LargestRemoveEvent    int64                    `json:"largest_remove_event"`
	DailyBreakdown        []*DailyStatsResponseDTO `json:"daily_breakdown,omitempty"`
	Success               bool                     `json:"success"`
	Message               string                   `json:"message"`
}

type GetStorageUsageEventsTrendAnalysisService interface {
	Execute(ctx context.Context, req *GetStorageUsageEventsTrendAnalysisRequestDTO) (*GetStorageUsageEventsTrendAnalysisResponseDTO, error)
}

type getStorageUsageEventsTrendAnalysisServiceImpl struct {
	config                                    *config.Configuration
	logger                                    *zap.Logger
	getStorageUsageEventsTrendAnalysisUseCase uc_storageusageevent.GetStorageUsageEventsTrendAnalysisUseCase
}

func NewGetStorageUsageEventsTrendAnalysisService(
	config *config.Configuration,
	logger *zap.Logger,
	getStorageUsageEventsTrendAnalysisUseCase uc_storageusageevent.GetStorageUsageEventsTrendAnalysisUseCase,
) GetStorageUsageEventsTrendAnalysisService {
	logger = logger.Named("GetStorageUsageEventsTrendAnalysisService")
	return &getStorageUsageEventsTrendAnalysisServiceImpl{
		config: config,
		logger: logger,
		getStorageUsageEventsTrendAnalysisUseCase: getStorageUsageEventsTrendAnalysisUseCase,
	}
}

func (svc *getStorageUsageEventsTrendAnalysisServiceImpl) Execute(ctx context.Context, req *GetStorageUsageEventsTrendAnalysisRequestDTO) (*GetStorageUsageEventsTrendAnalysisResponseDTO, error) {
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
	useCaseReq := &uc_storageusageevent.GetStorageUsageEventsRequest{
		UserID:      userID,
		TrendPeriod: req.TrendPeriod,
		Year:        req.Year,
		Month:       req.Month,
		Days:        req.Days,
	}

	//
	// STEP 4: Execute use case
	//
	analysis, err := svc.getStorageUsageEventsTrendAnalysisUseCase.Execute(ctx, useCaseReq)
	if err != nil {
		svc.logger.Error("Failed to get storage usage events trend analysis",
			zap.String("user_id", userID.String()),
			zap.String("trend_period", req.TrendPeriod),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Map domain models to response DTOs
	//
	dailyBreakdown := make([]*DailyStatsResponseDTO, len(analysis.DailyBreakdown))
	for i, daily := range analysis.DailyBreakdown {
		dailyBreakdown[i] = &DailyStatsResponseDTO{
			Date:         daily.Date,
			AddEvents:    daily.AddEvents,
			RemoveEvents: daily.RemoveEvents,
			BytesAdded:   daily.BytesAdded,
			BytesRemoved: daily.BytesRemoved,
			NetChange:    daily.NetChange,
		}
	}

	response := &GetStorageUsageEventsTrendAnalysisResponseDTO{
		UserID:                analysis.UserID,
		TrendPeriod:           analysis.TrendPeriod,
		StartDate:             analysis.StartDate,
		EndDate:               analysis.EndDate,
		TotalEvents:           analysis.TotalEvents,
		AddEvents:             analysis.AddEvents,
		RemoveEvents:          analysis.RemoveEvents,
		TotalBytesAdded:       analysis.TotalBytesAdded,
		TotalBytesRemoved:     analysis.TotalBytesRemoved,
		NetBytesChange:        analysis.NetBytesChange,
		AverageBytesPerAdd:    analysis.AverageBytesPerAdd,
		AverageBytesPerRemove: analysis.AverageBytesPerRemove,
		LargestAddEvent:       analysis.LargestAddEvent,
		LargestRemoveEvent:    analysis.LargestRemoveEvent,
		DailyBreakdown:        dailyBreakdown,
		Success:               true,
		Message:               "Storage usage events trend analysis completed successfully",
	}

	svc.logger.Debug("Storage usage events trend analysis completed successfully",
		zap.String("user_id", userID.String()),
		zap.String("trend_period", req.TrendPeriod),
		zap.Int("total_events", analysis.TotalEvents),
		zap.Int64("net_bytes_change", analysis.NetBytesChange))

	return response, nil
}
