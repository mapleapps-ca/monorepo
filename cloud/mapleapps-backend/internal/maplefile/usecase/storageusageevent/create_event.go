// cloud/mapleapps-backend/internal/maplefile/usecase/storageusageevent/create_event.go
package storageusageevent

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storageusageevent"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type CreateStorageUsageEventUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID, fileSize int64, operation string) error
}

type createStorageUsageEventUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   storageusageevent.StorageUsageEventRepository
}

func NewCreateStorageUsageEventUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo storageusageevent.StorageUsageEventRepository,
) CreateStorageUsageEventUseCase {
	logger = logger.Named("CreateStorageUsageEventUseCase")
	return &createStorageUsageEventUseCaseImpl{config, logger, repo}
}

func (uc *createStorageUsageEventUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID, fileSize int64, operation string) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if fileSize <= 0 {
		e["file_size"] = "File size must be greater than 0"
	}
	if operation == "" {
		e["operation"] = "Operation is required"
	} else if operation != "add" && operation != "remove" {
		e["operation"] = "Operation must be 'add' or 'remove'"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating create storage usage event",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Create storage usage event.
	//

	now := time.Now()
	event := &storageusageevent.StorageUsageEvent{
		UserID:    userID,
		EventDay:  now.Truncate(24 * time.Hour),
		EventTime: now,
		FileSize:  fileSize,
		Operation: operation,
	}

	err := uc.repo.Create(ctx, event)
	if err != nil {
		uc.logger.Error("Failed to create storage usage event",
			zap.String("user_id", userID.String()),
			zap.Int64("file_size", fileSize),
			zap.String("operation", operation),
			zap.Error(err))
		return err
	}

	uc.logger.Debug("Successfully created storage usage event",
		zap.String("user_id", userID.String()),
		zap.Int64("file_size", fileSize),
		zap.String("operation", operation))

	return nil
}
