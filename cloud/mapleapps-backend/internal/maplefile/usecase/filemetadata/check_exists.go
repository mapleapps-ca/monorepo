// monorepo/cloud/backend/internal/maplefile/usecase/filemetadata/check_exists.go
package filemetadata

import (
	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type CheckFileExistsUseCase interface {
	Execute(id gocql.UUID) (bool, error)
}

type checkFileExistsUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewCheckFileExistsUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) CheckFileExistsUseCase {
	logger = logger.Named("CheckFileExistsUseCase")
	return &checkFileExistsUseCaseImpl{config, logger, repo}
}

func (uc *checkFileExistsUseCaseImpl) Execute(id gocql.UUID) (bool, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.String() == "" {
		e["id"] = "File ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file existence check",
			zap.Any("error", e))
		return false, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Check existence in database.
	//

	return uc.repo.CheckIfExistsByID(id)
}
