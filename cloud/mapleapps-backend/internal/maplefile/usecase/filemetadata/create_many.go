// monorepo/cloud/backend/internal/maplefile/usecase/filemetadata/create_many.go
package filemetadata

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type CreateManyFileMetadataUseCase interface {
	Execute(files []*dom_file.File) error
}

type createManyFileMetadataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewCreateManyFileMetadataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) CreateManyFileMetadataUseCase {
	logger = logger.Named("CreateManyFileMetadataUseCase")
	return &createManyFileMetadataUseCaseImpl{config, logger, repo}
}

func (uc *createManyFileMetadataUseCaseImpl) Execute(files []*dom_file.File) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if files == nil || len(files) == 0 {
		e["files"] = "Files are required"
	} else {
		for i, file := range files {
			if file == nil {
				e[fmt.Sprintf("files[%d]", i)] = "File is required"
				continue
			}
			if file.CollectionID.String() == "" {
				e[fmt.Sprintf("files[%d].collection_id", i)] = "Collection ID is required"
			}
			if file.OwnerID.String() == "" {
				e[fmt.Sprintf("files[%d].owner_id", i)] = "Owner ID is required"
			}
			if file.EncryptedMetadata == "" {
				e[fmt.Sprintf("files[%d].encrypted_metadata", i)] = "Encrypted metadata is required"
			}
			if file.EncryptedFileKey.Ciphertext == nil || len(file.EncryptedFileKey.Ciphertext) == 0 {
				e[fmt.Sprintf("files[%d].encrypted_file_key", i)] = "Encrypted file key is required"
			}
			if file.EncryptionVersion == "" {
				e[fmt.Sprintf("files[%d].encryption_version", i)] = "Encryption version is required"
			}
			if file.EncryptedHash == "" {
				e[fmt.Sprintf("files[%d].encrypted_hash", i)] = "Encrypted hash is required"
			}
			if file.EncryptedFileObjectKey == "" {
				e[fmt.Sprintf("files[%d].encrypted_file_object_key", i)] = "Encrypted file object key is required"
			}
			if file.EncryptedFileSizeInBytes <= 0 {
				e[fmt.Sprintf("files[%d].encrypted_file_size_in_bytes", i)] = "Encrypted file size must be greater than 0"
			}
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata batch creation",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Insert into database.
	//

	return uc.repo.CreateMany(files)
}
