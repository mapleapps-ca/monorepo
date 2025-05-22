package filesyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
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
	localFileRepo           localfile.LocalFileRepository
	getLocalFileUseCase     uc_localfile.GetLocalFileUseCase
	updateLocalFileUseCase  uc_localfile.UpdateLocalFileUseCase
	createRemoteFileUseCase uc_remotefile.CreateRemoteFileUseCase
	uploadRemoteFileUseCase uc_remotefile.UploadRemoteFileUseCase
	fetchRemoteFileUseCase  uc_remotefile.FetchRemoteFileUseCase
}

// NewFileSyncerUploadService creates a new service for uploading local files
func NewFileSyncerUploadService(
	logger *zap.Logger,
	localFileRepo localfile.LocalFileRepository,
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

	// Determine action based on sync status
	switch localFile.SyncStatus {
	case localfile.SyncStatusLocalOnly:
		// Create new remote file using existing use case
		remoteFileResponse, err = s.createNewRemoteFile(ctx, localFile, encryptedData)
		if err != nil {
			return nil, err
		}
		action = "created"

	case localfile.SyncStatusSynced, localfile.SyncStatusModifiedLocally:
		if !input.ForceUpdate && localFile.SyncStatus == localfile.SyncStatusSynced {
			return nil, errors.NewAppError("file is already synced; use --force to re-upload", nil)
		}

		// Update existing remote file using existing use cases
		remoteFileResponse, err = s.updateExistingRemoteFile(ctx, localFile, encryptedData)
		if err != nil {
			return nil, err
		}
		action = "updated"

	default:
		return nil, errors.NewAppError("invalid sync status for upload", nil)
	}

	// Update local file sync status using existing use case
	updatedLocalFile, err := s.updateLocalFileUseCase.UpdateSyncStatus(
		ctx,
		localFile.ID,
		remoteFileResponse.ID,
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

// createNewRemoteFile creates a new remote file using existing use case
func (s *fileSyncerUploadService) createNewRemoteFile(
	ctx context.Context,
	localFile *localfile.LocalFile,
	encryptedData []byte,
) (*remotefile.RemoteFileResponse, error) {
	// Create remote file input
	createInput := uc_remotefile.CreateRemoteFileInput{
		CollectionID:          localFile.CollectionID,
		EncryptedFileID:       localFile.EncryptedFileID,
		FileSize:              int64(len(encryptedData)),
		EncryptedOriginalSize: "", // Could be encrypted with file key
		EncryptedMetadata:     localFile.EncryptedMetadata,
		EncryptedFileKey:      localFile.EncryptedFileKey,
		EncryptionVersion:     localFile.EncryptionVersion,
		EncryptedHash:         localFile.EncryptedHash,
		FileData:              encryptedData,
	}

	remoteFile, err := s.createRemoteFileUseCase.Execute(ctx, createInput)
	if err != nil {
		return nil, errors.NewAppError("failed to create remote file", err)
	}

	return remoteFile, nil
}

// updateExistingRemoteFile updates an existing remote file using existing use cases
func (s *fileSyncerUploadService) updateExistingRemoteFile(
	ctx context.Context,
	localFile *localfile.LocalFile,
	encryptedData []byte,
) (*remotefile.RemoteFileResponse, error) {
	// Check if remote file exists
	if localFile.RemoteID.IsZero() {
		return nil, errors.NewAppError("local file has no remote ID for update", nil)
	}

	// Fetch existing remote file to verify it exists using existing use case
	_, err := s.fetchRemoteFileUseCase.ByID(ctx, localFile.RemoteID)
	if err != nil {
		return nil, errors.NewAppError("failed to fetch existing remote file for update", err)
	}

	// Upload the new data to the existing remote file using existing use case
	if err := s.uploadRemoteFileUseCase.Execute(ctx, localFile.RemoteID, encryptedData); err != nil {
		return nil, errors.NewAppError("failed to upload updated file data", err)
	}

	// Fetch the updated remote file info using existing use case
	updatedRemoteFile, err := s.fetchRemoteFileUseCase.ByID(ctx, localFile.RemoteID)
	if err != nil {
		return nil, errors.NewAppError("failed to fetch updated remote file", err)
	}

	// Convert to response format
	return &remotefile.RemoteFileResponse{
		ID:                    updatedRemoteFile.ID,
		CollectionID:          updatedRemoteFile.CollectionID,
		OwnerID:               updatedRemoteFile.OwnerID,
		EncryptedFileID:       updatedRemoteFile.EncryptedFileID,
		FileObjectKey:         updatedRemoteFile.FileObjectKey,
		FileSize:              updatedRemoteFile.FileSize,
		EncryptedOriginalSize: updatedRemoteFile.EncryptedOriginalSize,
		EncryptedMetadata:     updatedRemoteFile.EncryptedMetadata,
		EncryptedFileKey:      updatedRemoteFile.EncryptedFileKey,
		EncryptionVersion:     updatedRemoteFile.EncryptionVersion,
		EncryptedHash:         updatedRemoteFile.EncryptedHash,
		ThumbnailObjectKey:    updatedRemoteFile.ThumbnailObjectKey,
		CreatedAt:             updatedRemoteFile.CreatedAt,
		ModifiedAt:            updatedRemoteFile.ModifiedAt,
	}, nil
}
