// monorepo/cloud/backend/internal/maplefile/usecase/filemetadata/delete.go
package filemetadata

import (
	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type SoftDeleteFileMetadataUseCase interface {
	Execute(id gocql.UUID) error
}

type softDeleteFileMetadataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewSoftDeleteFileMetadataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) SoftDeleteFileMetadataUseCase {
	logger = logger.Named("SoftDeleteFileMetadataUseCase")
	return &softDeleteFileMetadataUseCaseImpl{config, logger, repo}
}

func (uc *softDeleteFileMetadataUseCaseImpl) Execute(id gocql.UUID) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.String() == "" {
		e["id"] = "File ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata deletion",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Soft-delete from database.
	//

	return uc.repo.SoftDelete(id)
}
