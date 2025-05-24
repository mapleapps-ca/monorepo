// native/desktop/maplefile-cli/internal/service/fileupload/upload.go
package fileupload

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/fileupload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_crypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/crypto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	uc_fileupload "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/fileupload"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	pkg_crypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// UploadService handles three-step file upload to cloud
type UploadService interface {
	UploadFile(ctx context.Context, fileID primitive.ObjectID) (*fileupload.FileUploadResult, error)
}

type uploadService struct {
	logger                   *zap.Logger
	fileDTORepo              filedto.FileDTORepository
	fileRepo                 dom_file.FileRepository
	collectionRepo           dom_collection.CollectionRepository
	userRepo                 user.Repository
	getFileUseCase           uc_file.GetFileUseCase
	updateFileUseCase        uc_file.UpdateFileUseCase
	getCollectionUseCase     uc_collection.GetCollectionUseCase
	getUserByLoggedInUseCase uc_user.GetByIsLoggedInUseCase
	encryptFileUseCase       uc_fileupload.EncryptFileUseCase
	prepareUploadUseCase     uc_fileupload.PrepareFileUploadUseCase
	cryptoService            svc_crypto.CryptoService
}

func NewUploadService(
	logger *zap.Logger,
	fileDTORepo filedto.FileDTORepository,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
	userRepo user.Repository,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	getUserByLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	encryptFileUseCase uc_fileupload.EncryptFileUseCase,
	prepareUploadUseCase uc_fileupload.PrepareFileUploadUseCase,
	cryptoService svc_crypto.CryptoService,
) UploadService {
	return &uploadService{
		logger:                   logger,
		fileDTORepo:              fileDTORepo,
		fileRepo:                 fileRepo,
		collectionRepo:           collectionRepo,
		userRepo:                 userRepo,
		getFileUseCase:           getFileUseCase,
		updateFileUseCase:        updateFileUseCase,
		getCollectionUseCase:     getCollectionUseCase,
		getUserByLoggedInUseCase: getUserByLoggedInUseCase,
		encryptFileUseCase:       encryptFileUseCase,
		prepareUploadUseCase:     prepareUploadUseCase,
		cryptoService:            cryptoService,
	}
}

func (s *uploadService) UploadFile(ctx context.Context, fileID primitive.ObjectID) (*fileupload.FileUploadResult, error) {
	// startTime := time.Now()

	s.logger.Info("Starting three-step file upload", zap.String("fileID", fileID.Hex()))

	// Step 0: Validate and prepare
	file, collection, collectionKey, err := s.validateAndPrepare(ctx, fileID)
	if err != nil {
		return s.failedResult(fileID, err)
	}

	// Step 1: Create pending file in cloud
	pendingResponse, err := s.createPendingFile(ctx, file, collection, collectionKey)
	if err != nil {
		return s.failedResult(fileID, err)
	}

	// Step 2: Upload file content
	fileSize, thumbnailSize, err := s.uploadContent(ctx, file, pendingResponse, collectionKey)
	if err != nil {
		return s.failedResult(fileID, err)
	}

	// Step 3: Complete upload
	if err := s.completeUpload(ctx, pendingResponse.File.ID, fileSize, thumbnailSize); err != nil {
		return s.failedResult(fileID, err)
	}

	// Update local file record
	if err := s.updateLocalFile(ctx, file, pendingResponse.File.ID); err != nil {
		s.logger.Error("Failed to update local file after successful upload",
			zap.String("fileID", fileID.Hex()),
			zap.Error(err))
	}

	return &fileupload.FileUploadResult{
		FileID:             fileID,
		CloudFileID:        pendingResponse.File.ID,
		UploadedAt:         time.Now(),
		FileSizeBytes:      fileSize,
		ThumbnailSizeBytes: thumbnailSize,
		Success:            true,
	}, nil
}

func (s *uploadService) validateAndPrepare(ctx context.Context, fileID primitive.ObjectID) (*dom_file.File, *dom_collection.Collection, []byte, error) {
	// Get file
	file, err := s.getFileUseCase.Execute(ctx, fileID)
	if err != nil {
		return nil, nil, nil, errors.NewAppError("failed to get file", err)
	}

	// Check sync status
	if file.SyncStatus != dom_file.SyncStatusLocalOnly {
		return nil, nil, nil, errors.NewAppError(
			fmt.Sprintf("file sync status must be LocalOnly, got: %v", file.SyncStatus),
			nil,
		)
	}

	// Get collection
	collection, err := s.getCollectionUseCase.Execute(ctx, file.CollectionID)
	if err != nil {
		return nil, nil, nil, errors.NewAppError("failed to get collection", err)
	}

	// Get logged in user
	user, err := s.getUserByLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, nil, nil, errors.NewAppError("failed to get logged in user", err)
	}
	_ = user // TODO: Use the user's master key to decrypt the collection key

	// TODO: In production, decrypt the collection key using the user's master key
	// For now, we'll generate a placeholder
	collectionKey, err := pkg_crypto.GenerateRandomBytes(pkg_crypto.CollectionKeySize)
	if err != nil {
		return nil, nil, nil, errors.NewAppError("failed to get collection key", err)
	}

	return file, collection, collectionKey, nil
}

