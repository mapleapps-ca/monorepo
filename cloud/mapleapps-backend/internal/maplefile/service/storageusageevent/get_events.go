// monorepo/cloud/mapleapps-backend/internal/maplefile/service/storageusageevent/get_events.go
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

type GetStorageUsageEventsRequestDTO struct {
	TrendPeriod string      `json:"trend_period"` // "7days", "monthly", "yearly", "custom"
	Year        *int        `json:"year,omitempty"`
	Month       *time.Month `json:"month,omitempty"`
	Days        *int        `json:"days,omitempty"` // For custom day ranges
}

type StorageUsageEventResponseDTO struct {
	UserID    gocql.UUID `json:"user_id"`
	EventDay  time.Time  `json:"event_day"`
	EventTime time.Time  `json:"event_time"`
	FileSize  int64      `json:"file_size"`
	Operation string     `json:"operation"`
}

type GetStorageUsageEventsResponseDTO struct {
	UserID      gocql.UUID                      `json:"user_id"`
	TrendPeriod string                          `json:"trend_period"`
	StartDate   time.Time                       `json:"start_date"`
	EndDate     time.Time                       `json:"end_date"`
	Events      []*StorageUsageEventResponseDTO `json:"events"`
	EventCount  int                             `json:"event_count"`
	Success     bool                            `json:"success"`
	Message     string                          `json:"message"`
}

type GetStorageUsageEventsService interface {
	Execute(ctx context.Context, req *GetStorageUsageEventsRequestDTO) (*GetStorageUsageEventsResponseDTO, error)
}

type getStorageUsageEventsServiceImpl struct {
	config                       *config.Configuration
	logger                       *zap.Logger
	getStorageUsageEventsUseCase uc_storageusageevent.GetStorageUsageEventsUseCase
}

func NewGetStorageUsageEventsService(
	config *config.Configuration,
	logger *zap.Logger,
	getStorageUsageEventsUseCase uc_storageusageevent.GetStorageUsageEventsUseCase,
) GetStorageUsageEventsService {
	logger = logger.Named("GetStorageUsageEventsService")
	return &getStorageUsageEventsServiceImpl{
		config:                       config,
		logger:                       logger,
		getStorageUsageEventsUseCase: getStorageUsageEventsUseCase,
	}
}

func (svc *getStorageUsageEventsServiceImpl) Execute(ctx context.Context, req *GetStorageUsageEventsRequestDTO) (*GetStorageUsageEventsResponseDTO, error) {
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
	useCaseResp, err := svc.getStorageUsageEventsUseCase.Execute(ctx, useCaseReq)
	if err != nil {
		svc.logger.Error("Failed to get storage usage events",
			zap.String("user_id", userID.String()),
			zap.String("trend_period", req.TrendPeriod),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Map domain models to response DTOs
	//
	events := make([]*StorageUsageEventResponseDTO, len(useCaseResp.Events))
	for i, event := range useCaseResp.Events {
		events[i] = &StorageUsageEventResponseDTO{
			UserID:    event.UserID,
			EventDay:  event.EventDay,
			EventTime: event.EventTime,
			FileSize:  event.FileSize,
			Operation: event.Operation,
		}
	}

	response := &GetStorageUsageEventsResponseDTO{
		UserID:      useCaseResp.UserID,
		TrendPeriod: useCaseResp.TrendPeriod,
		StartDate:   useCaseResp.StartDate,
		EndDate:     useCaseResp.EndDate,
		Events:      events,
		EventCount:  useCaseResp.EventCount,
		Success:     true,
		Message:     "Storage usage events retrieved successfully",
	}

	svc.logger.Debug("Storage usage events retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.String("trend_period", req.TrendPeriod),
		zap.Int("event_count", len(events)))

	return response, nil
}
