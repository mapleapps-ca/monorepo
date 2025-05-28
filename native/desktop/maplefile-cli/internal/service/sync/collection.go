// internal/service/sync/collection.go
package sync

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	dom_syncdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
	syncdtoSvc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncstate"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
)

// SyncCollectionsInput represents input for syncing collections, allowing customization of batching.
type SyncCollectionsInput struct {
	BatchSize  int64 `json:"batch_size,omitempty"`  // The maximum number of items per batch received from the cloud sync service.
	MaxBatches int   `json:"max_batches,omitempty"` // The maximum number of batches to process in a single sync run.
}

// SyncCollectionService defines the interface for synchronizing collection data from a remote source (cloud)
// to the local storage.
type SyncCollectionService interface {
	// Execute performs the collection synchronization process.
	// It fetches collections in batches based on the current sync state, processes the changes,
	// and updates the local storage and sync state.
	Execute(ctx context.Context, input *SyncCollectionsInput) (*syncdto.SyncResult, error)
}

// syncCollectionService implements the SyncCollectionService interface, coordinating
// the retrieval of cloud collection data, processing it, and updating the local collection repository.
type syncCollectionService struct {
	logger *zap.Logger // Logger instance for structured logging.

	// Services for managing the sync state (cursor)
	syncStateGetService   syncstate.GetService
	syncStateSaveService  syncstate.SaveService
	syncStateResetService syncstate.ResetService

	// Service for fetching collection data from the remote source (cloud)
	syncDTOProgressService        syncdtoSvc.SyncProgressService
	getCollectionFromCloudUseCase uc_collectiondto.GetCollectionFromCloudUseCase

	// Use cases for interacting with the local collection repository
	createCollectionUseCase uc.CreateCollectionUseCase
	getCollectionUseCase    uc.GetCollectionUseCase
	updateCollectionUseCase uc.UpdateCollectionUseCase
	deleteCollectionUseCase uc.DeleteCollectionUseCase
}

// NewSyncCollectionService creates a new instance of syncCollectionService.
// It takes all necessary dependencies as arguments.
func NewSyncCollectionService(
	logger *zap.Logger,
	syncStateGetService syncstate.GetService,
	syncStateSaveService syncstate.SaveService,
	syncStateResetService syncstate.ResetService,
	syncDTOProgressService syncdtoSvc.SyncProgressService,
	getCollectionFromCloudUseCase uc_collectiondto.GetCollectionFromCloudUseCase,
	createCollectionUseCase uc.CreateCollectionUseCase,
	getCollectionUseCase uc.GetCollectionUseCase,
	updateCollectionUseCase uc.UpdateCollectionUseCase,
	deleteCollectionUseCase uc.DeleteCollectionUseCase,
) SyncCollectionService {
	return &syncCollectionService{
		logger: logger,

		syncStateGetService:   syncStateGetService,
		syncStateSaveService:  syncStateSaveService,
		syncStateResetService: syncStateResetService,

		syncDTOProgressService:        syncDTOProgressService,
		getCollectionFromCloudUseCase: getCollectionFromCloudUseCase,

		createCollectionUseCase: createCollectionUseCase,
		getCollectionUseCase:    getCollectionUseCase,
		updateCollectionUseCase: updateCollectionUseCase,
		deleteCollectionUseCase: deleteCollectionUseCase,
	}
}

