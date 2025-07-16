// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/quota/helper.go
package federateduser

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// FederatedUserStorageQuotaHelperUseCase provides helper methods for storage quota management
type FederatedUserStorageQuotaHelperUseCase interface {
	// CheckAndReserveQuota checks if user has quota and reserves it atomically
	CheckAndReserveQuota(ctx context.Context, userID gocql.UUID, fileSize int64) error

	// ReleaseQuota releases previously reserved quota (e.g., on upload failure)
	ReleaseQuota(ctx context.Context, userID gocql.UUID, fileSize int64) error

	// OnFileDeleted updates quota when a file is deleted
	OnFileDeleted(ctx context.Context, userID gocql.UUID, fileSize int64) error

	// OnFileUpdated handles quota changes when a file is updated
	OnFileUpdated(ctx context.Context, userID gocql.UUID, oldSize, newSize int64) error

	// GetQuotaInfo returns formatted quota information
	GetQuotaInfo(ctx context.Context, userID gocql.UUID) (*QuotaInfo, error)
}

// QuotaInfo represents quota information
type QuotaInfo struct {
	HasQuota        bool    `json:"has_quota"`
	UsedBytes       int64   `json:"used_bytes"`
	LimitBytes      int64   `json:"limit_bytes"`
	RemainingBytes  int64   `json:"remaining_bytes"`
	UsagePercentage float64 `json:"usage_percentage"`
	UserPlan        string  `json:"user_plan"`
	Message         string  `json:"message,omitempty"`
}

type federatedUserStorageQuotaHelperUseCase struct {
	logger         *zap.Logger
	storageUseCase FederatedUserStorageManagementUseCase
}

// NewFederatedUserStorageQuotaHelperUseCase creates a new storage quota helper
func NewFederatedUserStorageQuotaHelperUseCase(
	logger *zap.Logger,
	storageUseCase FederatedUserStorageManagementUseCase,
) FederatedUserStorageQuotaHelperUseCase {
	logger = logger.Named("FederatedUserStorageQuotaHelperUseCase")
	return &federatedUserStorageQuotaHelperUseCase{
		logger:         logger,
		storageUseCase: storageUseCase,
	}
}

func (h *federatedUserStorageQuotaHelperUseCase) CheckAndReserveQuota(ctx context.Context, userID gocql.UUID, fileSize int64) error {
	// First check if user has enough quota
	hasQuota, err := h.storageUseCase.CheckStorageQuota(ctx, userID, fileSize)
	if err != nil {
		h.logger.Error("Failed to check storage quota",
			zap.String("user_id", userID.String()),
			zap.Int64("file_size", fileSize),
			zap.Error(err))
		return fmt.Errorf("failed to check storage quota: %w", err)
	}

	if !hasQuota {
		// Get current usage for better error message
		usage, err := h.storageUseCase.GetStorageUsage(ctx, userID)
		if err == nil {
			message := fmt.Sprintf(
				"Insufficient storage quota. Plan: %s, Used: %d/%d bytes (%.1f%%). File size: %d bytes",
				usage.UserPlan,
				usage.StorageUsedBytes,
				usage.StorageLimitBytes,
				usage.UsagePercentage,
				fileSize,
			)
			return httperror.NewForBadRequestWithSingleField("storage", message)
		}
		return httperror.NewForBadRequestWithSingleField("storage", "Insufficient storage quota")
	}

	// Reserve the quota
	if err := h.storageUseCase.IncrementStorageUsage(ctx, userID, fileSize); err != nil {
		h.logger.Error("Failed to reserve storage quota",
			zap.String("user_id", userID.String()),
			zap.Int64("file_size", fileSize),
			zap.Error(err))
		return fmt.Errorf("failed to reserve storage quota: %w", err)
	}

	h.logger.Info("Storage quota reserved",
		zap.String("user_id", userID.String()),
		zap.Int64("file_size", fileSize))

	return nil
}

func (h *federatedUserStorageQuotaHelperUseCase) ReleaseQuota(ctx context.Context, userID gocql.UUID, fileSize int64) error {
	if err := h.storageUseCase.DecrementStorageUsage(ctx, userID, fileSize); err != nil {
		h.logger.Error("Failed to release storage quota",
			zap.String("user_id", userID.String()),
			zap.Int64("file_size", fileSize),
			zap.Error(err))
		return fmt.Errorf("failed to release storage quota: %w", err)
	}

	h.logger.Info("Storage quota released",
		zap.String("user_id", userID.String()),
		zap.Int64("file_size", fileSize))

	return nil
}

func (h *federatedUserStorageQuotaHelperUseCase) OnFileDeleted(ctx context.Context, userID gocql.UUID, fileSize int64) error {
	return h.ReleaseQuota(ctx, userID, fileSize)
}

func (h *federatedUserStorageQuotaHelperUseCase) OnFileUpdated(ctx context.Context, userID gocql.UUID, oldSize, newSize int64) error {
	sizeDiff := newSize - oldSize

	if sizeDiff == 0 {
		// No change in size
		return nil
	}

	if sizeDiff > 0 {
		// File size increased, check and reserve additional quota
		return h.CheckAndReserveQuota(ctx, userID, sizeDiff)
	}

	// File size decreased, release the difference
	return h.ReleaseQuota(ctx, userID, -sizeDiff)
}

func (h *federatedUserStorageQuotaHelperUseCase) GetQuotaInfo(ctx context.Context, userID gocql.UUID) (*QuotaInfo, error) {
	usage, err := h.storageUseCase.GetStorageUsage(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get storage usage",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	info := &QuotaInfo{
		HasQuota:        usage.RemainingBytes > 0,
		UsedBytes:       usage.StorageUsedBytes,
		LimitBytes:      usage.StorageLimitBytes,
		RemainingBytes:  usage.RemainingBytes,
		UsagePercentage: usage.UsagePercentage,
		UserPlan:        usage.UserPlan,
	}

	// Add helpful messages based on usage
	if usage.UsagePercentage >= 95 {
		info.Message = "Storage quota is almost full. Consider upgrading your plan."
	} else if usage.UsagePercentage >= 80 {
		info.Message = fmt.Sprintf("You have used %.1f%% of your storage quota.", usage.UsagePercentage)
	}

	return info, nil
}
