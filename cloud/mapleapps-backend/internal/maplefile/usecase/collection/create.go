// cloud/backend/internal/maplefile/usecase/collection/create.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type CreateCollectionUseCase interface {
	Execute(ctx context.Context, collection *dom_collection.Collection) error
}

type createCollectionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewCreateCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) CreateCollectionUseCase {
	logger = logger.Named("CreateCollectionUseCase")
	return &createCollectionUseCaseImpl{config, logger, repo}
}

func (uc *createCollectionUseCaseImpl) Execute(ctx context.Context, collection *dom_collection.Collection) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collection == nil {
		e["collection"] = "Collection is required"
	} else {
		if collection.OwnerID.IsZero() {
			e["owner_id"] = "Owner ID is required"
		}
		if collection.EncryptedName == "" {
			e["encrypted_name"] = "Collection name is required"
		}
		if collection.CollectionType == "" {
			e["collection_type"] = "Collection type is required"
		} else if collection.CollectionType != dom_collection.CollectionTypeFolder && collection.CollectionType != dom_collection.CollectionTypeAlbum {
			e["collection_type"] = "Collection type must be either 'folder' or 'album'"
		}
		if collection.EncryptedCollectionKey.Ciphertext == nil || len(collection.EncryptedCollectionKey.Ciphertext) == 0 {
			e["encrypted_collection_key"] = "Encrypted collection key is required"
		}
		if collection.State == "" {
			e["state"] = "File state is required"
		} else if collection.State != dom_collection.CollectionStateActive &&
			collection.State != dom_collection.CollectionStateDeleted &&
			collection.State != dom_collection.CollectionStateArchived {
			e["state"] = "Invalid collection state"
		}
		if err := dom_collection.IsValidStateTransition(dom_collection.CollectionStateActive, collection.State); err != nil {
			e["state"] = err.Error()
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating collection creation",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Insert into database.
	//

	return uc.repo.Create(ctx, collection)
}
