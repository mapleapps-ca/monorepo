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
	UploadFile(ctx context.Context, fileID primitive.ObjectID, userPassword string) (*fileupload.FileUploadResult, error)
}

type uploadService struct {
	logger                   *zap.Logger
	fileDTORepo              filedto.FileDTORepository
	fileRepo                 dom_file.FileRepository
	collectionRepo           dom_collection.CollectionRepository
	userRepo                 user.Repository
	getFileUseCase           uc_file.GetFileUseCase
	updateFileUseCase        uc_file.UpdateFileUseCase
	encryptFileUseCase       uc_fileupload.EncryptFileUseCase
	prepareUploadUseCase     uc_fileupload.PrepareFileUploadUseCase
	cryptoService            svc_crypto.CryptoService
	getUserByLoggedInUseCase uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase     uc_collection.GetCollectionUseCase
}

func NewUploadService(
	logger *zap.Logger,
	fileDTORepo filedto.FileDTORepository,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
	userRepo user.Repository,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	encryptFileUseCase uc_fileupload.EncryptFileUseCase,
	prepareUploadUseCase uc_fileupload.PrepareFileUploadUseCase,
	cryptoService svc_crypto.CryptoService,
	getUserByLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
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

func (s *uploadService) UploadFile(ctx context.Context, fileID primitive.ObjectID, userPassword string) (*fileupload.FileUploadResult, error) {
	// startTime := time.Now()
	s.logger.Info("Starting three-step file upload", zap.String("fileID", fileID.Hex()))

	//
	// Step 1: Validate and prepare
	//

	file, collection, collectionKey, err := s.validateAndPrepareE2EE(ctx, fileID, userPassword)
	if err != nil {
		return s.failedResult(fileID, err)
	}
	defer pkg_crypto.ClearBytes(collectionKey)

	//
	// Step 2: Create pending file in cloud
	//
	pendingResponse, err := s.createPendingFile(ctx, file, collection, collectionKey)
	if err != nil {
		return s.failedResult(fileID, err)
	}

	// Basic check to ensure cloud respects unified ID principle
	if pendingResponse.File.ID != file.ID {
		s.logger.Error("Cloud returned a different ID than expected",
			zap.String("expectedFileID", file.ID.Hex()),
			zap.String("cloudFileID", pendingResponse.File.ID.Hex()))
		// Decide on handling: treat as error, or proceed using cloud ID (violates unified principle)
		// For now, let's error out as per "unified ID" concept
		return s.failedResult(fileID, errors.NewAppError(fmt.Sprintf("cloud did not return expected file ID. Expected %s, got %s", file.ID.Hex(), pendingResponse.File.ID.Hex()), nil))
	}

	//
	// Step 3: Upload file content
	//
	fileSize, thumbnailSize, err := s.uploadEncryptedContent(ctx, file, pendingResponse)
	if err != nil {
		return s.failedResult(fileID, err)
	}

	//
	// Step 4: Complete upload
	//
	// Use the original file ID, as it's the unified ID
	if err := s.completeUpload(ctx, file.ID, fileSize, thumbnailSize); err != nil {
		return s.failedResult(fileID, err)
	}

	if err := s.updateLocalFile(ctx, file); err != nil {
		s.logger.Error("Failed to update local file after successful upload",
			zap.String("id", fileID.Hex()),
			zap.Error(err))
		// Decide if this is a fatal error or just log and continue
		// For now, log and continue as upload itself was successful
	}

	return &fileupload.FileUploadResult{
		FileID:             fileID, // This is the unified ID
		UploadedAt:         time.Now(),
		FileSizeBytes:      fileSize,
		ThumbnailSizeBytes: thumbnailSize,
		Success:            true,
	}, nil
}

func (s *uploadService) validateAndPrepareE2EE(ctx context.Context, fileID primitive.ObjectID, userPassword string) (*dom_file.File, *dom_collection.Collection, []byte, error) {
	// Confirm user's password is not empty.
	if userPassword == "" {
		return nil, nil, nil, errors.NewAppError("user password is required for E2EE operations", nil)
	}
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

	// Get logged in user
	user, err := s.getUserByLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, nil, nil, errors.NewAppError("failed to get logged in user", err)
	}
	if user == nil {
		return nil, nil, nil, errors.NewAppError("user not found", nil)
	}

	// Get collection
	collection, err := s.getCollectionUseCase.Execute(ctx, file.CollectionID)
	if err != nil {
		return nil, nil, nil, errors.NewAppError("failed to get collection", err)
	}
	if collection == nil {
		return nil, nil, nil, errors.NewAppError("collection not found", nil)
	}

	// Decrypt collection key using complete E2EE chain
	collectionKey, err := s.decryptCollectionKeyChain(user, collection, userPassword)
	if err != nil {
		return nil, nil, nil, errors.NewAppError("failed to decrypt collection key chain", err)
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

	// Prepare upload request - ensure this use case populates the request with file.ID
	request, err := s.prepareUploadUseCase.Execute(ctx, file, collection, collectionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to prepare upload request", err)
	}

	// Create pending file
	// This is expected to use the ID provided in the request (derived from file.ID)
	// and return the same ID in response.File.ID
	response, err := s.fileDTORepo.CreatePendingFileInCloud(ctx, request)
	if err != nil {
		s.logger.Error("failed to create pending file in cloud",
			zap.String("fileID", file.ID.Hex()), // Log the ID we tried to use
			zap.Any("error", err))
		return nil, errors.NewAppError("failed to create pending file in cloud", err)
	}

	if !response.Success {
		s.logger.Debug("Failed to create pending file in cloud",
			zap.String("fileID", file.ID.Hex()), // Log the ID we tried to use
			zap.String("cloud rejected file creation", response.Message))
		return nil, errors.NewAppError(fmt.Sprintf("cloud rejected file creation: %s", response.Message), nil)
	}

	s.logger.Info("Created pending file in cloud",
		zap.String("fileID", response.File.ID.Hex()), // Log the cloud's response ID (should match file.ID)
		zap.Time("urlExpiration", response.UploadURLExpirationTime))

	return response, nil
}

func (s *uploadService) completeUpload(ctx context.Context, fileID primitive.ObjectID, fileSize, thumbnailSize int64) error {
	s.logger.Debug("Completing file upload", zap.String("fileID", fileID.Hex()))

	request := &filedto.CompleteFileUploadRequest{
		ActualFileSizeInBytes:      fileSize,
		ActualThumbnailSizeInBytes: thumbnailSize,
		UploadConfirmed:            true,
		ThumbnailUploadConfirmed:   thumbnailSize > 0,
	}

	// Call complete using the fileID
	response, err := s.fileDTORepo.CompleteFileUploadInCloud(ctx, fileID, request)
	if err != nil {
		s.logger.Error("Failed to complete file upload",
			zap.String("fileID", fileID.Hex()),
			zap.Error(err))
		return errors.NewAppError("failed to complete file upload", err)
	}

	if !response.Success {
		s.logger.Error("Failed to complete file upload",
			zap.String("fileID", fileID.Hex()),
			zap.String("message", response.Message))
		return errors.NewAppError(fmt.Sprintf("cloud rejected upload completion: %s", response.Message), nil)
	}

	if !response.UploadVerified {
		s.logger.Error("Failed to complete file upload",
			zap.String("fileID", fileID.Hex()),
			zap.String("message", "cloud could not verify file upload"))
		return errors.NewAppError("cloud could not verify file upload", nil)
	}

	s.logger.Info("Successfully completed file upload",
		zap.String("fileID", fileID.Hex()),
		zap.Int64("fileSize", response.ActualFileSize))

	return nil
}

func (s *uploadService) updateLocalFile(ctx context.Context, file *dom_file.File) error {
	// Only update sync status and remove local paths if needed
	updateInput := uc_file.UpdateFileInput{
		ID: file.ID, // Use the unified ID
	}

	// Update sync status
	newStatus := dom_file.SyncStatusSynced
	updateInput.SyncStatus = &newStatus
	file.SyncStatus = newStatus // Update in-memory object as well

	// Execute the update using the use case
	if _, err := s.updateFileUseCase.Execute(ctx, updateInput); err != nil {
		s.logger.Error("Failed to update local file status and paths after successful upload",
			zap.String("id", file.ID.Hex()),
			zap.Error(err))
		// Return error, but the upload itself succeeded. This is a post-upload cleanup/status update error.
		return err
	}

	return nil // Update was successful
}

func (s *uploadService) failedResult(fileID primitive.ObjectID, err error) (*fileupload.FileUploadResult, error) {
	return &fileupload.FileUploadResult{
		FileID:  fileID, // This is the unified ID
		Success: false,
		Error:   err,
	}, err
}

// Same decryption chain as addService
func (s *uploadService) decryptCollectionKeyChain(user *user.User, collection *dom_collection.Collection, password string) ([]byte, error) {
	// STEP 1: Derive keyEncryptionKey from password
	keyEncryptionKey, err := pkg_crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer pkg_crypto.ClearBytes(keyEncryptionKey)

	// STEP 2: Decrypt masterKey with keyEncryptionKey
	masterKey, err := pkg_crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt master key - incorrect password?: %w", err)
	}
	defer pkg_crypto.ClearBytes(masterKey)

	// STEP 3: Decrypt collectionKey with masterKey
	return pkg_crypto.DecryptWithSecretBox(
		collection.EncryptedCollectionKey.Ciphertext,
		collection.EncryptedCollectionKey.Nonce,
		masterKey,
	)
}

