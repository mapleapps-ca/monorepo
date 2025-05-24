// cloud/backend/internal/maplefile/usecase/fileobjectstorage/get_object_size.go
package fileobjectstorage

import (
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type VerifyObjectExistsUseCase interface {
	Execute(storagePath string) (bool, error)
}

type verifyObjectExistsUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileObjectStorageRepository
}

func NewVerifyObjectExistsUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileObjectStorageRepository,
) VerifyObjectExistsUseCase {
	return &verifyObjectExistsUseCaseImpl{config, logger, repo}
}

func (uc *verifyObjectExistsUseCaseImpl) Execute(storagePath string) (bool, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if storagePath == "" {
		e["storage_path"] = "Storage path is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating verify if object exists",
			zap.Any("error", e))
		return false, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get object size.
	//

	exists, err := uc.repo.VerifyObjectExists(storagePath)
	if err != nil {
		uc.logger.Error("Failed to verify if object exists",
			zap.String("storage_path", storagePath),
			zap.Error(err))
		return false, err
	}

	return exists, nil
}
