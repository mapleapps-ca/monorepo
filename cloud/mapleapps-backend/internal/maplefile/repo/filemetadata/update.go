// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/update.go
package filemetadata

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

func (impl *fileMetadataRepositoryImpl) Update(file *dom_file.File) error {
	if file == nil {
		return fmt.Errorf("file cannot be nil")
	}

	if !impl.isValidUUID(file.ID) {
		return fmt.Errorf("file ID is required")
	}

	// Get existing file to compare changes
	existing, err := impl.Get(file.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing file: %w", err)
	}

	if existing == nil {
		return fmt.Errorf("file not found")
	}

	// Update modified timestamp
	file.ModifiedAt = time.Now()

	// Serialize encrypted file key
	encryptedKeyJSON, err := impl.serializeEncryptedFileKey(file.EncryptedFileKey)
	if err != nil {
		return fmt.Errorf("failed to serialize encrypted file key: %w", err)
	}

	batch := impl.Session.NewBatch(gocql.LoggedBatch)

	// 1. Update main table
	batch.Query(`UPDATE mapleapps.maplefile_files_by_id SET
		collection_id = ?, owner_id = ?, encrypted_metadata = ?, encrypted_file_key = ?,
		encryption_version = ?, encrypted_hash = ?, encrypted_file_object_key = ?,
		encrypted_file_size_in_bytes = ?, encrypted_thumbnail_object_key = ?,
		encrypted_thumbnail_size_in_bytes = ?, created_at = ?, created_by_user_id = ?,
		modified_at = ?, modified_by_user_id = ?, version = ?, state = ?,
		tombstone_version = ?, tombstone_expiry = ?
		WHERE id = ?`,
		file.CollectionID, file.OwnerID, file.EncryptedMetadata, encryptedKeyJSON,
		file.EncryptionVersion, file.EncryptedHash, file.EncryptedFileObjectKey,
		file.EncryptedFileSizeInBytes, file.EncryptedThumbnailObjectKey,
		file.EncryptedThumbnailSizeInBytes, file.CreatedAt, file.CreatedByUserID,
		file.ModifiedAt, file.ModifiedByUserID, file.Version, file.State,
		file.TombstoneVersion, file.TombstoneExpiry, file.ID)

	// 2. Update collection table - delete old entry and insert new one
	if existing.CollectionID != file.CollectionID || existing.ModifiedAt != file.ModifiedAt {
		batch.Query(`DELETE FROM mapleapps.maplefile_files_by_collection_id_with_desc_modified_at_and_asc_file_id
			WHERE collection_id = ? AND modified_at = ? AND file_id = ?`,
			existing.CollectionID, existing.ModifiedAt, file.ID)

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
	}

	// 3. Update owner table - delete old entry and insert new one
	if existing.OwnerID != file.OwnerID || existing.ModifiedAt != file.ModifiedAt {
		batch.Query(`DELETE FROM mapleapps.maplefile_files_by_owner_id_with_desc_modified_at_and_asc_file_id
			WHERE owner_id = ? AND modified_at = ? AND file_id = ?`,
			existing.OwnerID, existing.ModifiedAt, file.ID)

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
	}

	// 4. Update created_by table - only if creator changed (rare) or created date changed
	if existing.CreatedByUserID != file.CreatedByUserID || existing.CreatedAt != file.CreatedAt {
		batch.Query(`DELETE FROM mapleapps.maplefile_files_by_created_by_user_id_with_desc_created_at_and_asc_file_id
			WHERE created_by_user_id = ? AND created_at = ? AND file_id = ?`,
			existing.CreatedByUserID, existing.CreatedAt, file.ID)

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
	}

	// 5. Update user sync table - delete old entry and insert new one for owner
	batch.Query(`DELETE FROM mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
		WHERE user_id = ? AND modified_at = ? AND file_id = ?`,
		existing.OwnerID, existing.ModifiedAt, file.ID)

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
		impl.Logger.Error("failed to update file",
			zap.String("file_id", file.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to update file: %w", err)
	}

	impl.Logger.Info("file updated successfully",
		zap.String("file_id", file.ID.String()))

	return nil
}
