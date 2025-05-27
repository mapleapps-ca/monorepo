// internal/usecase/sync/sync.go
package sync

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/sync"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
)

// SyncUseCase defines the interface for sync operations
type SyncUseCase interface {
	SyncCollections(ctx context.Context) (*sync.SyncResult, error)
	SyncFiles(ctx context.Context) (*sync.SyncResult, error)
	FullSync(ctx context.Context) (*sync.SyncResult, error)
	ResetSync(ctx context.Context) error
}

// syncUseCase implements the SyncUseCase interface
type syncUseCase struct {
	logger                  *zap.Logger
	syncRepository          sync.SyncRepository
	syncStateRepository     sync.SyncStateRepository
	collectionRepository    collection.CollectionRepository
	collectionDTORepository collectiondto.CollectionDTORepository
	fileRepository          file.FileRepository
	fileDTORepository       filedto.FileDTORepository
	transactionManager      transaction.Manager
}

// NewSyncUseCase creates a new use case for sync operations
func NewSyncUseCase(
	logger *zap.Logger,
	syncRepository sync.SyncRepository,
	syncStateRepository sync.SyncStateRepository,
	collectionRepository collection.CollectionRepository,
	collectionDTORepository collectiondto.CollectionDTORepository,
	fileRepository file.FileRepository,
	fileDTORepository filedto.FileDTORepository,
	transactionManager transaction.Manager,
) SyncUseCase {
	return &syncUseCase{
		logger:                  logger,
		syncRepository:          syncRepository,
		syncStateRepository:     syncStateRepository,
		collectionRepository:    collectionRepository,
		collectionDTORepository: collectionDTORepository,
		fileRepository:          fileRepository,
		fileDTORepository:       fileDTORepository,
		transactionManager:      transactionManager,
	}
}

