// internal/service/filesyncer/create_local_from_cloud.go
package filesyncer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
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
	logger                     *zap.Logger
	cloudRepository            filedto.FileDTORepository
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase       uc_collection.GetCollectionUseCase
	createFileUseCase          uc_file.CreateFileUseCase
}

// NewCreateLocalFileFromCloudFileService creates a new use case for creating local files from cloud
func NewCreateLocalFileFromCloudFileService(
	logger *zap.Logger,
	cloudRepository filedto.FileDTORepository,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	createFileUseCase uc_file.CreateFileUseCase,
) CreateLocalFileFromCloudFileService {
	logger = logger.Named("CreateLocalFileFromCloudFileService")
	return &createLocalFileFromCloudFileService{
		logger:                     logger,
		cloudRepository:            cloudRepository,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
		getCollectionUseCase:       getCollectionUseCase,
		createFileUseCase:          createFileUseCase,
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
	// Step 6: Decrypt the E2EE key chain
	//
	collectionKey, err := s.decryptCollectionKeyChain(user, collection, password)
	if err != nil {
		s.logger.Error("failed to decrypt collection key chain", zap.Error(err))
		return nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}
	defer crypto.ClearBytes(collectionKey)

	// Decrypt the file key using collection key
	newFileKey, err := crypto.DecryptWithSecretBox(
		newFile.EncryptedFileKey.Ciphertext,
		newFile.EncryptedFileKey.Nonce,
		collectionKey,
	)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file key", err)
	}
	defer crypto.ClearBytes(newFileKey)

	//
	// Step 7: Decrypt file metadata
	//
	decryptedMetadata, err := s.decryptFileMetadata(newFile.EncryptedMetadata, newFileKey)
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
	// STEP X Create local (metadata-only) file from cloud data
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

// mapFileDTOToDomain maps a FileDTO to a File domain object
func mapFileDTOToDomain(dto *filedto.FileDTO) *dom_file.File {
	if dto == nil {
		return nil
	}

	state := dto.State
	if state == "" {
		state = dom_file.FileStateActive // Default to active
	}

	return &dom_file.File{
		ID:                     dto.ID,
		CollectionID:           dto.CollectionID,
		OwnerID:                dto.OwnerID,
		EncryptedMetadata:      dto.EncryptedMetadata,
		EncryptedFileKey:       dto.EncryptedFileKey,
		EncryptionVersion:      dto.EncryptionVersion,
		EncryptedHash:          dto.EncryptedHash,
		EncryptedFileSize:      dto.EncryptedFileSizeInBytes,
		EncryptedThumbnailSize: dto.EncryptedThumbnailSizeInBytes,
		Name:                   "[Encrypted]", // Placeholder until decryption
		EncryptedFilePath:      "TODO",
		Metadata:               nil,                        //TODO from EncryptedMetadata
		MimeType:               "application/octet-stream", // Default
		FilePath:               "TODO",
		FileSize:               0, //TODO
		CreatedAt:              dto.CreatedAt,
		CreatedByUserID:        dto.CreatedByUserID,
		ModifiedAt:             dto.ModifiedAt,
		ModifiedByUserID:       dto.ModifiedByUserID,
		Version:                dto.Version,
		State:                  state,
		SyncStatus:             dom_file.SyncStatusCloudOnly,
		StorageMode:            dom_file.StorageModeEncryptedOnly,
	}
}

// DownloadResult represents the result of a file download with decryption (COPIED FROM `internal/service/filedownload/download.go`)
type DownloadResult struct {
	FileID            primitive.ObjectID     `json:"file_id"`
	DecryptedData     []byte                 `json:"decrypted_data"`
	DecryptedMetadata *dom_file.FileMetadata `json:"decrypted_metadata"`
	ThumbnailData     []byte                 `json:"thumbnail_data,omitempty"`
	OriginalSize      int64                  `json:"original_size"`
	ThumbnailSize     int64                  `json:"thumbnail_size"`
}

// decryptCollectionKeyChain decrypts the complete E2EE chain to get the collection key (COPIED FROM `internal/service/filedownload/download.go`)
func (s *createLocalFileFromCloudFileService) decryptCollectionKeyChain(user *dom_user.User, collection *dom_collection.Collection, password string) ([]byte, error) {
	s.logger.Debug("🔑 Starting E2EE key chain decryption",
		zap.String("userID", user.ID.Hex()),
		zap.String("collectionID", collection.ID.Hex()))

	// STEP 1: Derive keyEncryptionKey from password
	s.logger.Debug("🧠 Step 1: Deriving key encryption key from password")
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		s.logger.Error("❌ Failed to derive key encryption key", zap.Error(err))
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ClearBytes(keyEncryptionKey)
	s.logger.Debug("✅ Successfully derived key encryption key")

	// STEP 2: Decrypt masterKey with keyEncryptionKey (ChaCha20-Poly1305)
	s.logger.Debug("🧠 Step 2: Decrypting master key with key encryption key")
	if len(user.EncryptedMasterKey.Ciphertext) == 0 || len(user.EncryptedMasterKey.Nonce) == 0 {
		s.logger.Error("❌ User encrypted master key is empty or invalid",
			zap.Int("ciphertextLen", len(user.EncryptedMasterKey.Ciphertext)),
			zap.Int("nonceLen", len(user.EncryptedMasterKey.Nonce)))
		return nil, fmt.Errorf("user encrypted master key is invalid")
	}

	masterKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		s.logger.Error("❌ Failed to decrypt master key - this usually means incorrect password",
			zap.Error(err),
			zap.String("userID", user.ID.Hex()))
		return nil, fmt.Errorf("failed to decrypt master key - incorrect password?: %w", err)
	}
	defer crypto.ClearBytes(masterKey)
	s.logger.Debug("✅ Successfully decrypted master key")

	// STEP 3: Decrypt collectionKey with masterKey (ChaCha20-Poly1305)
	s.logger.Debug("🧠 Step 3: Decrypting collection key with master key")
	if collection.EncryptedCollectionKey == nil {
		s.logger.Error("❌ Collection has no encrypted key", zap.String("collectionID", collection.ID.Hex()))
		return nil, errors.NewAppError("collection has no encrypted key", nil)
	}

	if len(collection.EncryptedCollectionKey.Ciphertext) == 0 || len(collection.EncryptedCollectionKey.Nonce) == 0 {
		s.logger.Error("❌ Collection encrypted key is empty or invalid",
			zap.Int("ciphertextLen", len(collection.EncryptedCollectionKey.Ciphertext)),
			zap.Int("nonceLen", len(collection.EncryptedCollectionKey.Nonce)))
		return nil, fmt.Errorf("collection encrypted key is invalid")
	}

	collectionKey, err := crypto.DecryptWithSecretBox(
		collection.EncryptedCollectionKey.Ciphertext,
		collection.EncryptedCollectionKey.Nonce,
		masterKey,
	)
	if err != nil {
		s.logger.Error("❌ Failed to decrypt collection key", zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt collection key: %w", err)
	}
	s.logger.Debug("✅ Successfully decrypted collection key")

	return collectionKey, nil
}

// decryptFileMetadata decrypts the encrypted file metadata using ChaCha20-Poly1305 (COPIED FROM `internal/service/filedownload/download.go`)
func (s *createLocalFileFromCloudFileService) decryptFileMetadata(encryptedMetadata string, fileKey []byte) (*dom_file.FileMetadata, error) {
	s.logger.Debug("🔑 Decrypting file metadata")

	// The encrypted metadata is stored as base64 encoded (nonce + ciphertext)
	// Format: base64(12-byte-nonce + ciphertext) for ChaCha20-Poly1305
	combined, err := base64.StdEncoding.DecodeString(encryptedMetadata)
	if err != nil {
		s.logger.Error("❌ Failed to decode encrypted metadata from base64", zap.Error(err))
		return nil, fmt.Errorf("failed to decode encrypted metadata: %w", err)
	}

	// Split nonce and ciphertext for ChaCha20-Poly1305 (12-byte nonce)
	if len(combined) < crypto.ChaCha20Poly1305NonceSize {
		s.logger.Error("❌ Combined data too short",
			zap.Int("expectedMinSize", crypto.ChaCha20Poly1305NonceSize),
			zap.Int("actualSize", len(combined)))
		return nil, fmt.Errorf("combined data too short: expected at least %d bytes for ChaCha20-Poly1305, got %d", crypto.ChaCha20Poly1305NonceSize, len(combined))
	}

	nonce := make([]byte, crypto.ChaCha20Poly1305NonceSize)
	copy(nonce, combined[:crypto.ChaCha20Poly1305NonceSize])

	ciphertext := make([]byte, len(combined)-crypto.ChaCha20Poly1305NonceSize)
	copy(ciphertext, combined[crypto.ChaCha20Poly1305NonceSize:])

	// Decrypt metadata using ChaCha20-Poly1305
	decryptedBytes, err := crypto.DecryptWithSecretBox(ciphertext, nonce, fileKey)
	if err != nil {
		s.logger.Error("❌ Failed to decrypt metadata", zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt metadata: %w", err)
	}

	// Parse JSON metadata
	var metadata dom_file.FileMetadata
	if err := json.Unmarshal(decryptedBytes, &metadata); err != nil {
		s.logger.Error("❌ Failed to parse decrypted metadata JSON", zap.Error(err))
		return nil, fmt.Errorf("failed to parse decrypted metadata: %w", err)
	}

	s.logger.Debug("✅ Successfully decrypted file metadata",
		zap.String("fileName", metadata.Name),
		zap.String("mimeType", metadata.MimeType),
		zap.Int64("size", metadata.Size),
		zap.Int64("created", metadata.Created),
		zap.String("encryptedFilePath", metadata.EncryptedFilePath),
		zap.Int64("EncryptedFileSize", metadata.EncryptedFileSize),
		zap.String("decryptedFilePath", metadata.DecryptedFilePath),
		zap.Int64("decryptedFileSize", metadata.DecryptedFileSize),
		zap.String("encryptedThumbnailPath", metadata.EncryptedThumbnailPath),
		zap.Int64("encryptedThumbnailSize", metadata.EncryptedThumbnailSize),
		zap.String("decryptedThumbnailPath", metadata.DecryptedThumbnailPath),
		zap.Int64("decryptedThumbnailSize", metadata.DecryptedThumbnailSize))
	return &metadata, nil
}
