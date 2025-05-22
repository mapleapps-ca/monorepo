package filesyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	dom_localfile "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	uc_localfile "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	uc_remotefile "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotefile"
)

// UploadInput represents the input for uploading a local file
type UploadInput struct {
	LocalFileID string `json:"local_file_id"`
	ForceUpdate bool   `json:"force_update,omitempty"`
}

// UploadOutput represents the result of uploading a local file
type UploadOutput struct {
	Success    bool                           `json:"success"`
	Message    string                         `json:"message"`
	Action     string                         `json:"action"` // "created" or "updated"
	LocalFile  *localfile.LocalFile           `json:"local_file"`
	RemoteFile *remotefile.RemoteFileResponse `json:"remote_file"`
}

// UploadService defines the interface for uploading local files
type UploadService interface {
	Upload(ctx context.Context, input UploadInput) (*UploadOutput, error)
}

// fileSyncerUploadService implements the UploadService interface
type fileSyncerUploadService struct {
	logger                  *zap.Logger
	localFileRepo           dom_localfile.LocalFileRepository
	getLocalFileUseCase     uc_localfile.GetLocalFileUseCase
	updateLocalFileUseCase  uc_localfile.UpdateLocalFileUseCase
	createRemoteFileUseCase uc_remotefile.CreateRemoteFileUseCase
	uploadRemoteFileUseCase uc_remotefile.UploadRemoteFileUseCase
	fetchRemoteFileUseCase  uc_remotefile.FetchRemoteFileUseCase
}

// NewFileSyncerUploadService creates a new service for uploading local files
func NewFileSyncerUploadService(
	logger *zap.Logger,
	localFileRepo dom_localfile.LocalFileRepository,
	getLocalFileUseCase uc_localfile.GetLocalFileUseCase,
	updateLocalFileUseCase uc_localfile.UpdateLocalFileUseCase,
	createRemoteFileUseCase uc_remotefile.CreateRemoteFileUseCase,
	uploadRemoteFileUseCase uc_remotefile.UploadRemoteFileUseCase,
	fetchRemoteFileUseCase uc_remotefile.FetchRemoteFileUseCase,
) UploadService {
	return &fileSyncerUploadService{
		logger:                  logger,
		localFileRepo:           localFileRepo,
		getLocalFileUseCase:     getLocalFileUseCase,
		updateLocalFileUseCase:  updateLocalFileUseCase,
		createRemoteFileUseCase: createRemoteFileUseCase,
		uploadRemoteFileUseCase: uploadRemoteFileUseCase,
		fetchRemoteFileUseCase:  fetchRemoteFileUseCase,
	}
}

