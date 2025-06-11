// cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/get.go
package filemetadata

import (
	"fmt"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

func (impl *fileMetadataRepositoryImpl) Get(id gocql.UUID) (*dom_file.File, error) {
	return impl.getFileByID(id, true) // state-aware
}

func (impl *fileMetadataRepositoryImpl) GetWithAnyState(id gocql.UUID) (*dom_file.File, error) {
	return impl.getFileByID(id, false) // state-agnostic
}

func (impl *fileMetadataRepositoryImpl) getFileByID(id gocql.UUID, stateAware bool) (*dom_file.File, error) {
	var (
		collectionID, ownerID, createdByUserID, modifiedByUserID gocql.UUID
		encryptedMetadata, encryptedKeyJSON, encryptionVersion   string
		encryptedHash, encryptedFileObjectKey                    string
		encryptedThumbnailObjectKey                              string
		encryptedFileSizeInBytes, encryptedThumbnailSizeInBytes  int64
		createdAt, modifiedAt, tombstoneExpiry                   time.Time
		version, tombstoneVersion                                uint64
		state                                                    string
	)

	query := `SELECT id, collection_id, owner_id, encrypted_metadata, encrypted_file_key,
		encryption_version, encrypted_hash, encrypted_file_object_key, encrypted_file_size_in_bytes,
		encrypted_thumbnail_object_key, encrypted_thumbnail_size_in_bytes,
		created_at, created_by_user_id, modified_at, modified_by_user_id, version,
		state, tombstone_version, tombstone_expiry
		FROM mapleapps.maplefile_files_by_id WHERE id = ?`

	err := impl.Session.Query(query, id).Scan(
		&id, &collectionID, &ownerID, &encryptedMetadata, &encryptedKeyJSON,
		&encryptionVersion, &encryptedHash, &encryptedFileObjectKey, &encryptedFileSizeInBytes,
		&encryptedThumbnailObjectKey, &encryptedThumbnailSizeInBytes,
		&createdAt, &createdByUserID, &modifiedAt, &modifiedByUserID, &version,
		&state, &tombstoneVersion, &tombstoneExpiry)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	// Apply state filtering if state-aware mode is enabled
	if stateAware && state != dom_file.FileStateActive {
		return nil, nil
	}

	// Deserialize encrypted file key
	encryptedFileKey, err := impl.deserializeEncryptedFileKey(encryptedKeyJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize encrypted file key: %w", err)
	}

	file := &dom_file.File{
		ID:                            id,
		CollectionID:                  collectionID,
		OwnerID:                       ownerID,
		EncryptedMetadata:             encryptedMetadata,
		EncryptedFileKey:              encryptedFileKey,
		EncryptionVersion:             encryptionVersion,
		EncryptedHash:                 encryptedHash,
		EncryptedFileObjectKey:        encryptedFileObjectKey,
		EncryptedFileSizeInBytes:      encryptedFileSizeInBytes,
		EncryptedThumbnailObjectKey:   encryptedThumbnailObjectKey,
		EncryptedThumbnailSizeInBytes: encryptedThumbnailSizeInBytes,
		CreatedAt:                     createdAt,
		CreatedByUserID:               createdByUserID,
		ModifiedAt:                    modifiedAt,
		ModifiedByUserID:              modifiedByUserID,
		Version:                       version,
		State:                         state,
		TombstoneVersion:              tombstoneVersion,
		TombstoneExpiry:               tombstoneExpiry,
	}

	return file, nil
}

func (impl *fileMetadataRepositoryImpl) GetByIDs(ids []gocql.UUID) ([]*dom_file.File, error) {
	if len(ids) == 0 {
		return []*dom_file.File{}, nil
	}

	// Use a buffered channel to collect results from goroutines
	resultsChan := make(chan *dom_file.File, len(ids))
	var wg sync.WaitGroup

	// Launch a goroutine for each ID lookup
	for _, id := range ids {
		wg.Add(1)
		go func(id gocql.UUID) {
			defer wg.Done()

			// Call the existing state-aware Get method
			file, err := impl.Get(id)

			if err != nil {
				impl.Logger.Warn("failed to get file by ID",
					zap.String("file_id", id.String()),
					zap.Error(err))
				// Send nil on error to indicate failure/absence for this ID
				resultsChan <- nil
				return
			}

			// Get returns nil for ErrNotFound or inactive state when stateAware is true.
			// Send the potentially nil file result to the channel.
			resultsChan <- file

		}(id) // Pass id into the closure
	}

	// Goroutine to close the channel once all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results from the channel
	var files []*dom_file.File
	for file := range resultsChan {
		// Only append non-nil files (found and active)
		if file != nil {
			files = append(files, file)
		}
	}

	// The original function logs warnings for errors but doesn't return an error
	// from GetByIDs itself. We maintain this behavior.
	return files, nil
}

func (impl *fileMetadataRepositoryImpl) GetByCollection(collectionID gocql.UUID) ([]*dom_file.File, error) {
	var fileIDs []gocql.UUID

	query := `SELECT file_id FROM mapleapps.maplefile_files_by_collection_id_with_desc_modified_at_and_asc_file_id
		WHERE collection_id = ? AND state = ?`

	iter := impl.Session.Query(query, collectionID, dom_file.FileStateActive).Iter()

	var fileID gocql.UUID
	for iter.Scan(&fileID) {
		fileIDs = append(fileIDs, fileID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get files by collection: %w", err)
	}

	return impl.loadMultipleFiles(fileIDs)
}

func (impl *fileMetadataRepositoryImpl) loadMultipleFiles(fileIDs []gocql.UUID) ([]*dom_file.File, error) {
	if len(fileIDs) == 0 {
		return []*dom_file.File{}, nil
	}

	// Use a buffered channel to collect results from goroutines
	// We expect up to len(fileIDs) results, some of which might be nil.
	resultsChan := make(chan *dom_file.File, len(fileIDs))
	var wg sync.WaitGroup

	// Launch a goroutine for each ID lookup
	for _, id := range fileIDs {
		wg.Add(1)
		go func(id gocql.UUID) {
			defer wg.Done()

			// Call the existing state-aware Get method
			// This method returns nil if the file is not found, or if it's
			// found but not in the 'active' state.
			file, err := impl.Get(id)

			if err != nil {
				// Log the error but continue processing other IDs.
				impl.Logger.Warn("failed to load file",
					zap.String("file_id", id.String()),
					zap.Error(err))
				// Send nil on error, consistent with how Get returns nil for not found/inactive.
				resultsChan <- nil
				return
			}

			// Get returns nil for ErrNotFound or inactive state when stateAware is true.
			// Send the potentially nil file result to the channel.
			resultsChan <- file

		}(id) // Pass id into the closure
	}

	// Goroutine to close the channel once all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results from the channel
	var files []*dom_file.File
	for file := range resultsChan {
		// Only append non-nil files (found and active, or found but error logged)
		if file != nil {
			files = append(files, file)
		}
	}

	// The original function logged warnings for errors but didn't return an error
	// from loadMultipleFiles itself. We maintain this behavior.
	return files, nil
}
