// internal/usecase/collectionsyncer/upload.go
package collectionsyncer

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_localcollection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
	dom_remotecollection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
	uc_localcollection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localcollection"
	uc_remotecollection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotecollection"
)

// UploadToRemoteUseCase defines the interface for uploading local collections to remote
type UploadToRemoteUseCase interface {
	Execute(ctx context.Context, localID primitive.ObjectID) (*remotecollection.RemoteCollectionResponse, error)
}

// uploadToRemoteUseCase implements the UploadToRemoteUseCase interface
type uploadToRemoteUseCase struct {
	logger              *zap.Logger
	localRepository     dom_localcollection.LocalCollectionRepository
	getLocalUseCase     uc_localcollection.GetLocalCollectionUseCase
	createRemoteUseCase uc_remotecollection.CreateRemoteCollectionUseCase
}

// NewUploadToRemoteUseCase creates a new use case for uploading local collections
func NewUploadToRemoteUseCase(
	logger *zap.Logger,
	localRepository dom_localcollection.LocalCollectionRepository,
	getLocalUseCase uc_localcollection.GetLocalCollectionUseCase,
	createRemoteUseCase uc_remotecollection.CreateRemoteCollectionUseCase,
) UploadToRemoteUseCase {
	return &uploadToRemoteUseCase{
		logger:              logger,
		localRepository:     localRepository,
		getLocalUseCase:     getLocalUseCase,
		createRemoteUseCase: createRemoteUseCase,
	}
}

// Execute uploads a local collection to the remote server
func (uc *uploadToRemoteUseCase) Execute(
	ctx context.Context,
	localID primitive.ObjectID,
) (*dom_remotecollection.RemoteCollectionResponse, error) {
	// Validate inputs
	if localID.IsZero() {
		return nil, errors.NewAppError("local collection ID is required", nil)
	}

	// Get the local collection
	localCollection, err := uc.getLocalUseCase.Execute(ctx, localID)
	if err != nil {
		return nil, errors.NewAppError("failed to get local collection", err)
	}

	if localCollection == nil {
		return nil, errors.NewAppError("local collection not found", nil)
	}

	// Create input for the remote collection creation
	input := uc_remotecollection.CreateRemoteCollectionInput{
		EncryptedName:          localCollection.EncryptedName,
		Type:                   localCollection.Type,
		EncryptedCollectionKey: localCollection.EncryptedCollectionKey,
		EncryptedPathSegments:  localCollection.EncryptedPathSegments,
	}

	// Set parent ID if it's not zero
	if !localCollection.ParentID.IsZero() {
		input.ParentID = &localCollection.ParentID
	}

	// Create the remote collection
	response, err := uc.createRemoteUseCase.Execute(ctx, input)
	if err != nil {
		return nil, errors.NewAppError("failed to create remote collection", err)
	}

	// Update the local collection with sync info
	localCollection.LastSyncedAt = time.Now()
	localCollection.IsModifiedLocally = false

	// Save the updated local collection
	err = uc.localRepository.Save(ctx, localCollection)
	if err != nil {
		return nil, errors.NewAppError("failed to update local collection sync status", err)
	}

	return response, nil
}
