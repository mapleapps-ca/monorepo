// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser/storage.go
package federateduser

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// FederatedUserStorageManagementUseCase handles storage quota operations
type FederatedUserStorageManagementUseCase interface {
	// IncrementStorageUsage adds to the user's storage usage
	IncrementStorageUsage(ctx context.Context, userID gocql.UUID, bytes int64) error

	// DecrementStorageUsage removes from the user's storage usage
	DecrementStorageUsage(ctx context.Context, userID gocql.UUID, bytes int64) error

	// CheckStorageQuota checks if user has enough quota for upload
	CheckStorageQuota(ctx context.Context, userID gocql.UUID, requiredBytes int64) (bool, error)

	// GetStorageUsage returns current usage and limit
	GetStorageUsage(ctx context.Context, userID gocql.UUID) (*StorageUsageDTO, error)

	// UpgradeUserPlan upgrades the user's plan
	UpgradeUserPlan(ctx context.Context, userID gocql.UUID, newPlan string) error
}

// StorageUsageDTO represents storage usage information
type StorageUsageDTO struct {
	UserID            gocql.UUID `json:"user_id"`
	UserPlan          string     `json:"user_plan"`
	StorageUsedBytes  int64      `json:"storage_used_bytes"`
	StorageLimitBytes int64      `json:"storage_limit_bytes"`
	RemainingBytes    int64      `json:"remaining_bytes"`
	UsagePercentage   float64    `json:"usage_percentage"`
}

type federatedUserStorageManagementUseCaseImpl struct {
	config     *config.Configuration
	logger     *zap.Logger
	repo       dom_user.FederatedUserRepository
	getByID    FederatedUserGetByIDUseCase
	updateUser FederatedUserUpdateUseCase
}

func NewFederatedUserStorageManagementUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_user.FederatedUserRepository,
	getByID FederatedUserGetByIDUseCase,
	updateUser FederatedUserUpdateUseCase,
) FederatedUserStorageManagementUseCase {
	logger = logger.Named("FederatedUserStorageManagementUseCase")
	return &federatedUserStorageManagementUseCaseImpl{
		config:     config,
		logger:     logger,
		repo:       repo,
		getByID:    getByID,
		updateUser: updateUser,
	}
}

func (uc *federatedUserStorageManagementUseCaseImpl) IncrementStorageUsage(ctx context.Context, userID gocql.UUID, bytes int64) error {
	// Validate input
	if bytes < 0 {
		return httperror.NewForBadRequestWithSingleField("bytes", "Cannot increment by negative value")
	}

	// Get user
	user, err := uc.getByID.Execute(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return httperror.NewForBadRequestWithSingleField("user_id", "User not found")
	}

	// Check if user would exceed quota
	newUsage := user.StorageUsedBytes + bytes
	if newUsage > user.StorageLimitBytes {
		uc.logger.Warn("Storage quota exceeded",
			zap.String("user_id", userID.String()),
			zap.Int64("current_usage", user.StorageUsedBytes),
			zap.Int64("attempted_add", bytes),
			zap.Int64("limit", user.StorageLimitBytes))
		return httperror.NewForBadRequestWithSingleField("storage", "Storage quota exceeded")
	}

	// Update storage usage
	user.StorageUsedBytes = newUsage

	// Persist changes
	if err := uc.updateUser.Execute(ctx, user); err != nil {
		uc.logger.Error("Failed to update storage usage",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return err
	}

	uc.logger.Info("Storage usage incremented",
		zap.String("user_id", userID.String()),
		zap.Int64("added_bytes", bytes),
		zap.Int64("new_usage", newUsage))

	return nil
}

func (uc *federatedUserStorageManagementUseCaseImpl) DecrementStorageUsage(ctx context.Context, userID gocql.UUID, bytes int64) error {
	// Validate input
	if bytes < 0 {
		return httperror.NewForBadRequestWithSingleField("bytes", "Cannot decrement by negative value")
	}

	// Get user
	user, err := uc.getByID.Execute(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return httperror.NewForBadRequestWithSingleField("user_id", "User not found")
	}

	// Calculate new usage, ensure it doesn't go below 0
	newUsage := user.StorageUsedBytes - bytes
	if newUsage < 0 {
		newUsage = 0
	}

	// Update storage usage
	user.StorageUsedBytes = newUsage

	// Persist changes
	if err := uc.updateUser.Execute(ctx, user); err != nil {
		uc.logger.Error("Failed to update storage usage",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return err
	}

	uc.logger.Info("Storage usage decremented",
		zap.String("user_id", userID.String()),
		zap.Int64("removed_bytes", bytes),
		zap.Int64("new_usage", newUsage))

	return nil
}

func (uc *federatedUserStorageManagementUseCaseImpl) CheckStorageQuota(ctx context.Context, userID gocql.UUID, requiredBytes int64) (bool, error) {
	// Get user
	user, err := uc.getByID.Execute(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, httperror.NewForBadRequestWithSingleField("user_id", "User not found")
	}

	// Check if user has enough quota
	hasQuota := user.CanUpload(requiredBytes)

	if !hasQuota {
		uc.logger.Debug("Insufficient storage quota",
			zap.String("user_id", userID.String()),
			zap.Int64("current_usage", user.StorageUsedBytes),
			zap.Int64("required_bytes", requiredBytes),
			zap.Int64("limit", user.StorageLimitBytes))
	}

	return hasQuota, nil
}

func (uc *federatedUserStorageManagementUseCaseImpl) GetStorageUsage(ctx context.Context, userID gocql.UUID) (*StorageUsageDTO, error) {
	// Get user
	user, err := uc.getByID.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, httperror.NewForBadRequestWithSingleField("user_id", "User not found")
	}

	return &StorageUsageDTO{
		UserID:            user.ID,
		UserPlan:          user.UserPlan,
		StorageUsedBytes:  user.StorageUsedBytes,
		StorageLimitBytes: user.StorageLimitBytes,
		RemainingBytes:    user.GetRemainingStorage(),
		UsagePercentage:   user.GetStorageUsagePercentage(),
	}, nil
}

func (uc *federatedUserStorageManagementUseCaseImpl) UpgradeUserPlan(ctx context.Context, userID gocql.UUID, newPlan string) error {
	// Validate plan
	if _, exists := dom_user.FederatedUserPlanStorageLimits[newPlan]; !exists {
		return httperror.NewForBadRequestWithSingleField("plan", fmt.Sprintf("Invalid plan: %s", newPlan))
	}

	// Get user
	user, err := uc.getByID.Execute(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return httperror.NewForBadRequestWithSingleField("user_id", "User not found")
	}

	// Check if already on this plan
	if user.UserPlan == newPlan {
		return httperror.NewForBadRequestWithSingleField("plan", "User is already on this plan")
	}

	oldPlan := user.UserPlan

	// Upgrade plan
	user.UpgradePlan(newPlan)

	// Persist changes
	if err := uc.updateUser.Execute(ctx, user); err != nil {
		uc.logger.Error("Failed to upgrade user plan",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return err
	}

	uc.logger.Info("User plan upgraded",
		zap.String("user_id", userID.String()),
		zap.String("old_plan", oldPlan),
		zap.String("new_plan", newPlan),
		zap.Int64("new_limit", user.StorageLimitBytes))

	return nil
}
