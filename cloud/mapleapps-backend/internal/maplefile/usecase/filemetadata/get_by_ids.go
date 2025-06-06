// cloud/backend/internal/maplefile/usecase/filemetadata/get_by_ids.go
package filemetadata

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetFileMetadataByIDsUseCase interface {
	Execute(ids []gocql.UUID) ([]*dom_file.File, error)
}

type getFileMetadataByIDsUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewGetFileMetadataByIDsUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) GetFileMetadataByIDsUseCase {
	logger = logger.Named("GetFileMetadataByIDsUseCase")
	return &getFileMetadataByIDsUseCaseImpl{config, logger, repo}
}

func (uc *getFileMetadataByIDsUseCaseImpl) Execute(ids []gocql.UUID) ([]*dom_file.File, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if ids == nil || len(ids) == 0 {
		e["ids"] = "File IDs are required"
	} else {
		for i, id := range ids {
			if id.String() == "" {
				e[fmt.Sprintf("ids[%d]", i)] = "File ID is required"
			}
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata retrieval by IDs",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	return uc.repo.GetByIDs(ids)
}
