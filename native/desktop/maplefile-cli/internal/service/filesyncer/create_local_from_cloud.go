// internal/service/filesyncer/create_local_from_cloud.go
package filesyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	svc_filecrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filecrypto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// CreateLocalFileFromCloudFileService defines the interface for creating a local file from a cloud file
type CreateLocalFileFromCloudFileService interface {
	Execute(ctx context.Context, cloudID primitive.ObjectID, password string) (*dom_file.File, error)
}

// createLocalFileFromCloudFileService implements the CreateLocalFileFromCloudFileService interface
type createLocalFileFromCloudFileService struct {
	logger                      *zap.Logger
	cloudRepository             filedto.FileDTORepository
	getUserByIsLoggedInUseCase  uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase        uc_collection.GetCollectionUseCase
	createFileUseCase           uc_file.CreateFileUseCase
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService
	fileDecryptionService       svc_filecrypto.FileDecryptionService
}

// NewCreateLocalFileFromCloudFileService creates a new use case for creating local files from cloud
func NewCreateLocalFileFromCloudFileService(
	logger *zap.Logger,
	cloudRepository filedto.FileDTORepository,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	createFileUseCase uc_file.CreateFileUseCase,
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
	fileDecryptionService svc_filecrypto.FileDecryptionService,
) CreateLocalFileFromCloudFileService {
	logger = logger.Named("CreateLocalFileFromCloudFileService")
	return &createLocalFileFromCloudFileService{
		logger:                      logger,
		cloudRepository:             cloudRepository,
		getUserByIsLoggedInUseCase:  getUserByIsLoggedInUseCase,
		getCollectionUseCase:        getCollectionUseCase,
		createFileUseCase:           createFileUseCase,
		collectionDecryptionService: collectionDecryptionService,
		fileDecryptionService:       fileDecryptionService,
	}
}

// Execute creates a new local file from cloud file data
func (s *createLocalFileFromCloudFileService) Execute(ctx context.Context, cloudFileID primitive.ObjectID, password string) (*dom_file.File, error) {
	//
	// STEP 1: Validate the input
	//
	e := make(map[string]string)

	if cloudFileID.IsZero() {
		e["cloudFileID"] = "Cloud file ID is required"
	}
	if password == "" {
		e["password"] = "Password is required"
	}

	// If any errors were found, return bad request error
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get the file from cloud
	//
	cloudFileDTO, err := s.cloudRepository.DownloadByIDFromCloud(ctx, cloudFileID)
	if err != nil {
		return nil, errors.NewAppError("failed to get file from the cloud", err)
	}
	if cloudFileDTO == nil {
		err := errors.NewAppError("cloud file not found", nil)
		s.logger.Error("❌ Failed to fetch file from cloud",
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 3: Validate cloud file state
	//
	if cloudFileDTO.State == "deleted" {
		s.logger.Debug("⏭️ Skipping local file creation from the cloud because it has been deleted",
			zap.String("id", cloudFileDTO.ID.Hex()))
		return nil, nil
	}

	//
	// STEP 4: Map from cloud to local and decrypt the data.
	//
	newFile := mapFileDTOToDomain(cloudFileDTO)

	// Note: We're creating a file record without the actual file content
	// The content will be downloaded separately when needed (onload operation)
	newFile.SyncStatus = dom_file.SyncStatusCloudOnly

	//
	// Step 5: Get user and collection for E2EE key chain
	//

	user, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get logged in user", err)
	}
	if user == nil {
		return nil, errors.NewAppError("user not found", nil)
	}
	collection, err := s.getCollectionUseCase.Execute(ctx, newFile.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection", err)
	}
	if collection == nil {
		return nil, errors.NewAppError("collection not found", nil)
	}

	//
	// Step 6: Decrypt the E2EE key chain to get collection key
	//
	collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, user, collection, password)
	if err != nil {
		s.logger.Error("failed to decrypt collection key chain", zap.Error(err))
		return nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}
	defer crypto.ClearBytes(collectionKey)

	//
	// Step 7: Decrypt the file key using collection key
	//
	newFileKey, err := s.fileDecryptionService.DecryptFileKey(ctx, newFile.EncryptedFileKey, collectionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file key", err)
	}
	defer crypto.ClearBytes(newFileKey)

	//
	// Step 8: Decrypt file metadata
	//
	decryptedMetadata, err := s.fileDecryptionService.DecryptFileMetadata(ctx, newFile.EncryptedMetadata, newFileKey)
	if err != nil {
		s.logger.Error("failed to decrypt file metadata", zap.Error(err))
		return nil, errors.NewAppError("failed to decrypt file metadata", err)
	}

	// Save our decrypted data.
	newFile.Name = decryptedMetadata.Name
	newFile.MimeType = decryptedMetadata.MimeType
	newFile.Metadata = decryptedMetadata
	newFile.EncryptedFilePath = decryptedMetadata.EncryptedFilePath
	newFile.EncryptedFileSize = decryptedMetadata.EncryptedFileSize
	newFile.FilePath = decryptedMetadata.DecryptedFilePath
	newFile.FileSize = decryptedMetadata.DecryptedFileSize
	newFile.EncryptedThumbnailPath = decryptedMetadata.EncryptedThumbnailPath // Developer Note: Future feature
	newFile.EncryptedThumbnailSize = decryptedMetadata.EncryptedThumbnailSize // Developer Note: Future feature
	newFile.ThumbnailPath = decryptedMetadata.DecryptedThumbnailPath          // Developer Note: Future feature
	newFile.ThumbnailSize = decryptedMetadata.DecryptedThumbnailSize          // Developer Note: Future feature
	newFile.StorageMode = dom_file.StorageModeHybrid

	//
	// STEP 9: Create local (metadata-only) file from cloud data
	//

	// Execute the use case to create the local file record
	if err := s.createFileUseCase.Execute(ctx, newFile); err != nil {
		s.logger.Error("❌ Failed to create new (local) file from the cloud",
			zap.String("id", cloudFileDTO.ID.Hex()),
			zap.Error(err))
		return nil, err
	}

	s.logger.Debug("✅ Successfully created local file from cloud",
		zap.String("id", newFile.ID.Hex()),
		zap.String("state", newFile.State))

	return newFile, nil
}
