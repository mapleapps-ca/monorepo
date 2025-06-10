// internal/service/collection/update.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// UpdateInput represents the input for updating a local collection
type UpdateInput struct {
	ID             gocql.UUID `json:"id"`
	Name           *string    `json:"name,omitempty"`
	CollectionType *string    `json:"collection_type,omitempty"`
	UserPassword   string     `json:"user_password,omitempty"`
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
	logger                      *zap.Logger
	updateUseCase               uc.UpdateCollectionUseCase
	getUseCase                  uc.GetCollectionUseCase
	getUserByIsLoggedInUseCase  uc_user.GetByIsLoggedInUseCase
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService
	collectionEncryptionService svc_collectioncrypto.CollectionEncryptionService
}

// NewUpdateService creates a new service for updating local collections
func NewUpdateService(
	logger *zap.Logger,
	updateUseCase uc.UpdateCollectionUseCase,
	getUseCase uc.GetCollectionUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
	collectionEncryptionService svc_collectioncrypto.CollectionEncryptionService,
) UpdateService {
	logger = logger.Named("UpdateService")
	return &updateService{
		logger:                      logger,
		updateUseCase:               updateUseCase,
		getUseCase:                  getUseCase,
		getUserByIsLoggedInUseCase:  getUserByIsLoggedInUseCase,
		collectionDecryptionService: collectionDecryptionService,
		collectionEncryptionService: collectionEncryptionService,
	}
}

// Update updates a local collection using proper E2EE
func (s *updateService) Update(ctx context.Context, input UpdateInput) (*UpdateOutput, error) {
	// Validate inputs
	if input.ID.String() == "" {
		s.logger.Error("❌ collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Validate collection type if provided
	if input.CollectionType != nil &&
		*input.CollectionType != collection.CollectionTypeFolder &&
		*input.CollectionType != collection.CollectionTypeAlbum {
		s.logger.Error("❌ invalid collection type", zap.String("type", *input.CollectionType))
		return nil, errors.NewAppError("collection type must be either 'folder' or 'album'", nil)
	}

	// Prepare the use case input
	useCaseInput := uc.UpdateCollectionInput{
		ID: input.ID,
	}

	// Proper E2EE encryption instead of base64 encoding
	if input.Name != nil {
		if input.UserPassword == "" {
			s.logger.Error("❌ user password is required for E2EE name encryption")
			return nil, errors.NewAppError("user password is required for E2EE name encryption", nil)
		}

		// Get user for E2EE operations
		user, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
		if err != nil {
			return nil, errors.NewAppError("failed to get logged in user", err)
		}
		if user == nil {
			return nil, errors.NewAppError("user not found", nil)
		}

		// Get collection for E2EE operations
		currentCollection, err := s.getUseCase.Execute(ctx, input.ID)
		if err != nil {
			return nil, errors.NewAppError("failed to get collection for E2EE", err)
		}
		if currentCollection == nil {
			return nil, errors.NewAppError("collection not found", nil)
		}

		// Decrypt collection key using E2EE chain
		s.logger.Debug("🔐 Decrypting collection key for name encryption using crypto service")
		collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, user, currentCollection, input.UserPassword)
		if err != nil {
			return nil, errors.NewAppError("failed to decrypt collection key for name encryption", err)
		}
		defer crypto.ClearBytes(collectionKey)

		// Encrypt the name using proper E2EE
		s.logger.Debug("🔐 Encrypting collection name using crypto service")
		encryptedName, err := s.collectionEncryptionService.ExecuteForEncryptData(ctx, *input.Name, collectionKey)
		if err != nil {
			return nil, errors.NewAppError("failed to encrypt collection name", err)
		}

		useCaseInput.EncryptedName = &encryptedName
		useCaseInput.DecryptedName = input.Name

		s.logger.Debug("✅ Successfully encrypted collection name using crypto service")
	}

	// Set collection type if provided
	if input.CollectionType != nil {
		useCaseInput.CollectionType = input.CollectionType
	}

	// Call the use case to update the collection
	collection, err := s.updateUseCase.Execute(ctx, useCaseInput)
	if err != nil {
		s.logger.Error("❌ failed to update local collection",
			zap.String("id", input.ID.String()),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("✅ Successfully updated collection using E2EE crypto service",
		zap.String("id", input.ID.String()))

	return &UpdateOutput{
		Collection: collection,
	}, nil
}
