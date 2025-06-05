// internal/service/filedownload/download.go
package filedownload

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	svc_filecrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filecrypto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/filedto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// DecryptedFileMetadata represents decrypted file metadata
type DecryptedFileMetadata struct {
	Name     string `json:"name"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Created  int64  `json:"created"`
}

// DownloadResult represents the result of a file download with decryption
type DownloadResult struct {
	FileID            primitive.ObjectID     `json:"file_id"`
	DecryptedData     []byte                 `json:"decrypted_data"`
	DecryptedMetadata *DecryptedFileMetadata `json:"decrypted_metadata"`
	ThumbnailData     []byte                 `json:"thumbnail_data,omitempty"`
	OriginalSize      int64                  `json:"original_size"`
	ThumbnailSize     int64                  `json:"thumbnail_size"`
}

// DownloadService handles file download operations with E2EE decryption
type DownloadService interface {
	DownloadAndDecryptFile(ctx context.Context, fileID primitive.ObjectID, userPassword string, urlDuration time.Duration) (*DownloadResult, error)
}

type downloadService struct {
	logger                         *zap.Logger
	getPresignedDownloadURLUseCase filedto.GetPresignedDownloadURLUseCase
	downloadFileUseCase            filedto.DownloadFileUseCase
	getFileUseCase                 uc_file.GetFileUseCase
	getUserByIsLoggedInUseCase     uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase           uc_collection.GetCollectionUseCase
	collectionDecryptionService    svc_collectioncrypto.CollectionDecryptionService
	fileDecryptionService          svc_filecrypto.FileDecryptionService
}

func NewDownloadService(
	logger *zap.Logger,
	getPresignedDownloadURLUseCase filedto.GetPresignedDownloadURLUseCase,
	downloadFileUseCase filedto.DownloadFileUseCase,
	getFileUseCase uc_file.GetFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
	fileDecryptionService svc_filecrypto.FileDecryptionService,
) DownloadService {
	logger = logger.Named("DownloadService")
	return &downloadService{
		logger:                         logger,
		getPresignedDownloadURLUseCase: getPresignedDownloadURLUseCase,
		downloadFileUseCase:            downloadFileUseCase,
		getFileUseCase:                 getFileUseCase,
		getUserByIsLoggedInUseCase:     getUserByIsLoggedInUseCase,
		getCollectionUseCase:           getCollectionUseCase,
		collectionDecryptionService:    collectionDecryptionService,
		fileDecryptionService:          fileDecryptionService,
	}
}

func (s *downloadService) DownloadAndDecryptFile(ctx context.Context, fileID primitive.ObjectID, userPassword string, urlDuration time.Duration) (*DownloadResult, error) {
	s.logger.Info("👇 Starting E2EE file download and decryption", zap.String("fileID", fileID.Hex()))

	//
	// Step 1: Validate inputs
	//
	if fileID.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}
	if userPassword == "" {
		return nil, errors.NewAppError("user password is required for E2EE decryption", nil)
	}

	//
	// Step 2: Get file metadata (contains encrypted file key and metadata)
	//
	file, err := s.getFileUseCase.Execute(ctx, fileID)
	if err != nil {
		return nil, errors.NewAppError("failed to get file metadata", err)
	}
	if file == nil {
		return nil, errors.NewAppError("file not found", nil)
	}

	//
	// Step 3: Get user and collection for E2EE key chain
	//
	user, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get logged in user", err)
	}
	if user == nil {
		return nil, errors.NewAppError("user not found", nil)
	}

	collection, err := s.getCollectionUseCase.Execute(ctx, file.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection", err)
	}
	if collection == nil {
		return nil, errors.NewAppError("collection not found", nil)
	}

	//
	// Step 4: Decrypt the E2EE key chain to get collection key
	//
	collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, user, collection, userPassword)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}
	defer crypto.ClearBytes(collectionKey)

	//
	// Step 5: Decrypt the file key using collection key
	//
	fileKey, err := s.fileDecryptionService.DecryptFileKey(ctx, file.EncryptedFileKey, collectionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file key", err)
	}
	defer crypto.ClearBytes(fileKey)

	//
	// Step 6: Decrypt file metadata
	//
	decryptedMetadata, err := s.fileDecryptionService.DecryptFileMetadata(ctx, file.EncryptedMetadata, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file metadata", err)
	}

	//
	// Step 7: Get presigned download URLs
	//
	s.logger.Debug("🌐 Getting presigned download URLs")
	urlResponse, err := s.getPresignedDownloadURLUseCase.Execute(ctx, fileID, urlDuration)
	if err != nil {
		return nil, errors.NewAppError("failed to get presigned download URLs", err)
	}

	if !urlResponse.Success {
		return nil, errors.NewAppError("server failed to generate presigned URLs: "+urlResponse.Message, nil)
	}
	s.logger.Debug("✅ Successfully got presigned download URLs")

	//
	// Step 8: Download encrypted file content
	//
	s.logger.Debug("📥 Downloading encrypted file content")
	downloadRequest := &filedto.DownloadRequest{
		PresignedURL:          urlResponse.PresignedDownloadURL,
		PresignedThumbnailURL: urlResponse.PresignedThumbnailURL,
	}

	downloadResponse, err := s.downloadFileUseCase.Execute(ctx, downloadRequest)
	if err != nil {
		return nil, errors.NewAppError("failed to download file content", err)
	}
	s.logger.Debug("✅ Successfully downloaded encrypted file content")

	//
	// Step 9: Decrypt the file content
	//
	s.logger.Debug("🔑 Decrypting file content")
	decryptedData, err := s.fileDecryptionService.DecryptFileContent(ctx, downloadResponse.FileData, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file content", err)
	}
	s.logger.Debug("✅ Successfully decrypted file content")

	//
	// Step 10: Decrypt thumbnail if present
	//
	var thumbnailData []byte
	if downloadResponse.ThumbnailData != nil && len(downloadResponse.ThumbnailData) > 0 {
		s.logger.Debug("🔑 Decrypting thumbnail data")
		thumbnailData, err = s.fileDecryptionService.DecryptFileContent(ctx, downloadResponse.ThumbnailData, fileKey)
		if err != nil {
			s.logger.Warn("⚠️ Failed to decrypt thumbnail, continuing without it", zap.Error(err))
			thumbnailData = nil
		} else {
			s.logger.Debug("✅ Successfully decrypted thumbnail data")
		}
	}

	// Convert file metadata to the expected format
	resultMetadata := &DecryptedFileMetadata{
		Name:     decryptedMetadata.Name,
		MimeType: decryptedMetadata.MimeType,
		Size:     decryptedMetadata.Size,
		Created:  decryptedMetadata.Created,
	}

	result := &DownloadResult{
		FileID:            fileID,
		DecryptedData:     decryptedData,
		DecryptedMetadata: resultMetadata,
		ThumbnailData:     thumbnailData,
		OriginalSize:      int64(len(decryptedData)),
		ThumbnailSize:     int64(len(thumbnailData)),
	}

	s.logger.Info("✅ Successfully completed E2EE file download and decryption",
		zap.String("fileID", fileID.Hex()),
		zap.String("fileName", resultMetadata.Name),
		zap.Int64("originalSize", result.OriginalSize))

	return result, nil
}
