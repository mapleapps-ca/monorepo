// cloud/backend/internal/vault/usecase/encryptedfile/getbyid.go
package encryptedfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
)

// GetEncryptedFileByIDUseCase defines operations for retrieving an encrypted file by ID
type GetEncryptedFileByIDUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) (*domain.EncryptedFile, error)
}

type getEncryptedFileByIDUseCaseImpl struct {
	repository domain.Repository
}

// NewGetEncryptedFileByIDUseCase creates a new instance of the use case
func NewGetEncryptedFileByIDUseCase(repository domain.Repository) GetEncryptedFileByIDUseCase {
	return &getEncryptedFileByIDUseCaseImpl{
		repository: repository,
	}
}

// Execute retrieves an encrypted file by its ID - simplified to just repository operations
func (uc *getEncryptedFileByIDUseCaseImpl) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) (*domain.EncryptedFile, error) {
	// Simply delegate to repository for retrieval
	return uc.repository.GetByID(ctx, id)
}
