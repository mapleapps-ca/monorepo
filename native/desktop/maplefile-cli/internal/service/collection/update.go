// internal/service/collection/update.go
package collection

import (
	"context"
	"encoding/base64"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
)

// UpdateInput represents the input for updating a local collection
type UpdateInput struct {
	ID             string  `json:"id"`
	Name           *string `json:"name,omitempty"`
	CollectionType *string `json:"collection_type,omitempty"`
}

// UpdateOutput represents the result of updating a local collection
type UpdateOutput struct {
	Collection *collection.Collection `json:"collection"`
}

// UpdateService defines the interface for updating local collections
type UpdateService interface {
	Update(ctx context.Context, input UpdateInput) (*UpdateOutput, error)
}

// updateService implements the UpdateService interface
type updateService struct {
	logger        *zap.Logger
	updateUseCase uc.UpdateCollectionUseCase
}

// NewUpdateService creates a new service for updating local collections
func NewUpdateService(
	logger *zap.Logger,
	updateUseCase uc.UpdateCollectionUseCase,
) UpdateService {
	logger = logger.Named("UpdateService")
	return &updateService{
		logger:        logger,
		updateUseCase: updateUseCase,
	}
}

// Update updates a local collection
func (s *updateService) Update(ctx context.Context, input UpdateInput) (*UpdateOutput, error) {
	// Validate inputs
	if input.ID == "" {
		s.logger.Error("collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Convert ID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		s.logger.Error("invalid collection ID format", zap.String("id", input.ID), zap.Error(err))
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Validate collection type if provided
	if input.CollectionType != nil &&
		*input.CollectionType != collection.CollectionTypeFolder &&
		*input.CollectionType != collection.CollectionTypeAlbum {
		s.logger.Error("invalid collection type", zap.String("type", *input.CollectionType))
		return nil, errors.NewAppError("collection type must be either 'folder' or 'album'", nil)
	}

	// Prepare the use case input
	useCaseInput := uc.UpdateCollectionInput{
		ID: objectID,
	}

	// Set name if provided
	if input.Name != nil {
		// Encrypt the name
		nameBytes := []byte(*input.Name)
		encryptedName := base64.StdEncoding.EncodeToString(nameBytes)
		useCaseInput.EncryptedName = &encryptedName
		useCaseInput.DecryptedName = input.Name
	}

	// Set collection type if provided
	if input.CollectionType != nil {
		useCaseInput.CollectionType = input.CollectionType
	}

	// Call the use case to update the collection
	collection, err := s.updateUseCase.Execute(ctx, useCaseInput)
	if err != nil {
		s.logger.Error("failed to update local collection", zap.String("id", input.ID), zap.Error(err))
		return nil, err
	}

	return &UpdateOutput{
		Collection: collection,
	}, nil
}