// Execute synchronizes collections from the cloud based on the current sync state.
// It fetches collection data in batches, processes each collection (create/update/delete),
// and updates the sync state upon successful completion of fetching batches.
func (s *syncCollectionService) Execute(ctx context.Context, input *SyncCollectionsInput) (*syncdto.SyncResult, error) {
	s.logger.Info("Starting collection synchronization")

	// Set default input parameters if not provided
	if input == nil {
		input = &SyncCollectionsInput{}
	}
	if input.BatchSize <= 0 {
		input.BatchSize = 50 // Default batch size
	}
	if input.MaxBatches <= 0 {
		input.MaxBatches = 100 // Default max batches
	}

	s.logger.Debug("Collection sync input parameters",
		zap.Int("batchSize", int(input.BatchSize)),   // Cast to int for logging
		zap.Int("maxBatches", int(input.MaxBatches))) // Cast to int for logging

	// Retrieve the current sync state to determine the starting point for the sync
	s.logger.Debug("Getting current sync state for collections")
	syncStateOutput, err := s.syncStateGetService.GetSyncState(ctx)
	if err != nil {
		s.logger.Error("Failed to get sync state for collections", zap.Error(err))
		return nil, errors.NewAppError("failed to get sync state", err)
	}
	s.logger.Debug("Successfully retrieved sync state for collections",
		zap.Time("lastCollectionSync", syncStateOutput.SyncState.LastCollectionSync),
		zap.String("lastCollectionID", syncStateOutput.SyncState.LastCollectionID.Hex())) // Convert ObjectID to string for logging

	// Build the sync cursor based on the retrieved sync state
	var currentSyncCursor *dom_syncdto.SyncCursorDTO
	if !syncStateOutput.SyncState.LastCollectionSync.IsZero() {
		// If a previous sync state exists, use it to create the cursor
		currentSyncCursor = &dom_syncdto.SyncCursorDTO{
			LastModified: syncStateOutput.SyncState.LastCollectionSync,
			LastID:       syncStateOutput.SyncState.LastCollectionID,
		}
		s.logger.Debug("Using existing cursor for collection sync",
			zap.Time("lastModified", currentSyncCursor.LastModified),
			zap.String("lastID", currentSyncCursor.LastID.Hex())) // Convert ObjectID to string for logging
	} else {
		// If no previous sync state exists, start syncing from the beginning (nil cursor)
		s.logger.Debug("No previous sync state found for collections, starting from beginning")
	}

	// Prepare input for the progress service to fetch collections
	progressInput := &syncdtoSvc.SyncProgressInput{
		SyncType:       "collections",         // Type of data being synced
		StartCursor:    currentSyncCursor,     // Cursor indicating where to start fetching
		BatchSize:      input.BatchSize,       // Requested batch size
		MaxBatches:     int(input.MaxBatches), // Maximum number of batches to retrieve
		TimeoutSeconds: 300,                   // Timeout for the entire fetching process (5 minutes)
	}
	s.logger.Debug("Calling progress service for GetAllCollections",
		zap.Any("progressInput", progressInput))

	// Fetch collection data in batches from the remote sync service
	progressOutput, err := s.syncDTOProgressService.GetAllCollections(ctx, progressInput)
	if err != nil {
		s.logger.Error("Failed to get collections sync data from progress service", zap.Error(err))
		return nil, errors.NewAppError("failed to get collections sync data", err)
	}

	// Log summary of the fetched sync data
	s.logger.Info("Received collection sync data summary",
		zap.Int("totalItems", progressOutput.TotalItems),                  // Total number of items across all batches
		zap.Int("batchesReceived", len(progressOutput.CollectionBatches)), // Number of batches received
		zap.Any("finalCursor", progressOutput.FinalCursor))                // The cursor to use for the next sync

	// Initialize the sync result structure
	collectionSyncResult := &dom_syncdto.SyncResult{
		CollectionsProcessed: progressOutput.TotalItems,
	}

	// Process each batch of collections received from the sync service
	// Analyze the sync data to determine what was added/updated/deleted
	// This is a simplified implementation - in a real scenario, you'd compare
	// with local data to determine the actual operations needed
	for batchIndex, batch := range progressOutput.CollectionBatches {
		s.logger.Debug("Processing collection batch",
			zap.Int("batchIndex", batchIndex),
			zap.Int("itemsInBatch", len(batch.Collections)))

		// Process each individual collection within the current batch
		for _, cloudCollection := range batch.Collections {
			// Log detailed information about the collection being analyzed
			s.logger.Debug("Beginning to analyze collection for syncing...",
				zap.String("id", cloudCollection.ID.Hex()),
				zap.Uint64("version", cloudCollection.Version),
				zap.Time("modified_at", cloudCollection.ModifiedAt),
				zap.String("state", cloudCollection.State),
				zap.Any("parent_id", cloudCollection.ParentID), // Use Any for potential nil or different types
				zap.Uint64("tombstone_version", cloudCollection.TombstoneVersion),
				zap.Time("tombstone_expiry", cloudCollection.TombstoneExpiry),
			)

			// Attempt to lookup the existing local collection record using the ID from the cloud data.
			existingLocalCollection, err := s.getCollectionUseCase.Execute(ctx, cloudCollection.ID)
			if err != nil {
				// Log error if lookup fails but continue processing other items
				s.logger.Error("Failed to get local collection",
					zap.String("id", cloudCollection.ID.Hex()),
					zap.Error(err))
				// Depending on error type, might need to handle specifically (e.g., not found vs actual DB error)
				continue // Skip processing this collection if local lookup fails
			}

			// The following switch is a simplified placeholder based only on cloudCollection.State:
			switch cloudCollection.State {
			case "active":
				//
				// CASE 1: If the local collection is not found, create a new one.
				//
				if existingLocalCollection == nil { // This check needs correction
					// Check that the collection is not deleted prior to adding it again into our app.
					if err := s.CreateLocalCollectionFromCloud(ctx, &cloudCollection); err != nil {
						s.logger.Error("Failed to get cloud collection and save it locally",
							zap.String("id", cloudCollection.ID.Hex()),
							zap.Error(err))
						// Depending on error type, might need to handle specifically (e.g., not found vs actual DB error)
						continue // Skip processing this collection if local lookup fails
					}
				} else {
					//
					// CASE 2: If the local collection exists, check if it needs to be updated.
					//

					// TODO: Add logic here to compare cloudCollection and existingLocalCollection
					// and call s.updateCollectionUseCase.Execute if necessary.
					s.logger.Debug("Local collection found, need to implement update logic.",
						zap.String("id", cloudCollection.ID.Hex()))
					// For now, just incrementing updated count as a placeholder
					collectionSyncResult.CollectionsUpdated++
				}

			case "deleted":
				// Simplified logic: If cloud state is "deleted", assume it needs to be deleted locally.
				// TODO: Add logic here to check if existingLocalCollection exists and then call s.deleteCollectionUseCase.Execute
				s.logger.Debug("Collection marked as deleted (needs local deletion)",
					zap.String("id", cloudCollection.ID.Hex())) // Optional: log each item
				collectionSyncResult.CollectionsDeleted++
			case "":
				// Handle empty state as an error
				errorMsg := "empty collection state received from sync service"
				s.logger.Warn(errorMsg,
					zap.String("id", cloudCollection.ID.Hex())) // Convert ObjectID to string for logging
				collectionSyncResult.Errors = append(collectionSyncResult.Errors, errorMsg+" for ID: "+cloudCollection.ID.Hex()) // Convert ObjectID to string for concatenation
			default:
				// Handle unknown state as an error
				errorMsg := "unknown collection state received from sync service: " + cloudCollection.State
				s.logger.Warn(errorMsg,
					zap.String("id", cloudCollection.ID.Hex())) // Convert ObjectID to string for logging
				collectionSyncResult.Errors = append(collectionSyncResult.Errors, errorMsg+" for ID: "+cloudCollection.ID.Hex()) // Convert ObjectID to string for concatenation
			}
		}
	}

	// TODO: UNCOMMENT THE CODE BELOW WHEN THE SYNC CODE ABOVE IS COMPLETED
	// This block saves the final cursor received from the progress service,
	// allowing the next sync run to resume from where this one left off.
	// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - //

	// // Update sync state if we processed any data and got a final cursor
	// if progressOutput.TotalItems > 0 && progressOutput.FinalCursor != nil {
	// 	saveInput := &syncstate.SaveInput{
	// 		LastCollectionSync: &progressOutput.FinalCursor.LastModified,
	// 		LastCollectionID:   &progressOutput.FinalCursor.LastID, // Use pointer to ObjectID if SaveInput expects pointer
	// 	}
	// 	s.logger.Debug("Attempting to save sync state for collections",
	// 		zap.Time("lastCollectionSync", *saveInput.LastCollectionSync),
	// 		zap.String("lastCollectionID", saveInput.LastCollectionID.Hex())) // Convert ObjectID to string for logging

	// 	_, err = s.syncStateSaveService.SaveSyncState(ctx, saveInput)
	// 	if err != nil {
	// 		s.logger.Error("Failed to update sync state for collections", zap.Error(err))
	// 		// Don't fail the entire operation for sync state update failure, just log and add to errors
	// 		collectionSyncResult.Errors = append(collectionSyncResult.Errors, "failed to update sync state: "+err.Error())
	// 	} else {
	// 		s.logger.Info("Successfully updated sync state for collections")
	// 	}
	// } else if progressOutput.TotalItems > 0 && progressOutput.FinalCursor == nil {
	// 	// This case indicates an issue where items were processed but no final cursor was provided.
	// 	s.logger.Warn("Processed items but did not receive a final cursor for collections. Sync state not updated.")
	// 	collectionSyncResult.Errors = append(collectionSyncResult.Errors, "processed items but no final sync cursor received")
	// } else {
	// 	// No items processed, likely nothing new to sync in this run.
	// 	s.logger.Info("No items processed for collections. Sync state not updated.")
	// }
	//
	// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - //

	// Log final summary of the synchronization process
	s.logger.Info("Collection synchronization completed",
		zap.Int("processed", collectionSyncResult.CollectionsProcessed), // Total items received from sync service
		zap.Int("updated", collectionSyncResult.CollectionsUpdated),     // Items locally created or updated (based on placeholder logic)
		zap.Int("deleted", collectionSyncResult.CollectionsDeleted),     // Items marked for local deletion (based on placeholder logic)
		zap.Int("errors", len(collectionSyncResult.Errors)))             // Number of errors encountered during processing

	return collectionSyncResult, nil
}