// Upload uploads a local file to the remote backend
func (s *fileSyncerUploadService) Upload(ctx context.Context, input UploadInput) (*UploadOutput, error) {
	// Validate inputs
	if input.LocalFileID == "" {
		return nil, errors.NewAppError("local file ID is required", nil)
	}

	// Convert ID to ObjectID
	localFileID, err := primitive.ObjectIDFromHex(input.LocalFileID)
	if err != nil {
		return nil, errors.NewAppError("invalid local file ID format", err)
	}

	// Get the local file using existing use case
	localFile, err := s.getLocalFileUseCase.ByID(ctx, localFileID)
	if err != nil {
		return nil, errors.NewAppError("failed to get local file", err)
	}

	// Validate that the file can be uploaded (must have encrypted data)
	if err := s.validateFileForUpload(localFile); err != nil {
		return nil, err
	}

	// Load encrypted file data
	encryptedData, err := s.loadEncryptedFileData(ctx, localFile)
	if err != nil {
		return nil, errors.NewAppError("failed to load encrypted file data", err)
	}

	var remoteFileResponse *remotefile.RemoteFileResponse
	var action string

	// Determine action based on whether the local file has a `remote_id` or not.
	if localFile.RemoteID.IsZero() {
		// Create new remote file using existing use case
		remoteFileResponse, err = s.createNewRemoteFile(ctx, localFile, encryptedData)
		if err != nil {
			return nil, err
		}
		action = "created"
	} else {
		// Update existing remote file using existing use cases
		remoteFileResponse, err = s.updateExistingRemoteFile(ctx, localFile, encryptedData)
		if err != nil {
			return nil, err
		}
		action = "updated"
	}

	// Update local file sync status using existing use case
	updatedLocalFile, err := s.updateLocalFileUseCase.UpdateSyncStatus(
		ctx,
		localFile.ID,
		remoteFileResponse.ID, // A.K.A. "RemoteID" inside `localFile` domain entity.
		localfile.SyncStatusSynced,
	)
	if err != nil {
		s.logger.Error("Failed to update local file sync status after upload",
			zap.String("localFileID", localFile.ID.Hex()),
			zap.String("remoteFileID", remoteFileResponse.ID.Hex()),
			zap.Error(err))
		// Don't fail the entire operation, just log the warning
		updatedLocalFile = localFile
		updatedLocalFile.RemoteID = remoteFileResponse.ID
		updatedLocalFile.SyncStatus = localfile.SyncStatusSynced
	}

	// Prepare success message
	var message string
	switch action {
	case "created":
		message = "File uploaded successfully and created on remote backend"
	case "updated":
		message = "File uploaded successfully and updated on remote backend"
	default:
		message = "File uploaded successfully"
	}

	return &UploadOutput{
		Success:    true,
		Message:    message,
		Action:     action,
		LocalFile:  updatedLocalFile,
		RemoteFile: remoteFileResponse,
	}, nil
}

// validateFileForUpload validates that a local file can be uploaded
func (s *fileSyncerUploadService) validateFileForUpload(file *localfile.LocalFile) error {
	// Must have encrypted file path or be in encrypted_only or hybrid mode
	if file.StorageMode == localfile.StorageModeDecryptedOnly {
		return errors.NewAppError("cannot upload decrypted-only files; only encrypted files are allowed", nil)
	}

	if file.EncryptedFilePath == "" &&
		(file.StorageMode == localfile.StorageModeEncryptedOnly || file.StorageMode == localfile.StorageModeHybrid) {
		return errors.NewAppError("file has no encrypted data available for upload", nil)
	}

	// Must have valid encryption metadata
	if file.EncryptedFileKey.Ciphertext == nil || len(file.EncryptedFileKey.Ciphertext) == 0 {
		return errors.NewAppError("file has no encrypted file key", nil)
	}

	if file.EncryptionVersion == "" || file.EncryptionVersion == "unencrypted" {
		return errors.NewAppError("file is not encrypted", nil)
	}

	return nil
}

// loadEncryptedFileData loads the encrypted file data from the local filesystem
func (s *fileSyncerUploadService) loadEncryptedFileData(ctx context.Context, file *localfile.LocalFile) ([]byte, error) {
	// For files that have encrypted data, we need to load it
	if file.StorageMode == localfile.StorageModeEncryptedOnly || file.StorageMode == localfile.StorageModeHybrid {
		return s.localFileRepo.LoadFileData(ctx, file)
	}

	return nil, errors.NewAppError("no encrypted data available for upload", nil)
}

