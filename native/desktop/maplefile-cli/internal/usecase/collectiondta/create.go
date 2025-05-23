// internal/usecase/collectiondto/create.go
package collectiondto

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// CreateCollectionDTOInput defines the input for creating a cloud collection
type CreateCollectionDTOInput struct {
	EncryptedName          string
	Type                   string
	ParentID               *primitive.ObjectID
	EncryptedPathSegments  []string
	EncryptedCollectionKey keys.EncryptedCollectionKey
}

// CreateCollectionDTOUseCase defines the interface for creating a cloud collection
type CreateCollectionDTOUseCase interface {
	Execute(ctx context.Context, collectionDTO *collectiondto.CollectionDTO, accessToken string) (*primitive.ObjectID, error)
}

// createCollectionDTOUseCase implements the CreateCollectionDTOUseCase interface
type createCollectionDTOUseCase struct {
	logger     *zap.Logger
	repository collectiondto.CollectionDTORepository
}

// NewCreateCollectionDTOUseCase creates a new use case for creating cloud collections
func NewCreateCollectionDTOUseCase(
	logger *zap.Logger,
	repository collectiondto.CollectionDTORepository,
) CreateCollectionDTOUseCase {
	return &createCollectionDTOUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new cloud collection
func (uc *createCollectionDTOUseCase) Execute(ctx context.Context, dto *collectiondto.CollectionDTO, accessToken string) (*primitive.ObjectID, error) {
	// TODO: Validation

	// Call the repository to create the collection
	response, err := uc.repository.CreateInCloud(ctx, dto, accessToken)
	if err != nil {
		return nil, errors.NewAppError("failed to create cloud collection", err)
	}

	return response, nil
}