// Upload already encrypted content (no re-encryption needed)
func (s *uploadService) uploadEncryptedContent(ctx context.Context, file *dom_file.File, pendingResponse *filedto.CreatePendingFileResponse) (int64, int64, error) {
	s.logger.Debug("Uploading encrypted file content", zap.String("fileID", file.ID.Hex()))

	// Read already encrypted file
	encryptedData, err := os.ReadFile(file.EncryptedFilePath)
	if err != nil {
		return 0, 0, errors.NewAppError("failed to read encrypted file", err)
	}

	// Upload encrypted data directly (no re-encryption)
	if err := s.fileDTORepo.UploadFileToCloud(ctx, pendingResponse.PresignedUploadURL, encryptedData); err != nil {
		return 0, 0, errors.NewAppError("failed to upload encrypted file content", err)
	}

	// Handle thumbnail if exists
	var thumbnailSize int64
	if file.EncryptedThumbnailPath != "" && pendingResponse.PresignedThumbnailURL != "" {
		thumbnailData, err := os.ReadFile(file.EncryptedThumbnailPath)
		if err != nil {
			s.logger.Warn("Failed to read encrypted thumbnail",
				zap.String("fileID", file.ID.Hex()),
				zap.Error(err))
		} else {
			if err := s.fileDTORepo.UploadThumbnailToCloud(ctx, pendingResponse.PresignedThumbnailURL, thumbnailData); err != nil {
				s.logger.Warn("Failed to upload encrypted thumbnail",
					zap.String("fileID", file.ID.Hex()),
					zap.Error(err))
			} else {
				thumbnailSize = int64(len(thumbnailData))
			}
		}
	}

	return int64(len(encryptedData)), thumbnailSize, nil
}