func (s *uploadService) createPendingFile(
	ctx context.Context,
	file *dom_file.File,
	collection *dom_collection.Collection,
	collectionKey []byte,
) (*filedto.CreatePendingFileResponse, error) {
	s.logger.Debug("Creating pending file in cloud", zap.String("fileID", file.ID.Hex()))

	// Prepare upload request
	request, err := s.prepareUploadUseCase.Execute(ctx, file, collection, collectionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to prepare upload request", err)
	}

	// Create pending file
	response, err := s.fileDTORepo.CreatePendingFileInCloud(ctx, request)
	if err != nil {
		return nil, errors.NewAppError("failed to create pending file in cloud", err)
	}

	if !response.Success {
		return nil, errors.NewAppError(fmt.Sprintf("cloud rejected file creation: %s", response.Message), nil)
	}

	s.logger.Info("Created pending file in cloud",
		zap.String("cloudFileID", response.File.ID.Hex()),
		zap.Time("urlExpiration", response.UploadURLExpirationTime))

	return response, nil
}

func (s *uploadService) uploadContent(
	ctx context.Context,
	file *dom_file.File,
	pendingResponse *filedto.CreatePendingFileResponse,
	collectionKey []byte,
) (int64, int64, error) {
	s.logger.Debug("Uploading file content", zap.String("fileID", file.ID.Hex()))

	// Get file key by decrypting with collection key
	fileKey, err := s.cryptoService.DecryptFileKey(
		ctx,
		file.EncryptedFileKey.Ciphertext,
		file.EncryptedFileKey.Nonce,
		collectionKey,
	)
	if err != nil {
		return 0, 0, errors.NewAppError("failed to decrypt file key", err)
	}

	// Determine file data to upload
	var fileData []byte
	var actualFileSize int64

	switch file.StorageMode {
	case dom_file.StorageModeEncryptedOnly, dom_file.StorageModeHybrid:
		// Read encrypted file
		if file.EncryptedFilePath == "" {
			return 0, 0, errors.NewAppError("encrypted file path is empty", nil)
		}

		data, err := os.ReadFile(file.EncryptedFilePath)
		if err != nil {
			return 0, 0, errors.NewAppError("failed to read encrypted file", err)
		}
		fileData = data
		actualFileSize = int64(len(data))

	case dom_file.StorageModeDecryptedOnly:
		// Read and encrypt on the fly
		if file.FilePath == "" {
			return 0, 0, errors.NewAppError("file path is empty", nil)
		}

		encryptedData, hash, err := s.encryptFileUseCase.Execute(ctx, file.FilePath, fileKey)
		if err != nil {
			return 0, 0, errors.NewAppError("failed to encrypt file", err)
		}
		fileData = encryptedData
		actualFileSize = int64(len(encryptedData))
		_ = hash // TODO: Verify hash matches
	}

	// Upload file content
	if err := s.fileDTORepo.UploadFileToCloud(ctx, pendingResponse.PresignedUploadURL, fileData); err != nil {
		return 0, 0, errors.NewAppError("failed to upload file content", err)
	}

	// Upload thumbnail if exists
	var thumbnailSize int64
	if file.EncryptedThumbnailPath != "" && pendingResponse.PresignedThumbnailURL != "" {
		thumbnailData, err := os.ReadFile(file.EncryptedThumbnailPath)
		if err != nil {
			s.logger.Warn("Failed to read thumbnail",
				zap.String("path", file.EncryptedThumbnailPath),
				zap.Error(err))
		} else {
			if err := s.fileDTORepo.UploadThumbnailToCloud(ctx, pendingResponse.PresignedThumbnailURL, thumbnailData); err != nil {
				s.logger.Warn("Failed to upload thumbnail", zap.Error(err))
			} else {
				thumbnailSize = int64(len(thumbnailData))
			}
		}
	}

	// Clear sensitive data
	pkg_crypto.ClearBytes(fileKey)

	return actualFileSize, thumbnailSize, nil
}

func (s *uploadService) completeUpload(ctx context.Context, cloudFileID primitive.ObjectID, fileSize, thumbnailSize int64) error {
	s.logger.Debug("Completing file upload", zap.String("cloudFileID", cloudFileID.Hex()))

	request := &filedto.CompleteFileUploadRequest{
		ActualFileSizeInBytes:      fileSize,
		ActualThumbnailSizeInBytes: thumbnailSize,
		UploadConfirmed:            true,
		ThumbnailUploadConfirmed:   thumbnailSize > 0,
	}

	response, err := s.fileDTORepo.CompleteFileUploadInCloud(ctx, cloudFileID, request)
	if err != nil {
		return errors.NewAppError("failed to complete file upload", err)
	}

	if !response.Success {
		return errors.NewAppError(fmt.Sprintf("cloud rejected upload completion: %s", response.Message), nil)
	}

	if !response.UploadVerified {
		return errors.NewAppError("cloud could not verify file upload", nil)
	}

	s.logger.Info("Successfully completed file upload",
		zap.String("cloudFileID", cloudFileID.Hex()),
		zap.Int64("fileSize", response.ActualFileSize))

	return nil
}

func (s *uploadService) updateLocalFile(ctx context.Context, file *dom_file.File, cloudFileID primitive.ObjectID) error {
	updateInput := uc_file.UpdateFileInput{
		ID: file.ID,
	}

	// Update sync status
	newStatus := dom_file.SyncStatusSynced
	updateInput.SyncStatus = &newStatus

	// Update cloud ID
	file.ID = cloudFileID

	_, err := s.updateFileUseCase.Execute(ctx, updateInput)
	return err
}

func (s *uploadService) failedResult(fileID primitive.ObjectID, err error) (*fileupload.FileUploadResult, error) {
	return &fileupload.FileUploadResult{
		FileID:  fileID,
		Success: false,
		Error:   err,
	}, err
}
