// native/desktop/maplefile-cli/internal/usecase/fileupload/prepare_upload.go
package fileupload

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// PrepareFileUploadUseCase prepares a file for upload
type PrepareFileUploadUseCase interface {
	Execute(ctx context.Context, file *dom_file.File, collection *dom_collection.Collection, collectionKey []byte) (*filedto.CreatePendingFileRequest, error)
}

type prepareFileUploadUseCase struct {
	logger *zap.Logger
}

func NewPrepareFileUploadUseCase(logger *zap.Logger) PrepareFileUploadUseCase {
	logger = logger.Named("PrepareFileUploadUseCase")
	return &prepareFileUploadUseCase{
		logger: logger,
	}
}

func (uc *prepareFileUploadUseCase) Execute(
	ctx context.Context,
	file *dom_file.File,
	collection *dom_collection.Collection,
	collectionKey []byte,
) (*filedto.CreatePendingFileRequest, error) {
	// Validate inputs
	if file == nil || collection == nil || len(collectionKey) == 0 {
		return nil, errors.NewAppError("invalid inputs", nil)
	}

	// Determine file size based on storage mode
	var expectedFileSize int64
	if file.StorageMode == dom_file.StorageModeEncryptedOnly || file.StorageMode == dom_file.StorageModeHybrid {
		expectedFileSize = file.EncryptedFileSize
	} else {
		// For decrypted-only mode, we'll encrypt on the fly
		// Add overhead for encryption
		expectedFileSize = file.FileSize + crypto.SecretBoxOverhead
	}

	// Create request
	request := &filedto.CreatePendingFileRequest{
		ID:                           file.ID,
		CollectionID:                 collection.ID,
		EncryptedMetadata:            file.EncryptedMetadata,
		EncryptedFileKey:             file.EncryptedFileKey,
		EncryptionVersion:            "v1",
		EncryptedHash:                file.EncryptedHash,
		ExpectedFileSizeInBytes:      expectedFileSize,
		ExpectedThumbnailSizeInBytes: file.EncryptedThumbnailSize,
	}

	return request, nil
}
