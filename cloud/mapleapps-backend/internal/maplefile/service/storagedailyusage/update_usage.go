// monorepo/cloud/mapleapps-backend/internal/maplefile/service/storagedailyusage/update_usage.go
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

type UpdateStorageUsageRequestDTO struct {
	UsageDay    *time.Time `json:"usage_day,omitempty"` // Optional, defaults to today
	TotalBytes  int64      `json:"total_bytes"`
	AddBytes    int64      `json:"add_bytes"`
	RemoveBytes int64      `json:"remove_bytes"`
	IsIncrement bool       `json:"is_increment"` // If true, increment existing values; if false, set absolute values
}

type UpdateStorageUsageResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UpdateStorageUsageService interface {
	Execute(ctx context.Context, req *UpdateStorageUsageRequestDTO) (*UpdateStorageUsageResponseDTO, error)
}

type updateStorageUsageServiceImpl struct {
	config                    *config.Configuration
	logger                    *zap.Logger
	updateStorageUsageUseCase uc_storagedailyusage.UpdateStorageUsageUseCase
}

func NewUpdateStorageUsageService(
	config *config.Configuration,
	logger *zap.Logger,
	updateStorageUsageUseCase uc_storagedailyusage.UpdateStorageUsageUseCase,
) UpdateStorageUsageService {
	logger = logger.Named("UpdateStorageUsageService")
	return &updateStorageUsageServiceImpl{
		config:                    config,
		logger:                    logger,
		updateStorageUsageUseCase: updateStorageUsageUseCase,
	}
}

func (svc *updateStorageUsageServiceImpl) Execute(ctx context.Context, req *UpdateStorageUsageRequestDTO) (*UpdateStorageUsageResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Update details are required")
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
	useCaseReq := &uc_storagedailyusage.UpdateStorageUsageRequest{
		UserID:      userID,
		UsageDay:    req.UsageDay,
		TotalBytes:  req.TotalBytes,
		AddBytes:    req.AddBytes,
		RemoveBytes: req.RemoveBytes,
		IsIncrement: req.IsIncrement,
	}

	//
	// STEP 4: Execute use case
	//
	err := svc.updateStorageUsageUseCase.Execute(ctx, useCaseReq)
	if err != nil {
		svc.logger.Error("Failed to update storage usage",
			zap.String("user_id", userID.String()),
			zap.Int64("total_bytes", req.TotalBytes),
			zap.Int64("add_bytes", req.AddBytes),
			zap.Int64("remove_bytes", req.RemoveBytes),
			zap.Bool("is_increment", req.IsIncrement),
			zap.Error(err))
		return nil, err
	}

	response := &UpdateStorageUsageResponseDTO{
		Success: true,
		Message: "Storage usage updated successfully",
	}

	svc.logger.Debug("Storage usage updated successfully",
		zap.String("user_id", userID.String()),
		zap.Int64("total_bytes", req.TotalBytes),
		zap.Int64("add_bytes", req.AddBytes),
		zap.Int64("remove_bytes", req.RemoveBytes),
		zap.Bool("is_increment", req.IsIncrement))

	return response, nil
}
