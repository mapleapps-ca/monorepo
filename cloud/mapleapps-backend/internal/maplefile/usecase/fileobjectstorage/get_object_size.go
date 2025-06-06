// cloud/backend/internal/maplefile/usecase/fileobjectstorage/get_object_size.go
package fileobjectstorage

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetObjectSizeUseCase interface {
	Execute(storagePath string) (int64, error)
}

type getObjectSizeUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileObjectStorageRepository
}

func NewGetObjectSizeUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileObjectStorageRepository,
) GetObjectSizeUseCase {
	logger = logger.Named("GetObjectSizeUseCase")
	return &getObjectSizeUseCaseImpl{config, logger, repo}
}

func (uc *getObjectSizeUseCaseImpl) Execute(storagePath string) (int64, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if storagePath == "" {
		e["storage_path"] = "Storage path is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get object size",
			zap.Any("error", e))
		return 0, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get object size.
	//

	size, err := uc.repo.GetObjectSize(storagePath)
	if err != nil {
		uc.logger.Error("Failed to get object size",
			zap.String("storage_path", storagePath),
			zap.Error(err))
		return 0, err
	}

	uc.logger.Debug("Retrieved object size",
		zap.String("storage_path", storagePath),
		zap.Int64("size", size))

	return size, nil
}
