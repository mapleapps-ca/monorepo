// cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/create.go
package filemetadata

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

func (impl *fileMetadataRepositoryImpl) Create(file *dom_file.File) error {
	if file == nil {
		return fmt.Errorf("file cannot be nil")
	}

	if !impl.isValidUUID(file.ID) {
		return fmt.Errorf("file ID is required")
	}

	if !impl.isValidUUID(file.CollectionID) {
		return fmt.Errorf("collection ID is required")
	}

	if !impl.isValidUUID(file.OwnerID) {
		return fmt.Errorf("owner ID is required")
	}

	// Set creation timestamp if not set
	if file.CreatedAt.IsZero() {
		file.CreatedAt = time.Now()
	}

	if file.ModifiedAt.IsZero() {
		file.ModifiedAt = file.CreatedAt
	}

	// Ensure state is set
	if file.State == "" {
		file.State = dom_file.FileStateActive
	}

	// Serialize encrypted file key
	encryptedKeyJSON, err := impl.serializeEncryptedFileKey(file.EncryptedFileKey)
	if err != nil {
		return fmt.Errorf("failed to serialize encrypted file key: %w", err)
	}

	batch := impl.Session.NewBatch(gocql.LoggedBatch)

	// 1. Insert into main table
	batch.Query(`INSERT INTO mapleapps.maplefile_files_by_id
		(id, collection_id, owner_id, encrypted_metadata, encrypted_file_key, encryption_version,
		 encrypted_hash, encrypted_file_object_key, encrypted_file_size_in_bytes,
		 encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
		 created_at, created_by_user_id, modified_at, modified_by_user_id, version,
		 state, tombstone_version, tombstone_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		file.ID, file.CollectionID, file.OwnerID, file.EncryptedMetadata, encryptedKeyJSON,
		file.EncryptionVersion, file.EncryptedHash, file.EncryptedFileObjectKey,
		file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
		file.EncryptedThumbnailSizeInBytes, file.CreatedAt, file.CreatedByUserID,
		file.ModifiedAt, file.ModifiedByUserID, file.Version, file.State,
		file.TombstoneVersion, file.TombstoneExpiry)

	// 2. Insert into collection table
	batch.Query(`INSERT INTO mapleapps.maplefile_files_by_collection_id_with_desc_modified_at_and_asc_file_id
		(collection_id, modified_at, file_id, owner_id, encrypted_metadata, encrypted_file_key,
		 encryption_version, encrypted_hash, encrypted_file_object_key, encrypted_file_size_in_bytes,
		 encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
		 created_at, created_by_user_id, modified_by_user_id, version,
		 state, tombstone_version, tombstone_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		file.CollectionID, file.ModifiedAt, file.ID, file.OwnerID, file.EncryptedMetadata,
		encryptedKeyJSON, file.EncryptionVersion, file.EncryptedHash, file.EncryptedFileObjectKey,
		file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
		file.EncryptedThumbnailSizeInBytes, file.CreatedAt, file.CreatedByUserID,
		file.ModifiedByUserID, file.Version, file.State, file.TombstoneVersion, file.TombstoneExpiry)

	// 3. Insert into owner table
	batch.Query(`INSERT INTO mapleapps.maplefile_files_by_owner_id_with_desc_modified_at_and_asc_file_id
		(owner_id, modified_at, file_id, collection_id, encrypted_metadata, encrypted_file_key,
		 encryption_version, encrypted_hash, encrypted_file_object_key, encrypted_file_size_in_bytes,
		 encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
		 created_at, created_by_user_id, modified_by_user_id, version,
		 state, tombstone_version, tombstone_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		file.OwnerID, file.ModifiedAt, file.ID, file.CollectionID, file.EncryptedMetadata,
		encryptedKeyJSON, file.EncryptionVersion, file.EncryptedHash, file.EncryptedFileObjectKey,
		file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
		file.EncryptedThumbnailSizeInBytes, file.CreatedAt, file.CreatedByUserID,
		file.ModifiedByUserID, file.Version, file.State, file.TombstoneVersion, file.TombstoneExpiry)

	// 4. Insert into created_by table
	batch.Query(`INSERT INTO mapleapps.maplefile_files_by_created_by_user_id_with_desc_created_at_and_asc_file_id
		(created_by_user_id, created_at, file_id, collection_id, owner_id, encrypted_metadata,
		 encrypted_file_key, encryption_version, encrypted_hash, encrypted_file_object_key,
		 encrypted_file_size_in_bytes, encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
		 modified_at, modified_by_user_id, version, state, tombstone_version, tombstone_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		file.CreatedByUserID, file.CreatedAt, file.ID, file.CollectionID, file.OwnerID,
		file.EncryptedMetadata, encryptedKeyJSON, file.EncryptionVersion, file.EncryptedHash,
		file.EncryptedFileObjectKey, file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
		file.EncryptedThumbnailSizeInBytes, file.ModifiedAt, file.ModifiedByUserID, file.Version,
		file.State, file.TombstoneVersion, file.TombstoneExpiry)

	// 5. Insert into user sync table (for owner and any collection members)
	batch.Query(`INSERT INTO mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
		(user_id, modified_at, file_id, collection_id, owner_id, encrypted_metadata,
		 encrypted_file_key, encryption_version, encrypted_hash, encrypted_file_object_key,
		 encrypted_file_size_in_bytes, encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
		 created_at, created_by_user_id, modified_by_user_id, version,
		 state, tombstone_version, tombstone_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		file.OwnerID, file.ModifiedAt, file.ID, file.CollectionID, file.OwnerID,
		file.EncryptedMetadata, encryptedKeyJSON, file.EncryptionVersion, file.EncryptedHash,
		file.EncryptedFileObjectKey, file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
		file.EncryptedThumbnailSizeInBytes, file.CreatedAt, file.CreatedByUserID,
		file.ModifiedByUserID, file.Version, file.State, file.TombstoneVersion, file.TombstoneExpiry)

	// Execute batch
	if err := impl.Session.ExecuteBatch(batch); err != nil {
		impl.Logger.Error("failed to create file",
			zap.String("file_id", file.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to create file: %w", err)
	}

	impl.Logger.Info("file created successfully",
		zap.String("file_id", file.ID.String()),
		zap.String("collection_id", file.CollectionID.String()))

	return nil
}

func (impl *fileMetadataRepositoryImpl) CreateMany(files []*dom_file.File) error {
	if len(files) == 0 {
		return nil
	}

	batch := impl.Session.NewBatch(gocql.LoggedBatch)

	for _, file := range files {
		if file == nil {
			continue
		}

		// Set timestamps if not set
		if file.CreatedAt.IsZero() {
			file.CreatedAt = time.Now()
		}
		if file.ModifiedAt.IsZero() {
			file.ModifiedAt = file.CreatedAt
		}
		if file.State == "" {
			file.State = dom_file.FileStateActive
		}

		encryptedKeyJSON, err := impl.serializeEncryptedFileKey(file.EncryptedFileKey)
		if err != nil {
			return fmt.Errorf("failed to serialize encrypted file key for file %s: %w", file.ID.String(), err)
		}

		// Add to all 5 tables (same as Create but in batch)
		batch.Query(`INSERT INTO mapleapps.maplefile_files_by_id
			(id, collection_id, owner_id, encrypted_metadata, encrypted_file_key, encryption_version,
			 encrypted_hash, encrypted_file_object_key, encrypted_file_size_in_bytes,
			 encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
			 created_at, created_by_user_id, modified_at, modified_by_user_id, version,
			 state, tombstone_version, tombstone_expiry)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			file.ID, file.CollectionID, file.OwnerID, file.EncryptedMetadata, encryptedKeyJSON,
			file.EncryptionVersion, file.EncryptedHash, file.EncryptedFileObjectKey,
			file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
			file.EncryptedThumbnailSizeInBytes, file.CreatedAt, file.CreatedByUserID,
			file.ModifiedAt, file.ModifiedByUserID, file.Version, file.State,
			file.TombstoneVersion, file.TombstoneExpiry)

		// 2. Insert into collection table
		batch.Query(`INSERT INTO mapleapps.maplefile_files_by_collection_id_with_desc_modified_at_and_asc_file_id
			(collection_id, modified_at, file_id, owner_id, encrypted_metadata, encrypted_file_key,
			 encryption_version, encrypted_hash, encrypted_file_object_key, encrypted_file_size_in_bytes,
			 encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
			 created_at, created_by_user_id, modified_by_user_id, version,
			 state, tombstone_version, tombstone_expiry)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			file.CollectionID, file.ModifiedAt, file.ID, file.OwnerID, file.EncryptedMetadata,
			encryptedKeyJSON, file.EncryptionVersion, file.EncryptedHash, file.EncryptedFileObjectKey,
			file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
			file.EncryptedThumbnailSizeInBytes, file.CreatedAt, file.CreatedByUserID,
			file.ModifiedByUserID, file.Version, file.State, file.TombstoneVersion, file.TombstoneExpiry)

		// 3. Insert into owner table
		batch.Query(`INSERT INTO mapleapps.maplefile_files_by_owner_id_with_desc_modified_at_and_asc_file_id
			(owner_id, modified_at, file_id, collection_id, encrypted_metadata, encrypted_file_key,
			 encryption_version, encrypted_hash, encrypted_file_object_key, encrypted_file_size_in_bytes,
			 encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
			 created_at, created_by_user_id, modified_by_user_id, version,
			 state, tombstone_version, tombstone_expiry)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			file.OwnerID, file.ModifiedAt, file.ID, file.CollectionID, file.EncryptedMetadata,
			encryptedKeyJSON, file.EncryptionVersion, file.EncryptedHash, file.EncryptedFileObjectKey,
			file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
			file.EncryptedThumbnailSizeInBytes, file.CreatedAt, file.CreatedByUserID,
			file.ModifiedByUserID, file.Version, file.State, file.TombstoneVersion, file.TombstoneExpiry)

		// 4. Insert into created_by table
		batch.Query(`INSERT INTO mapleapps.maplefile_files_by_created_by_user_id_with_desc_created_at_and_asc_file_id
			(created_by_user_id, created_at, file_id, collection_id, owner_id, encrypted_metadata,
			 encrypted_file_key, encryption_version, encrypted_hash, encrypted_file_object_key,
			 encrypted_file_size_in_bytes, encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
			 modified_at, modified_by_user_id, version, state, tombstone_version, tombstone_expiry)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			file.CreatedByUserID, file.CreatedAt, file.ID, file.CollectionID, file.OwnerID,
			file.EncryptedMetadata, encryptedKeyJSON, file.EncryptionVersion, file.EncryptedHash,
			file.EncryptedFileObjectKey, file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
			file.EncryptedThumbnailSizeInBytes, file.ModifiedAt, file.ModifiedByUserID, file.Version,
			file.State, file.TombstoneVersion, file.TombstoneExpiry)

		// 5. Insert into user sync table (for owner and any collection members)
		batch.Query(`INSERT INTO mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
			(user_id, modified_at, file_id, collection_id, owner_id, encrypted_metadata,
			 encrypted_file_key, encryption_version, encrypted_hash, encrypted_file_object_key,
			 encrypted_file_size_in_bytes, encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
			 created_at, created_by_user_id, modified_by_user_id, version,
			 state, tombstone_version, tombstone_expiry)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			file.OwnerID, file.ModifiedAt, file.ID, file.CollectionID, file.OwnerID,
			file.EncryptedMetadata, encryptedKeyJSON, file.EncryptionVersion, file.EncryptedHash,
			file.EncryptedFileObjectKey, file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
			file.EncryptedThumbnailSizeInBytes, file.CreatedAt, file.CreatedByUserID,
			file.ModifiedByUserID, file.Version, file.State, file.TombstoneVersion, file.TombstoneExpiry)
	}

	if err := impl.Session.ExecuteBatch(batch); err != nil {
		impl.Logger.Error("failed to create multiple files", zap.Error(err))
		return fmt.Errorf("failed to create multiple files: %w", err)
	}

	impl.Logger.Info("multiple files created successfully", zap.Int("count", len(files)))
	return nil
}
