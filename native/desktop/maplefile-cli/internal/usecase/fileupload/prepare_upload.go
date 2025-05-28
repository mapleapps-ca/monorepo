// native/desktop/maplefile-cli/internal/usecase/fileupload/prepare_upload.go
package fileupload

import (
	"context"
	"encoding/json"

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

	// Prepare metadata
	metadata := map[string]interface{}{
		"name":      file.Name,
		"mime_type": file.MimeType,
		"size":      file.FileSize,
	}

	// Convert to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.NewAppError("failed to marshal metadata", err)
	}

	// File key should already be encrypted in the file record
	// Convert domain EncryptedFileKey to DTO format
	encryptedFileKey := filedto.EncryptedFileKey{
		Ciphertext: file.EncryptedFileKey.Ciphertext,
		Nonce:      file.EncryptedFileKey.Nonce,
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
		EncryptedMetadata:            crypto.EncodeToBase64(metadataJSON),
		EncryptedFileKey:             encryptedFileKey,
		EncryptionVersion:            "v1",
		EncryptedHash:                file.EncryptedHash,
		ExpectedFileSizeInBytes:      expectedFileSize,
		ExpectedThumbnailSizeInBytes: file.EncryptedThumbnailSize,
	}

	return request, nil
}
