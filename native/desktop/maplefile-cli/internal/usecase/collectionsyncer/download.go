// internal/usecase/collectionsyncer/download.go
package collectionsyncer

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
	uc_remotecollection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotecollection"
)

// DownloadToLocalUseCase defines the interface for downloading remote collections to local
type DownloadToLocalUseCase interface {
	Execute(ctx context.Context, remoteID primitive.ObjectID) (*localcollection.LocalCollection, error)
}

// downloadToLocalUseCase implements the DownloadToLocalUseCase interface
type downloadToLocalUseCase struct {
	logger             *zap.Logger
	localRepository    localcollection.LocalCollectionRepository
	fetchRemoteUseCase uc_remotecollection.FetchRemoteCollectionUseCase
	listLocalUseCase   ListLocalCollectionsWithServerIDUseCase
}

// NewDownloadToLocalUseCase creates a new use case for downloading remote collections
func NewDownloadToLocalUseCase(
	logger *zap.Logger,
	localRepository localcollection.LocalCollectionRepository,
	fetchRemoteUseCase uc_remotecollection.FetchRemoteCollectionUseCase,
	listLocalUseCase ListLocalCollectionsWithServerIDUseCase,
) DownloadToLocalUseCase {
	return &downloadToLocalUseCase{
		logger:             logger,
		localRepository:    localRepository,
		fetchRemoteUseCase: fetchRemoteUseCase,
		listLocalUseCase:   listLocalUseCase,
	}
}

// Execute downloads a remote collection and creates/updates a local copy
func (uc *downloadToLocalUseCase) Execute(
	ctx context.Context,
	remoteID primitive.ObjectID,
) (*localcollection.LocalCollection, error) {
	// Validate inputs
	if remoteID.IsZero() {
		return nil, errors.NewAppError("remote collection ID is required", nil)
	}

	// Check if we already have a local copy
	existingLocal, err := uc.listLocalUseCase.FindByServerID(ctx, remoteID)
	if err != nil {
		return nil, errors.NewAppError("failed to check for existing local copy", err)
	}

	// Fetch the remote collection
	remoteCollection, err := uc.fetchRemoteUseCase.Execute(ctx, remoteID)
	if err != nil {
		return nil, errors.NewAppError("failed to fetch remote collection", err)
	}

	if remoteCollection == nil {
		return nil, errors.NewAppError("remote collection not found", nil)
	}

	// If we have a local copy, update it
	if existingLocal != nil {
		// Update fields from remote
		existingLocal.EncryptedName = remoteCollection.EncryptedName
		existingLocal.Type = remoteCollection.Type
		existingLocal.EncryptedCollectionKey = remoteCollection.EncryptedCollectionKey
		existingLocal.EncryptedPathSegments = remoteCollection.EncryptedPathSegments
		existingLocal.LastSyncedAt = time.Now()
		existingLocal.IsModifiedLocally = false

		// Save the updated collection
		err = uc.localRepository.Save(ctx, existingLocal)
		if err != nil {
			return nil, errors.NewAppError("failed to update local collection", err)
		}

		return existingLocal, nil
	}

	// Create a new local collection
	localCollection := &localcollection.LocalCollection{
		ID:                     primitive.NewObjectID(),
		OwnerID:                remoteCollection.OwnerID,
		EncryptedName:          remoteCollection.EncryptedName,
		Type:                   remoteCollection.Type,
		ParentID:               remoteCollection.ParentID,
		AncestorIDs:            remoteCollection.AncestorIDs,
		EncryptedPathSegments:  remoteCollection.EncryptedPathSegments,
		EncryptedCollectionKey: remoteCollection.EncryptedCollectionKey,
		CreatedAt:              time.Now(),
		ModifiedAt:             time.Now(),
		LastSyncedAt:           time.Now(),
		IsModifiedLocally:      false,
	}

	// Create the local collection
	err = uc.localRepository.Create(ctx, localCollection)
	if err != nil {
		return nil, errors.NewAppError("failed to create local collection", err)
	}

	return localCollection, nil
}
