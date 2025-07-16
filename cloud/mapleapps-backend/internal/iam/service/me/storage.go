// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service/me/storage.go
package me

import (
	"context"
	"errors"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
	storage_utils "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/utils"
)

// GetStorageUsageResponseDTO represents the storage usage response
type GetStorageUsageResponseDTO struct {
	UserID                gocql.UUID                  `json:"user_id"`
	UserPlan              string                      `json:"user_plan"`
	StorageUsedBytes      int64                       `json:"storage_used_bytes"`
	StorageUsedFormatted  storage_utils.FormattedSize `json:"storage_used_formatted"`
	StorageLimitBytes     int64                       `json:"storage_limit_bytes"`
	StorageLimitFormatted storage_utils.FormattedSize `json:"storage_limit_formatted"`
	RemainingBytes        int64                       `json:"remaining_bytes"`
	RemainingFormatted    storage_utils.FormattedSize `json:"remaining_formatted"`
	UsagePercentage       float64                     `json:"usage_percentage"`
}

// UpgradePlanRequestDTO represents the plan upgrade request
type UpgradePlanRequestDTO struct {
	NewPlan string `json:"new_plan"`
}

// UpgradePlanResponseDTO represents the plan upgrade response
type UpgradePlanResponseDTO struct {
	Success           bool                        `json:"success"`
	Message           string                      `json:"message"`
	OldPlan           string                      `json:"old_plan"`
	NewPlan           string                      `json:"new_plan"`
	NewLimitBytes     int64                       `json:"new_limit_bytes"`
	NewLimitFormatted storage_utils.FormattedSize `json:"new_limit_formatted"`
}

// GetStorageUsageService handles getting storage usage
type GetStorageUsageService interface {
	Execute(ctx context.Context) (*GetStorageUsageResponseDTO, error)
}

// UpgradePlanService handles plan upgrades
type UpgradePlanService interface {
	Execute(ctx context.Context, req *UpgradePlanRequestDTO) (*UpgradePlanResponseDTO, error)
}

type getStorageUsageServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	storageUseCase uc_user.FederatedUserStorageManagementUseCase
}

type upgradePlanServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	storageUseCase uc_user.FederatedUserStorageManagementUseCase
	userGetByID    uc_user.FederatedUserGetByIDUseCase
}

func NewGetStorageUsageService(
	config *config.Configuration,
	logger *zap.Logger,
	storageUseCase uc_user.FederatedUserStorageManagementUseCase,
) GetStorageUsageService {
	logger = logger.Named("GetStorageUsageService")
	return &getStorageUsageServiceImpl{
		config:         config,
		logger:         logger,
		storageUseCase: storageUseCase,
	}
}

func NewUpgradePlanService(
	config *config.Configuration,
	logger *zap.Logger,
	storageUseCase uc_user.FederatedUserStorageManagementUseCase,
	userGetByID uc_user.FederatedUserGetByIDUseCase,
) UpgradePlanService {
	logger = logger.Named("UpgradePlanService")
	return &upgradePlanServiceImpl{
		config:         config,
		logger:         logger,
		storageUseCase: storageUseCase,
		userGetByID:    userGetByID,
	}
}

func (s *getStorageUsageServiceImpl) Execute(ctx context.Context) (*GetStorageUsageResponseDTO, error) {
	// Get user ID from context
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		s.logger.Error("Failed getting user ID from context")
		return nil, errors.New("user ID not found in context")
	}

	// Get storage usage
	usage, err := s.storageUseCase.GetStorageUsage(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get storage usage",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	// Format the sizes
	return &GetStorageUsageResponseDTO{
		UserID:                userID,
		UserPlan:              usage.UserPlan,
		StorageUsedBytes:      usage.StorageUsedBytes,
		StorageUsedFormatted:  storage_utils.FormatBytes(usage.StorageUsedBytes),
		StorageLimitBytes:     usage.StorageLimitBytes,
		StorageLimitFormatted: storage_utils.FormatBytes(usage.StorageLimitBytes),
		RemainingBytes:        usage.RemainingBytes,
		RemainingFormatted:    storage_utils.FormatBytes(usage.RemainingBytes),
		UsagePercentage:       usage.UsagePercentage,
	}, nil
}

func (s *upgradePlanServiceImpl) Execute(ctx context.Context, req *UpgradePlanRequestDTO) (*UpgradePlanResponseDTO, error) {
	// Validate request
	if req == nil || req.NewPlan == "" {
		return nil, httperror.NewForBadRequestWithSingleField("new_plan", "New plan is required")
	}

	// Get user ID from context
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		s.logger.Error("Failed getting user ID from context")
		return nil, errors.New("user ID not found in context")
	}

	// Get current user to know the old plan
	user, err := s.userGetByID.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	oldPlan := user.UserPlan

	// Upgrade the plan
	if err := s.storageUseCase.UpgradeUserPlan(ctx, userID, req.NewPlan); err != nil {
		s.logger.Error("Failed to upgrade plan",
			zap.String("user_id", userID.String()),
			zap.String("new_plan", req.NewPlan),
			zap.Error(err))
		return nil, err
	}

	// Get updated user to get new limit
	updatedUser, err := s.userGetByID.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UpgradePlanResponseDTO{
		Success:           true,
		Message:           "Plan upgraded successfully",
		OldPlan:           oldPlan,
		NewPlan:           req.NewPlan,
		NewLimitBytes:     updatedUser.StorageLimitBytes,
		NewLimitFormatted: storage_utils.FormatBytes(updatedUser.StorageLimitBytes),
	}, nil
}
