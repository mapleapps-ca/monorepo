// native/desktop/maplefile-cli/internal/service/fileupload/upload.go
package fileupload

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/fileupload"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	uc_fileupload "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/fileupload"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	pkg_crypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// FileUploadService handles three-step file upload to cloud
type FileUploadService interface {
	Execute(ctx context.Context, fileID gocql.UUID, userPassword string) (*fileupload.FileUploadResult, error)
}

type fileUploadService struct {
	logger                      *zap.Logger
	fileDTORepo                 filedto.FileDTORepository
	fileRepo                    dom_file.FileRepository
	collectionRepo              dom_collection.CollectionRepository
	userRepo                    user.Repository
	getFileUseCase              uc_file.GetFileUseCase
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService
	updateFileUseCase           uc_file.UpdateFileUseCase
	prepareUploadUseCase        uc_fileupload.PrepareFileUploadUseCase
	getUserByLoggedInUseCase    uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase        uc_collection.GetCollectionUseCase
}

func NewFileUploadService(
	logger *zap.Logger,
	fileDTORepo filedto.FileDTORepository,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
	userRepo user.Repository,
	getFileUseCase uc_file.GetFileUseCase,
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
	updateFileUseCase uc_file.UpdateFileUseCase,
	prepareUploadUseCase uc_fileupload.PrepareFileUploadUseCase,
	getUserByLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
) FileUploadService {
	logger = logger.Named("FileUploadService")
	return &fileUploadService{
		logger:                      logger,
		fileDTORepo:                 fileDTORepo,
		fileRepo:                    fileRepo,
		collectionRepo:              collectionRepo,
		userRepo:                    userRepo,
		getFileUseCase:              getFileUseCase,
		collectionDecryptionService: collectionDecryptionService,
		updateFileUseCase:           updateFileUseCase,
		getCollectionUseCase:        getCollectionUseCase,
		getUserByLoggedInUseCase:    getUserByLoggedInUseCase,
		prepareUploadUseCase:        prepareUploadUseCase,
	}
}

func (s *fileUploadService) Execute(ctx context.Context, fileID gocql.UUID, userPassword string) (*fileupload.FileUploadResult, error) {
	// startTime := time.Now()
	s.logger.Info("✨ Starting three-step file upload", zap.String("fileID", fileID.String()))

	//
	// Step 1: Validate and prepare
	//

	file, collection, collectionKey, err := s.validateAndPrepareE2EE(ctx, fileID, userPassword)
	if err != nil {
		return s.failedResult(fileID, err)
	}
	// Developer Note: This needs to be done since it is purposefully not done in the `collectionEncryptionService`.
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
		s.logger.Error("❌ Cloud returned a different ID than expected",
			zap.String("expectedFileID", file.ID.String()),
			zap.String("cloudFileID", pendingResponse.File.ID.String()))
		// Decide on handling: treat as error, or proceed using cloud ID (violates unified principle)
		// For now, let's error out as per "unified ID" concept
		return s.failedResult(fileID, errors.NewAppError(fmt.Sprintf("cloud did not return expected file ID. Expected %s, got %s", file.ID.String(), pendingResponse.File.ID.String()), nil))
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
		s.logger.Error("❌ Failed to update local file after successful upload",
			zap.String("id", file.ID.String()),
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

func (s *fileUploadService) validateAndPrepareE2EE(ctx context.Context, fileID gocql.UUID, userPassword string) (*dom_file.File, *dom_collection.Collection, []byte, error) {
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
	collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, user, collection, userPassword)
	if err != nil {
		return nil, nil, nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}

	return file, collection, collectionKey, nil
}

func (s *fileUploadService) createPendingFile(
	ctx context.Context,
	file *dom_file.File,
	collection *dom_collection.Collection,
	collectionKey []byte,
) (*filedto.CreatePendingFileResponse, error) {
	s.logger.Debug("⚙️ Creating pending file in cloud", zap.String("fileID", file.ID.String()))

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
		s.logger.Error("❌ failed to create pending file in cloud",
			zap.String("fileID", file.ID.String()), // Log the ID we tried to use
			zap.Any("error", err))
		return nil, errors.NewAppError("failed to create pending file in cloud", err)
	}

	if !response.Success {
		s.logger.Debug("🐛 Failed to create pending file in cloud",
			zap.String("fileID", file.ID.String()), // Log the ID we tried to use
			zap.String("cloud rejected file creation", response.Message))
		return nil, errors.NewAppError(fmt.Sprintf("cloud rejected file creation: %s", response.Message), nil)
	}

	s.logger.Info("✅ Created pending file in cloud",
		zap.String("fileID", response.File.ID.String()), // Log the cloud's response ID (should match file.ID)
		zap.Time("urlExpiration", response.UploadURLExpirationTime))

	return response, nil
}