func (s *syncCollectionService) CreateLocalCollectionFromCloud(ctx context.Context, collectionSyncItem *dom_syncdto.CollectionSyncItem) error {
	s.logger.Info("Starting creating local collection from cloud")
	collectionDTO, err := s.getCollectionFromCloudUseCase.Execute(ctx, collectionSyncItem.ID)
	if err != nil {
		s.logger.Error("Failed to fetch collection from cloud",
			zap.Error(err))
		return err
	}
	if collectionDTO == nil {
		err := errors.NewAppError("collection not found", nil)
		s.logger.Error("Failed to fetch collection from cloud",
			zap.Error(err))
		return err
	}

	// Create a new collection domain object from the cloud data using a mapping function.
	newCollection := mapCollectionDTOToDomain(collectionDTO)

	// Execute the use case to create the local collection record.
	if err := s.createCollectionUseCase.Execute(ctx, newCollection); err != nil {
		s.logger.Error("Failed to create new collection",
			zap.String("id", collectionDTO.ID.Hex()),
			zap.Error(err))
		return err
	}
	// // TODO: Increment result.CollectionsCreated instead of CollectionsUpdated?
	// // The current code increments CollectionsUpdated for both create and update paths in the simplified logic.
	// collectionSyncResult.CollectionsUpdated++ // Assuming this covers both new creations and updates for active state.
	// s.logger.Debug("Collection marked as active (created/updated placeholder)",
	// 	zap.String("id", cloudCollection.ID.Hex())) // Optional: log each item

	return nil
}

