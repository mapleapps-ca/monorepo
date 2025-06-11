// cloud/backend/internal/maplefile/usecase/filemetadata/get.go
package filemetadata

import (
	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetFileMetadataUseCase interface {
	Execute(id gocql.UUID) (*dom_file.File, error)
}

type getFileMetadataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewGetFileMetadataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) GetFileMetadataUseCase {
	logger = logger.Named("GetFileMetadataUseCase")
	return &getFileMetadataUseCaseImpl{config, logger, repo}
}

func (uc *getFileMetadataUseCaseImpl) Execute(id gocql.UUID) (*dom_file.File, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.String() == "" {
		e["id"] = "File ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata retrieval",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	file, err := uc.repo.Get(id)
	if err != nil {
		return nil, err
	}

	if file == nil {
		uc.logger.Debug("File metadata not found",
			zap.Any("id", id))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	return file, nil
}