func (s *fileUploadService) completeUpload(ctx context.Context, fileID gocql.UUID, fileSize, thumbnailSize int64) error {
	s.logger.Debug("⚙️ Completing file upload", zap.String("fileID", fileID.String()))

	request := &filedto.CompleteFileUploadRequest{
		ActualFileSizeInBytes:      fileSize,
		ActualThumbnailSizeInBytes: thumbnailSize,
		UploadConfirmed:            true,
		ThumbnailUploadConfirmed:   thumbnailSize > 0,
	}

	// Call complete using the fileID
	response, err := s.fileDTORepo.CompleteFileUploadInCloud(ctx, fileID, request)
	if err != nil {
		s.logger.Error("❌ Failed to complete file upload",
			zap.String("fileID", fileID.String()),
			zap.Error(err))
		return errors.NewAppError("failed to complete file upload", err)
	}

	if !response.Success {
		s.logger.Error("❌ Failed to complete file upload",
			zap.String("fileID", fileID.String()),
			zap.String("message", response.Message))
		return errors.NewAppError(fmt.Sprintf("cloud rejected upload completion: %s", response.Message), nil)
	}

	if !response.UploadVerified {
		s.logger.Error("❌ Failed to complete file upload",
			zap.String("fileID", fileID.String()),
			zap.String("message", "cloud could not verify file upload"))
		return errors.NewAppError("cloud could not verify file upload", nil)
	}

	s.logger.Info("✅ Successfully completed file upload",
		zap.String("fileID", fileID.String()),
		zap.Int64("fileSize", response.ActualFileSize))

	return nil
}

func (s *fileUploadService) updateLocalFile(ctx context.Context, file *dom_file.File) error {
	// Only update sync status and remove local paths if needed
	newStatus := dom_file.SyncStatusSynced
	updateInput := uc_file.UpdateFileInput{
		ID:         file.ID,
		SyncStatus: &newStatus, // Update sync status
	}

	// Execute the update using the use case
	if _, err := s.updateFileUseCase.Execute(ctx, updateInput); err != nil {
		s.logger.Error("❌ Failed to update local file status and paths after successful upload",
			zap.String("id", file.ID.String()),
			zap.Error(err))
		// Return error, but the upload itself succeeded. This is a post-upload cleanup/status update error.
		return err
	}

	return nil // Update was successful
}

func (s *fileUploadService) failedResult(fileID gocql.UUID, err error) (*fileupload.FileUploadResult, error) {
	return &fileupload.FileUploadResult{
		FileID:  fileID, // This is the unified ID
		Success: false,
		Error:   err,
	}, err
}

// Upload already encrypted content (no re-encryption needed)
func (s *fileUploadService) uploadEncryptedContent(ctx context.Context, file *dom_file.File, pendingResponse *filedto.CreatePendingFileResponse) (int64, int64, error) {
	s.logger.Debug("⚙️ Uploading encrypted file content", zap.String("fileID", file.ID.String()))

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
			s.logger.Warn("⚠️ Failed to read encrypted thumbnail",
				zap.String("fileID", file.ID.String()),
				zap.Error(err))
		} else {
			if err := s.fileDTORepo.UploadThumbnailToCloud(ctx, pendingResponse.PresignedThumbnailURL, thumbnailData); err != nil {
				s.logger.Warn("⚠️ Failed to upload encrypted thumbnail",
					zap.String("fileID", file.ID.String()),
					zap.Error(err))
			} else {
				thumbnailSize = int64(len(thumbnailData))
			}
		}
	}

	return int64(len(encryptedData)), thumbnailSize, nil
}