// mapCollectionDTOToDomain maps a single uc_collectiondto.CollectionDTO to a single dom_collection.Collection.
// It handles nested Members and Children recursively by calling helper slice mappers.
func mapCollectionDTOToDomain(dto *dom_collectiondto.CollectionDTO) *dom_collection.Collection {
	if dto == nil {
		return nil
	}

	// Assuming dom_collection.Collection has fields compatible with uc_collectiondto.CollectionDTO
	return &dom_collection.Collection{
		ID:                     dto.ID,
		OwnerID:                dto.OwnerID,
		EncryptedName:          dto.EncryptedName,
		CollectionType:         dto.CollectionType,
		EncryptedCollectionKey: dto.EncryptedCollectionKey,         // Assuming keys.EncryptedCollectionKey type is compatible
		Members:                mapMembersDTOToDomain(dto.Members), // Call helper for members slice
		ParentID:               dto.ParentID,
		AncestorIDs:            dto.AncestorIDs,
		Children:               mapChildrenDTOToDomain(dto.Children), // Call helper for children slice
		CreatedAt:              dto.CreatedAt,
		CreatedByUserID:        dto.CreatedByUserID,
		ModifiedAt:             dto.ModifiedAt,
		ModifiedByUserID:       dto.ModifiedByUserID,
		Version:                dto.Version,
		State:                  dto.State,
		// Note: TombstoneVersion and TombstoneExpiry from sync item are not part of the main CollectionDTO
		// and are not mapped here into the domain object. They are specific to the sync process.
	}
}

// mapMembersDTOToDomain maps a slice of uc_collectiondto.CollectionMembershipDTO to a slice of dom_collection.CollectionMembership.
// Assuming dom_collection.CollectionMembership struct matches uc_collectiondto.CollectionMembershipDTO.
func mapMembersDTOToDomain(membersDTO []*dom_collectiondto.CollectionMembershipDTO) []*dom_collection.CollectionMembership {
	if membersDTO == nil {
		return nil
	}
	members := make([]*dom_collection.CollectionMembership, len(membersDTO))
	for i, memberDTO := range membersDTO {
		if memberDTO != nil {
			// Assuming dom_collection.CollectionMembership has fields matching uc_collectiondto.CollectionMembershipDTO
			members[i] = &dom_collection.CollectionMembership{
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
	}
	return members
}

// mapChildrenDTOToDomain maps a slice of uc_collectiondto.CollectionDTO to a slice of dom_collection.Collection.
// It recursively calls mapCollectionDTOToDomain for each child.
func mapChildrenDTOToDomain(childrenDTO []*dom_collectiondto.CollectionDTO) []*dom_collection.Collection {
	if childrenDTO == nil {
		return nil
	}
	children := make([]*dom_collection.Collection, len(childrenDTO))
	for i, childDTO := range childrenDTO {
		children[i] = mapCollectionDTOToDomain(childDTO) // Recursive call to single item mapper
	}
	return children
}
