// internal/usecase/collection/create.go
package collection

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

// CreateCollectionUseCase defines the interface for creating a local collection
type CreateCollectionUseCase interface {
	Execute(ctx context.Context, data *collection.Collection) error
}

// createCollectionUseCase implements the CreateCollectionUseCase interface
type createCollectionUseCase struct {
	logger     *zap.Logger
	repository collection.CollectionRepository
}

// NewCreateCollectionUseCase creates a new use case for creating local collections
func NewCreateCollectionUseCase(
	logger *zap.Logger,
	repository collection.CollectionRepository,
) CreateCollectionUseCase {
	return &createCollectionUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new local collection
func (uc *createCollectionUseCase) Execute(ctx context.Context, data *collection.Collection) error {
	// Validate inputs
	if data.EncryptedName == "" {
		return errors.NewAppError("encrypted name is required", nil)
	}

	if data.CollectionType != collection.CollectionTypeFolder && data.CollectionType != collection.CollectionTypeAlbum {
		return errors.NewAppError(fmt.Sprintf("invalid collection type: %s (must be '%s' or '%s')",
			data.CollectionType, collection.CollectionTypeFolder, collection.CollectionTypeAlbum), nil)
	}

	if data.EncryptedCollectionKey.Ciphertext == nil || len(data.EncryptedCollectionKey.Ciphertext) == 0 ||
		data.EncryptedCollectionKey.Nonce == nil || len(data.EncryptedCollectionKey.Nonce) == 0 {
		return errors.NewAppError("encrypted collection key is required", nil)
	}

	// Save the collection
	err := uc.repository.Create(ctx, data)
	if err != nil {
		return errors.NewAppError("failed to create local collection", err)
	}

	return nil
}