func (uc *syncUseCase) SyncCollections(ctx context.Context) (*sync.SyncResult, error) {
	uc.logger.Info("Starting collection sync")

	result := &sync.SyncResult{}

	// Get current sync state
	syncState, err := uc.syncStateRepository.GetSyncState(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get sync state", err)
	}

	// Prepare cursor for pagination
	var cursor *sync.SyncCursor
	if !syncState.LastCollectionSync.IsZero() {
		cursor = &sync.SyncCursor{
			LastModified: syncState.LastCollectionSync,
			LastID:       syncState.LastCollectionID,
		}
	}

	// Begin transaction
	if err := uc.transactionManager.Begin(); err != nil {
		return nil, errors.NewAppError("failed to begin transaction", err)
	}

	var latestModified time.Time
	var latestID primitive.ObjectID

	// Paginate through all collection changes
	for {
		// Get sync data from cloud
		response, err := uc.syncRepository.GetCollectionSyncData(ctx, cursor, 1000)
		if err != nil {
			uc.transactionManager.Rollback()
			return nil, errors.NewAppError("failed to get collection sync data", err)
		}

		// Process each collection in the response
		for _, serverCollection := range response.Collections {
			result.CollectionsProcessed++

			// Track latest modification time and ID
			if serverCollection.ModifiedAt.After(latestModified) {
				latestModified = serverCollection.ModifiedAt
				latestID = serverCollection.ID
			}

			// Get local collection if it exists
			localCollection, err := uc.collectionRepository.GetByID(ctx, serverCollection.ID)
			if err != nil {
				uc.logger.Error("Failed to get local collection",
					zap.String("collection_id", serverCollection.ID.Hex()),
					zap.Error(err))
				result.Errors = append(result.Errors, "Failed to get local collection: "+serverCollection.ID.Hex())
				continue
			}

			if localCollection == nil {
				// Collection doesn't exist locally, need to fetch and create it
				if serverCollection.State == collection.CollectionStateDeleted {
					// Don't create deleted collections
					uc.logger.Debug("Skipping deleted collection that doesn't exist locally",
						zap.String("collection_id", serverCollection.ID.Hex()))
					continue
				}

				if err := uc.createLocalCollectionFromCloud(ctx, serverCollection.ID); err != nil {
					uc.logger.Error("Failed to create local collection from cloud",
						zap.String("collection_id", serverCollection.ID.Hex()),
						zap.Error(err))
					result.Errors = append(result.Errors, "Failed to create collection: "+serverCollection.ID.Hex())
					continue
				}
				result.CollectionsAdded++
			} else {
				// Collection exists locally, check if update is needed
				if localCollection.Version != serverCollection.Version {
					if serverCollection.State == collection.CollectionStateDeleted {
						// Delete local collection
						if err := uc.collectionRepository.Delete(ctx, serverCollection.ID); err != nil {
							uc.logger.Error("Failed to delete local collection",
								zap.String("collection_id", serverCollection.ID.Hex()),
								zap.Error(err))
							result.Errors = append(result.Errors, "Failed to delete collection: "+serverCollection.ID.Hex())
							continue
						}
						result.CollectionsDeleted++
					} else {
						// Update local collection from cloud
						if err := uc.updateLocalCollectionFromCloud(ctx, serverCollection.ID); err != nil {
							uc.logger.Error("Failed to update local collection from cloud",
								zap.String("collection_id", serverCollection.ID.Hex()),
								zap.Error(err))
							result.Errors = append(result.Errors, "Failed to update collection: "+serverCollection.ID.Hex())
							continue
						}
						result.CollectionsUpdated++
					}
				}
			}
		}

		// Check if there are more pages
		if !response.HasMore {
			break
		}

		// Update cursor for next page
		cursor = response.NextCursor
	}

	// Update sync state
	if !latestModified.IsZero() {
		syncState.LastCollectionSync = latestModified
		syncState.LastCollectionID = latestID
		if err := uc.syncStateRepository.SaveSyncState(ctx, syncState); err != nil {
			uc.transactionManager.Rollback()
			return nil, errors.NewAppError("failed to save sync state", err)
		}
	}

	// Commit transaction
	if err := uc.transactionManager.Commit(); err != nil {
		return nil, errors.NewAppError("failed to commit transaction", err)
	}

	uc.logger.Info("Collection sync completed",
		zap.Int("processed", result.CollectionsProcessed),
		zap.Int("added", result.CollectionsAdded),
		zap.Int("updated", result.CollectionsUpdated),
		zap.Int("deleted", result.CollectionsDeleted),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}

func (uc *syncUseCase) SyncFiles(ctx context.Context) (*sync.SyncResult, error) {
	uc.logger.Info("Starting file sync")

	result := &sync.SyncResult{}

	// Get current sync state
	syncState, err := uc.syncStateRepository.GetSyncState(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get sync state", err)
	}

	// Prepare cursor for pagination
	var cursor *sync.SyncCursor
	if !syncState.LastFileSync.IsZero() {
		cursor = &sync.SyncCursor{
			LastModified: syncState.LastFileSync,
			LastID:       syncState.LastFileID,
		}
	}

	// Begin transaction
	if err := uc.transactionManager.Begin(); err != nil {
		return nil, errors.NewAppError("failed to begin transaction", err)
	}

	var latestModified time.Time
	var latestID primitive.ObjectID

	// Paginate through all file changes
	for {
		// Get sync data from cloud
		response, err := uc.syncRepository.GetFileSyncData(ctx, cursor, 1000)
		if err != nil {
			uc.transactionManager.Rollback()
			return nil, errors.NewAppError("failed to get file sync data", err)
		}

		// Process each file in the response
		for _, serverFile := range response.Files {
			result.FilesProcessed++

			// Track latest modification time and ID
			if serverFile.ModifiedAt.After(latestModified) {
				latestModified = serverFile.ModifiedAt
				latestID = serverFile.ID
			}

			// Get local file if it exists
			localFile, err := uc.fileRepository.Get(ctx, serverFile.ID)
			if err != nil {
				uc.logger.Error("Failed to get local file",
					zap.String("file_id", serverFile.ID.Hex()),
					zap.Error(err))
				result.Errors = append(result.Errors, "Failed to get local file: "+serverFile.ID.Hex())
				continue
			}

			if localFile == nil {
				// File doesn't exist locally, need to fetch and create it
				if serverFile.State == file.FileStateDeleted {
					// Don't create deleted files
					uc.logger.Debug("Skipping deleted file that doesn't exist locally",
						zap.String("file_id", serverFile.ID.Hex()))
					continue
				}

				if err := uc.createLocalFileFromCloud(ctx, serverFile.ID, serverFile.CollectionID); err != nil {
					uc.logger.Error("Failed to create local file from cloud",
						zap.String("file_id", serverFile.ID.Hex()),
						zap.Error(err))
					result.Errors = append(result.Errors, "Failed to create file: "+serverFile.ID.Hex())
					continue
				}
				result.FilesAdded++
			} else {
				// File exists locally, check if update is needed
				if localFile.Version != serverFile.Version {
					if serverFile.State == file.FileStateDeleted {
						// Delete local file
						if err := uc.fileRepository.Delete(ctx, serverFile.ID); err != nil {
							uc.logger.Error("Failed to delete local file",
								zap.String("file_id", serverFile.ID.Hex()),
								zap.Error(err))
							result.Errors = append(result.Errors, "Failed to delete file: "+serverFile.ID.Hex())
							continue
						}
						result.FilesDeleted++
					} else {
						// Update local file from cloud
						if err := uc.updateLocalFileFromCloud(ctx, serverFile.ID, serverFile.CollectionID); err != nil {
							uc.logger.Error("Failed to update local file from cloud",
								zap.String("file_id", serverFile.ID.Hex()),
								zap.Error(err))
							result.Errors = append(result.Errors, "Failed to update file: "+serverFile.ID.Hex())
							continue
						}
						result.FilesUpdated++
					}
				}
			}
		}

		// Check if there are more pages
		if !response.HasMore {
			break
		}

		// Update cursor for next page
		cursor = response.NextCursor
	}

	// Update sync state
	if !latestModified.IsZero() {
		syncState.LastFileSync = latestModified
		syncState.LastFileID = latestID
		if err := uc.syncStateRepository.SaveSyncState(ctx, syncState); err != nil {
			uc.transactionManager.Rollback()
			return nil, errors.NewAppError("failed to save sync state", err)
		}
	}

	// Commit transaction
	if err := uc.transactionManager.Commit(); err != nil {
		return nil, errors.NewAppError("failed to commit transaction", err)
	}

	uc.logger.Info("File sync completed",
		zap.Int("processed", result.FilesProcessed),
		zap.Int("added", result.FilesAdded),
		zap.Int("updated", result.FilesUpdated),
		zap.Int("deleted", result.FilesDeleted),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}

func (uc *syncUseCase) FullSync(ctx context.Context) (*sync.SyncResult, error) {
	uc.logger.Info("Starting full sync (collections and files)")

	// Sync collections first
	collectionResult, err := uc.SyncCollections(ctx)
	if err != nil {
		return nil, err
	}

	// Then sync files
	fileResult, err := uc.SyncFiles(ctx)
	if err != nil {
		return nil, err
	}

	// Combine results
	result := &sync.SyncResult{
		CollectionsProcessed: collectionResult.CollectionsProcessed,
		CollectionsAdded:     collectionResult.CollectionsAdded,
		CollectionsUpdated:   collectionResult.CollectionsUpdated,
		CollectionsDeleted:   collectionResult.CollectionsDeleted,
		FilesProcessed:       fileResult.FilesProcessed,
		FilesAdded:           fileResult.FilesAdded,
		FilesUpdated:         fileResult.FilesUpdated,
		FilesDeleted:         fileResult.FilesDeleted,
		Errors:               append(collectionResult.Errors, fileResult.Errors...),
	}

	return result, nil
}

func (uc *syncUseCase) ResetSync(ctx context.Context) error {
	uc.logger.Info("Resetting sync state")
	return uc.syncStateRepository.ResetSyncState(ctx)
}

// Helper methods

func (uc *syncUseCase) createLocalCollectionFromCloud(ctx context.Context, collectionID primitive.ObjectID) error {
	uc.logger.Debug("Creating local collection from cloud", zap.String("collection_id", collectionID.Hex()))

	// Fetch full collection data from cloud
	collectionDTO, err := uc.collectionDTORepository.GetFromCloudByID(ctx, collectionID)
	if err != nil {
		return errors.NewAppError("failed to fetch collection from cloud", err)
	}

	// Convert DTO to domain model
	localCollection := uc.convertCollectionDTOToDomain(collectionDTO)

	// Create local collection
	if err := uc.collectionRepository.Create(ctx, localCollection); err != nil {
		return errors.NewAppError("failed to create local collection", err)
	}

	uc.logger.Debug("Successfully created local collection from cloud", zap.String("collection_id", collectionID.Hex()))
	return nil
}

func (uc *syncUseCase) updateLocalCollectionFromCloud(ctx context.Context, collectionID primitive.ObjectID) error {
	uc.logger.Debug("Updating local collection from cloud", zap.String("collection_id", collectionID.Hex()))

	// Fetch full collection data from cloud
	collectionDTO, err := uc.collectionDTORepository.GetFromCloudByID(ctx, collectionID)
	if err != nil {
		return errors.NewAppError("failed to fetch collection from cloud", err)
	}

	// Convert DTO to domain model
	updatedCollection := uc.convertCollectionDTOToDomain(collectionDTO)

	// Save updated collection
	if err := uc.collectionRepository.Save(ctx, updatedCollection); err != nil {
		return errors.NewAppError("failed to update local collection", err)
	}

	uc.logger.Debug("Successfully updated local collection from cloud", zap.String("collection_id", collectionID.Hex()))
	return nil
}

func (uc *syncUseCase) createLocalFileFromCloud(ctx context.Context, fileID primitive.ObjectID, collectionID primitive.ObjectID) error {
	uc.logger.Debug("Creating local file from cloud", zap.String("file_id", fileID.Hex()))

	// Fetch full file data from cloud
	fileDTO, err := uc.fileDTORepository.DownloadByIDFromCloud(ctx, fileID)
	if err != nil {
		return errors.NewAppError("failed to fetch file from cloud", err)
	}

	// Convert DTO to domain model
	localFile := uc.convertFileDTOToDomain(fileDTO, collectionID)

	// Create local file
	if err := uc.fileRepository.Create(ctx, localFile); err != nil {
		return errors.NewAppError("failed to create local file", err)
	}

	uc.logger.Debug("Successfully created local file from cloud", zap.String("file_id", fileID.Hex()))
	return nil
}

func (uc *syncUseCase) updateLocalFileFromCloud(ctx context.Context, fileID primitive.ObjectID, collectionID primitive.ObjectID) error {
	uc.logger.Debug("Updating local file from cloud", zap.String("file_id", fileID.Hex()))

	// Fetch full file data from cloud
	fileDTO, err := uc.fileDTORepository.DownloadByIDFromCloud(ctx, fileID)
	if err != nil {
		return errors.NewAppError("failed to fetch file from cloud", err)
	}

	// Convert DTO to domain model
	updatedFile := uc.convertFileDTOToDomain(fileDTO, collectionID)

	// Save updated file
	if err := uc.fileRepository.Update(ctx, updatedFile); err != nil {
		return errors.NewAppError("failed to update local file", err)
	}

	uc.logger.Debug("Successfully updated local file from cloud", zap.String("file_id", fileID.Hex()))
	return nil
}

// Conversion helpers

func (uc *syncUseCase) convertCollectionDTOToDomain(dto *collectiondto.CollectionDTO) *collection.Collection {
	members := make([]*collection.CollectionMembership, len(dto.Members))
	for i, memberDTO := range dto.Members {
		members[i] = &collection.CollectionMembership{
			ID:                     memberDTO.ID,
			CollectionID:           memberDTO.CollectionID,
			RecipientID:            memberDTO.RecipientID,
			RecipientEmail:         memberDTO.RecipientEmail,
			GrantedByID:            memberDTO.GrantedByID,
			EncryptedCollectionKey: memberDTO.EncryptedCollectionKey,
			PermissionLevel:        memberDTO.PermissionLevel,
			CreatedAt:              memberDTO.CreatedAt,
			IsInherited:            memberDTO.IsInherited,
			InheritedFromID:        memberDTO.InheritedFromID,
		}
	}

	return &collection.Collection{
		ID:                     dto.ID,
		OwnerID:                dto.OwnerID,
		EncryptedName:          dto.EncryptedName,
		CollectionType:         dto.CollectionType,
		EncryptedCollectionKey: dto.EncryptedCollectionKey,
		Members:                members,
		ParentID:               dto.ParentID,
		AncestorIDs:            dto.AncestorIDs,
		Children:               nil, // Children are not synced directly
		CreatedAt:              dto.CreatedAt,
		CreatedByUserID:        dto.CreatedByUserID,
		ModifiedAt:             dto.ModifiedAt,
		ModifiedByUserID:       dto.ModifiedByUserID,
		Version:                dto.Version,
		State:                  dto.State,
		SyncStatus:             collection.SyncStatusSynced, // Mark as synced since we're syncing it
	}
}

func (uc *syncUseCase) convertFileDTOToDomain(dto *filedto.FileDTO, collectionID primitive.ObjectID) *file.File {
	return &file.File{
		ID:                     dto.ID,
		CollectionID:           collectionID,
		OwnerID:                dto.OwnerID,
		EncryptedMetadata:      dto.EncryptedMetadata,
		EncryptedFileKey:       dto.EncryptedFileKey,
		EncryptionVersion:      dto.EncryptionVersion,
		EncryptedHash:          dto.EncryptedHash,
		EncryptedFilePath:      "", // Path will be set when downloaded
		EncryptedFileSize:      dto.EncryptedFileSizeInBytes,
		FilePath:               "", // Path will be set when decrypted
		FileSize:               0,  // Size will be set when decrypted
		EncryptedThumbnailPath: "", // Path will be set when downloaded
		EncryptedThumbnailSize: dto.EncryptedThumbnailSizeInBytes,
		ThumbnailPath:          "", // Path will be set when decrypted
		ThumbnailSize:          0,  // Size will be set when decrypted
		LastSyncedAt:           time.Now(),
		SyncStatus:             file.SyncStatusCloudOnly, // Mark as cloud-only since we haven't downloaded content
		StorageMode:            file.StorageModeEncryptedOnly,
		CreatedAt:              dto.CreatedAt,
		CreatedByUserID:        dto.CreatedByUserID,
		ModifiedAt:             dto.ModifiedAt,
		ModifiedByUserID:       dto.ModifiedByUserID,
		Version:                dto.Version,
		State:                  dto.State,
	}
}
