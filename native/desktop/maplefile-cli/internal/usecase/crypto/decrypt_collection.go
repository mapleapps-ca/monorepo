// internal/usecase/crypto/decrypt_collection.go
package crypto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

// DecryptCollectionNameUseCase defines the interface for decrypting collection names
type DecryptCollectionNameUseCase interface {
	Execute(ctx context.Context, collection *collection.Collection, collectionKey []byte) (string, error)
}

// decryptCollectionNameUseCase implements the DecryptCollectionNameUseCase interface
type decryptCollectionNameUseCase struct {
	logger *zap.Logger
	crypto CryptoService
}

// NewDecryptCollectionNameUseCase creates a new use case for decrypting collection names
func NewDecryptCollectionNameUseCase(
	logger *zap.Logger,
	crypto CryptoService,
) DecryptCollectionNameUseCase {
	return &decryptCollectionNameUseCase{
		logger: logger,
		crypto: crypto,
	}
}

// Execute decrypts a collection name
func (uc *decryptCollectionNameUseCase) Execute(
	ctx context.Context,
	collection *collection.Collection,
	collectionKey []byte,
) (string, error) {
	// Validate inputs
	if collection == nil {
		return "", errors.NewAppError("collection is required", nil)
	}

	if len(collection.EncryptedName) == 0 {
		return "", errors.NewAppError("encrypted name is empty", nil)
	}

	if len(collectionKey) == 0 {
		return "", errors.NewAppError("collection key is required", nil)
	}

	// Decrypt the collection name
	decryptedName, err := uc.crypto.DecryptString(collection.EncryptedName, collectionKey)
	if err != nil {
		return "", errors.NewAppError("failed to decrypt collection name", err)
	}

	return decryptedName, nil
}

// CryptoService defines the interface for cryptographic operations
type CryptoService interface {
	DecryptString(encryptedString string, key []byte) (string, error)
	// Add other cryptographic methods as needed
}