// createNewRemoteFile creates a new remote file with complete upload flow and rollback
func (s *fileSyncerUploadService) createNewRemoteFile(
	ctx context.Context,
	localFile *localfile.LocalFile,
	encryptedData []byte,
) (*remotefile.RemoteFileResponse, error) {
	s.logger.Info("Creating new remote file with data upload",
		zap.String("localFileID", localFile.ID.Hex()),
		zap.String("encryptedFileID", localFile.EncryptedFileID),
		zap.Int("dataSize", len(encryptedData)))

	// Create remote file input with complete file data
	createInput := uc_remotefile.CreateRemoteFileInput{
		CollectionID:      localFile.CollectionID,
		EncryptedFileID:   localFile.EncryptedFileID,
		EncryptedFileSize: int64(len(encryptedData)),
		EncryptedMetadata: localFile.EncryptedMetadata,
		EncryptedFileKey:  localFile.EncryptedFileKey,
		EncryptionVersion: localFile.EncryptionVersion,
		EncryptedHash:     localFile.EncryptedHash,
		FileData:          encryptedData, // This will trigger backend upload
	}

	// Execute the complete create + upload flow (with automatic rollback on failure)
	remoteFile, err := s.createRemoteFileUseCase.Execute(ctx, createInput)
	if err != nil {
		s.logger.Error("Failed to create and upload remote file",
			zap.String("localFileID", localFile.ID.Hex()),
			zap.String("encryptedFileID", localFile.EncryptedFileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to create and upload remote file", err)
	}

	s.logger.Info("Remote file created and uploaded successfully",
		zap.String("localFileID", localFile.ID.Hex()),
		zap.String("remoteFileID", remoteFile.ID.Hex()),
		zap.String("fileObjectKey", remoteFile.FileObjectKey))

	return remoteFile, nil
}

// updateExistingRemoteFile updates an existing remote file with new data
func (s *fileSyncerUploadService) updateExistingRemoteFile(
	ctx context.Context,
	localFile *localfile.LocalFile,
	encryptedData []byte,
) (*remotefile.RemoteFileResponse, error) {
	s.logger.Info("Updating existing remote file with new data",
		zap.String("localFileID", localFile.ID.Hex()),
		zap.String("remoteFileID", localFile.RemoteID.Hex()),
		zap.Int("dataSize", len(encryptedData)))

	// Check if remote file exists
	if localFile.RemoteID.IsZero() {
		return nil, errors.NewAppError("local file has no remote ID for update", nil)
	}

	// Fetch existing remote file to verify it exists
	existingRemote, err := s.fetchRemoteFileUseCase.ByID(ctx, localFile.RemoteID)
	if err != nil {
		s.logger.Error("Failed to fetch existing remote file for update",
			zap.String("localFileID", localFile.ID.Hex()),
			zap.String("remoteFileID", localFile.RemoteID.Hex()),
			zap.Error(err))
		return nil, errors.NewAppError("failed to fetch existing remote file for update", err)
	}

	// Upload the new data to the existing remote file
	if err := s.uploadRemoteFileUseCase.Execute(ctx, localFile.RemoteID, encryptedData); err != nil {
		s.logger.Error("Failed to upload updated file data",
			zap.String("localFileID", localFile.ID.Hex()),
			zap.String("remoteFileID", localFile.RemoteID.Hex()),
			zap.Error(err))
		return nil, errors.NewAppError("failed to upload updated file data to backend/S3", err)
	}

	s.logger.Info("Successfully uploaded updated file data",
		zap.String("localFileID", localFile.ID.Hex()),
		zap.String("remoteFileID", localFile.RemoteID.Hex()))

	// Return updated remote file info
	return &remotefile.RemoteFileResponse{
		ID:                 existingRemote.ID,
		CollectionID:       existingRemote.CollectionID,
		OwnerID:            existingRemote.OwnerID,
		EncryptedFileID:    existingRemote.EncryptedFileID,
		FileObjectKey:      existingRemote.FileObjectKey,
		EncryptedFileSize:  int64(len(encryptedData)),   // Updated size
		EncryptedMetadata:  localFile.EncryptedMetadata, // Use local metadata (might be updated)
		EncryptedFileKey:   existingRemote.EncryptedFileKey,
		EncryptionVersion:  existingRemote.EncryptionVersion,
		EncryptedHash:      localFile.EncryptedHash, // Use local hash (might be updated)
		ThumbnailObjectKey: existingRemote.ThumbnailObjectKey,
		CreatedAt:          existingRemote.CreatedAt,
		ModifiedAt:         existingRemote.ModifiedAt,
	}, nil
}
