// monorepo/cloud/mapleapps-backend/internal/maplefile/service/storageusageevent/create_event.go
package storageusageevent

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	uc_storageusageevent "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/storageusageevent"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type CreateStorageUsageEventRequestDTO struct {
	FileSize  int64  `json:"file_size"`
	Operation string `json:"operation"` // "add" or "remove"
}

type CreateStorageUsageEventResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CreateStorageUsageEventService interface {
	Execute(ctx context.Context, req *CreateStorageUsageEventRequestDTO) (*CreateStorageUsageEventResponseDTO, error)
}

type createStorageUsageEventServiceImpl struct {
	config                         *config.Configuration
	logger                         *zap.Logger
	createStorageUsageEventUseCase uc_storageusageevent.CreateStorageUsageEventUseCase
}

func NewCreateStorageUsageEventService(
	config *config.Configuration,
	logger *zap.Logger,
	createStorageUsageEventUseCase uc_storageusageevent.CreateStorageUsageEventUseCase,
) CreateStorageUsageEventService {
	logger = logger.Named("CreateStorageUsageEventService")
	return &createStorageUsageEventServiceImpl{
		config:                         config,
		logger:                         logger,
		createStorageUsageEventUseCase: createStorageUsageEventUseCase,
	}
}

func (svc *createStorageUsageEventServiceImpl) Execute(ctx context.Context, req *CreateStorageUsageEventRequestDTO) (*CreateStorageUsageEventResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Event details are required")
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
	// STEP 3: Execute use case
	//
	err := svc.createStorageUsageEventUseCase.Execute(ctx, userID, req.FileSize, req.Operation)
	if err != nil {
		svc.logger.Error("Failed to create storage usage event",
			zap.String("user_id", userID.String()),
			zap.Int64("file_size", req.FileSize),
			zap.String("operation", req.Operation),
			zap.Error(err))
		return nil, err
	}

	response := &CreateStorageUsageEventResponseDTO{
		Success: true,
		Message: "Storage usage event created successfully",
	}

	svc.logger.Debug("Storage usage event created successfully",
		zap.String("user_id", userID.String()),
		zap.Int64("file_size", req.FileSize),
		zap.String("operation", req.Operation))

	return response, nil
}
