// cloud/mapleapps-backend/internal/maplefile/service/storagedailyusage/get_usage_summary.go
package storagedailyusage

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	uc_storagedailyusage "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/storagedailyusage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetStorageUsageSummaryRequestDTO struct {
	SummaryType string `json:"summary_type"` // "current_month", "current_year"
}

type StorageUsageSummaryResponseDTO struct {
	UserID       gocql.UUID `json:"user_id"`
	Period       string     `json:"period"`
	StartDate    string     `json:"start_date"`
	EndDate      string     `json:"end_date"`
	CurrentUsage int64      `json:"current_usage_bytes"`
	TotalAdded   int64      `json:"total_added_bytes"`
	TotalRemoved int64      `json:"total_removed_bytes"`
	NetChange    int64      `json:"net_change_bytes"`
	DaysWithData int        `json:"days_with_data"`
}

type GetStorageUsageSummaryResponseDTO struct {
	SummaryType string                          `json:"summary_type"`
	Summary     *StorageUsageSummaryResponseDTO `json:"summary"`
	Success     bool                            `json:"success"`
	Message     string                          `json:"message"`
}

type GetStorageUsageSummaryService interface {
	Execute(ctx context.Context, req *GetStorageUsageSummaryRequestDTO) (*GetStorageUsageSummaryResponseDTO, error)
}

type getStorageUsageSummaryServiceImpl struct {
	config                        *config.Configuration
	logger                        *zap.Logger
	getStorageUsageSummaryUseCase uc_storagedailyusage.GetStorageUsageSummaryUseCase
}

func NewGetStorageUsageSummaryService(
	config *config.Configuration,
	logger *zap.Logger,
	getStorageUsageSummaryUseCase uc_storagedailyusage.GetStorageUsageSummaryUseCase,
) GetStorageUsageSummaryService {
	logger = logger.Named("GetStorageUsageSummaryService")
	return &getStorageUsageSummaryServiceImpl{
		config:                        config,
		logger:                        logger,
		getStorageUsageSummaryUseCase: getStorageUsageSummaryUseCase,
	}
}

func (svc *getStorageUsageSummaryServiceImpl) Execute(ctx context.Context, req *GetStorageUsageSummaryRequestDTO) (*GetStorageUsageSummaryResponseDTO, error) {
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
	useCaseReq := &uc_storagedailyusage.GetStorageUsageSummaryRequest{
		UserID:      userID,
		SummaryType: req.SummaryType,
	}

	//
	// STEP 4: Execute use case
	//
	summary, err := svc.getStorageUsageSummaryUseCase.Execute(ctx, useCaseReq)
	if err != nil {
		svc.logger.Error("Failed to get storage usage summary",
			zap.String("user_id", userID.String()),
			zap.String("summary_type", req.SummaryType),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Map domain model to response DTO
	//
	summaryResponse := &StorageUsageSummaryResponseDTO{
		UserID:       summary.UserID,
		Period:       summary.Period,
		StartDate:    summary.StartDate.Format("2006-01-02"),
		EndDate:      summary.EndDate.Format("2006-01-02"),
		CurrentUsage: summary.CurrentUsage,
		TotalAdded:   summary.TotalAdded,
		TotalRemoved: summary.TotalRemoved,
		NetChange:    summary.NetChange,
		DaysWithData: summary.DaysWithData,
	}

	response := &GetStorageUsageSummaryResponseDTO{
		SummaryType: req.SummaryType,
		Summary:     summaryResponse,
		Success:     true,
		Message:     "Storage usage summary retrieved successfully",
	}

	svc.logger.Debug("Storage usage summary retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.String("summary_type", req.SummaryType),
		zap.Int64("current_usage", summary.CurrentUsage),
		zap.Int64("net_change", summary.NetChange))

	return response, nil
}
