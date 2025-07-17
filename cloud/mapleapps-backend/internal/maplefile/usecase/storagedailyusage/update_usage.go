// cloud/mapleapps-backend/internal/maplefile/usecase/storagedailyusage/update_usage.go
package storagedailyusage

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// UpdateStorageUsageRequest contains the update parameters
type UpdateStorageUsageRequest struct {
	UserID      gocql.UUID `json:"user_id"`
	UsageDay    *time.Time `json:"usage_day,omitempty"` // Optional, defaults to today
	TotalBytes  int64      `json:"total_bytes"`
	AddBytes    int64      `json:"add_bytes"`
	RemoveBytes int64      `json:"remove_bytes"`
	IsIncrement bool       `json:"is_increment"` // If true, increment existing values; if false, set absolute values
}

type UpdateStorageUsageUseCase interface {
	Execute(ctx context.Context, req *UpdateStorageUsageRequest) error
}

type updateStorageUsageUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   storagedailyusage.StorageDailyUsageRepository
}

func NewUpdateStorageUsageUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo storagedailyusage.StorageDailyUsageRepository,
) UpdateStorageUsageUseCase {
	logger = logger.Named("UpdateStorageUsageUseCase")
	return &updateStorageUsageUseCaseImpl{config, logger, repo}
}

func (uc *updateStorageUsageUseCaseImpl) Execute(ctx context.Context, req *UpdateStorageUsageRequest) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if req == nil {
		e["request"] = "Request is required"
	} else {
		if req.UserID.String() == "" {
			e["user_id"] = "User ID is required"
		}
		if req.AddBytes < 0 {
			e["add_bytes"] = "Add bytes cannot be negative"
		}
		if req.RemoveBytes < 0 {
			e["remove_bytes"] = "Remove bytes cannot be negative"
		}
		if !req.IsIncrement && req.TotalBytes < 0 {
			e["total_bytes"] = "Total bytes cannot be negative when setting absolute values"
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating update storage usage",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Set usage day if not provided.
	//

	usageDay := time.Now().Truncate(24 * time.Hour)
	if req.UsageDay != nil {
		usageDay = req.UsageDay.Truncate(24 * time.Hour)
	}

	//
	// STEP 3: Update or increment usage.
	//

	var err error

	if req.IsIncrement {
		// Increment existing values
		err = uc.repo.IncrementUsage(ctx, req.UserID, usageDay, req.TotalBytes, req.AddBytes, req.RemoveBytes)
	} else {
		// Set absolute values
		usage := &storagedailyusage.StorageDailyUsage{
			UserID:           req.UserID,
			UsageDay:         usageDay,
			TotalBytes:       req.TotalBytes,
			TotalAddBytes:    req.AddBytes,
			TotalRemoveBytes: req.RemoveBytes,
		}
		err = uc.repo.UpdateOrCreate(ctx, usage)
	}

	if err != nil {
		uc.logger.Error("Failed to update storage usage",
			zap.String("user_id", req.UserID.String()),
			zap.Time("usage_day", usageDay),
			zap.Int64("total_bytes", req.TotalBytes),
			zap.Int64("add_bytes", req.AddBytes),
			zap.Int64("remove_bytes", req.RemoveBytes),
			zap.Bool("is_increment", req.IsIncrement),
			zap.Error(err))
		return err
	}

	uc.logger.Debug("Successfully updated storage usage",
		zap.String("user_id", req.UserID.String()),
		zap.Time("usage_day", usageDay),
		zap.Int64("total_bytes", req.TotalBytes),
		zap.Int64("add_bytes", req.AddBytes),
		zap.Int64("remove_bytes", req.RemoveBytes),
		zap.Bool("is_increment", req.IsIncrement))

	return nil
}
