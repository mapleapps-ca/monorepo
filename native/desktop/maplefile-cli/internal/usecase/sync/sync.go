// native/desktop/maplefile-cli/internal/usecase/sync/sync.go
package sync

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/sync"
	dom_tx "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
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
	transactionManager      dom_tx.Manager
}

// NewSyncUseCase creates a new use case for sync operations
func NewSyncUseCase(
	logger *zap.Logger,
	syncRepository sync.SyncRepository,
	syncStateRepository sync.SyncStateRepository,
	collectionRepository collection.CollectionRepository,
	collectionDTORepository collectiondto.CollectionDTORepository,
	transactionManager dom_tx.Manager,
) SyncUseCase {
	return &syncUseCase{
		logger:                  logger,
		syncRepository:          syncRepository,
		syncStateRepository:     syncStateRepository,
		collectionRepository:    collectionRepository,
		collectionDTORepository: collectionDTORepository,
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
	var latestID string

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
				latestID = serverCollection.ID.Hex()
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
		if latestID != "" {
			syncState.LastCollectionID.UnmarshalText([]byte(latestID))
		}
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
	// Similar implementation for files
	uc.logger.Info("File sync not yet implemented")
	return &sync.SyncResult{}, nil
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
	// Fetch full collection data from cloud
	// This would require extending the existing collectionDTO repository
	// For now, return not implemented
	return errors.NewAppError("create local collection from cloud not yet implemented", nil)
}

func (uc *syncUseCase) updateLocalCollectionFromCloud(ctx context.Context, collectionID primitive.ObjectID) error {
	// Fetch full collection data from cloud and update local
	// This would require extending the existing collectionDTO repository
	// For now, return not implemented
	return errors.NewAppError("update local collection from cloud not yet implemented", nil)
}
