// cloud/backend/internal/maplefile/usecase/filemetadata/update.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type UpdateFileMetadataUseCase interface {
	Execute(ctx context.Context, file *dom_file.File) error
}

type updateFileMetadataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewUpdateFileMetadataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) UpdateFileMetadataUseCase {
	logger = logger.Named("UpdateFileMetadataUseCase")
	return &updateFileMetadataUseCaseImpl{config, logger, repo}
}

func (uc *updateFileMetadataUseCaseImpl) Execute(ctx context.Context, file *dom_file.File) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if file == nil {
		e["file"] = "File is required"
	} else {
		if file.ID.String() == "" {
			e["id"] = "File ID is required"
		}
		if file.CollectionID.String() == "" {
			e["collection_id"] = "Collection ID is required"
		}
		if file.OwnerID.String() == "" {
			e["owner_id"] = "Owner ID is required"
		}
		if file.EncryptedMetadata == "" {
			e["encrypted_metadata"] = "Encrypted metadata is required"
		}
		if file.EncryptedFileKey.Ciphertext == nil || len(file.EncryptedFileKey.Ciphertext) == 0 {
			e["encrypted_file_key"] = "Encrypted file key is required"
		}
		if file.EncryptionVersion == "" {
			e["encryption_version"] = "Encryption version is required"
		}
		if file.EncryptedHash == "" {
			e["encrypted_hash"] = "Encrypted hash is required"
		}
		if file.EncryptedFileObjectKey == "" {
			e["encrypted_file_object_key"] = "Encrypted file object key is required"
		}
		if file.EncryptedFileSizeInBytes <= 0 {
			e["encrypted_file_size_in_bytes"] = "Encrypted file size must be greater than 0"
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating file metadata update",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Update in database.
	//

	return uc.repo.Update(file)
}
