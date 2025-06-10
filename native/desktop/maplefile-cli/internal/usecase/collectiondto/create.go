// internal/usecase/collectiondto/create.go
package collectiondto

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// CreateCollectionInCloudUseCase defines the interface for creating a cloud collection
type CreateCollectionInCloudUseCase interface {
	Execute(ctx context.Context, dto *collectiondto.CollectionDTO) (*gocql.UUID, error)
}

// createCollectionInCloudUseCase implements the CreateCollectionInCloudUseCase interface
type createCollectionInCloudUseCase struct {
	logger     *zap.Logger
	repository collectiondto.CollectionDTORepository
}

// NewCreateCollectionInCloudUseCase creates a new use case for creating cloud collections
func NewCreateCollectionInCloudUseCase(
	logger *zap.Logger,
	repository collectiondto.CollectionDTORepository,
) CreateCollectionInCloudUseCase {
	logger = logger.Named("CreateCollectionInCloudUseCase")
	return &createCollectionInCloudUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new cloud collection
func (uc *createCollectionInCloudUseCase) Execute(ctx context.Context, dto *collectiondto.CollectionDTO) (*gocql.UUID, error) {
	//
	// STEP 1: Validate the input
	//

	e := make(map[string]string)

	if dto == nil {
		e["dto"] = "DTO is required"
	} else {
		// Validate required fields
		if dto.OwnerID.IsZero() {
			e["owner_id"] = "OwnerID is required"
		}
		if dto.EncryptedName == "" {
			e["encrypted_name"] = "EncryptedName is required"
		}
		if dto.CollectionType == "" {
			e["collection_type"] = "Collection type is required"
		}
		if dto.CreatedAt.IsZero() {
			e["created_at"] = "CreatedAt is required"
		}
		if dto.ModifiedAt.IsZero() {
			e["modified_at"] = "ModifiedAt is required"
		}
		if dto.EncryptedCollectionKey == nil {
			e["encrypted_collection_key"] = "EncryptedCollectionKey is required"
		} else {
			if dto.EncryptedCollectionKey.Ciphertext == nil {
				e["ciphertext"] = "EncryptedCollectionKey-Ciphertext is required"
			}
			if dto.EncryptedCollectionKey.Nonce == nil {
				e["nonce"] = "EncryptedCollectionKey-Nonce is required"
			}
			if dto.EncryptedCollectionKey.KeyVersion == 0 {
				e["key_version"] = "EncryptedCollectionKey-KeyVersion is required"
			}
		}
	}

	// If any errors were found, return bad request error
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Submit our collection to the cloud.
	//

	// Call the repository to create the collection
	cloudIDResponse, err := uc.repository.CreateInCloud(ctx, dto)
	if err != nil {
		return nil, errors.NewAppError("failed to create collection in the cloud", err)
	}

	//
	// STEP 3: Return our unique Cloud ID response from the cloud.
	//

	return cloudIDResponse, nil
}
