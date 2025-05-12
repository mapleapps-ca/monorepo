// cloud/backend/internal/vault/usecase/encryptedfile/delete.go
package encryptedfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
)

// DeleteEncryptedFileUseCase defines operations for deleting an encrypted file
type DeleteEncryptedFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
}

type deleteEncryptedFileUseCaseImpl struct {
	repository domain.Repository
}

// NewDeleteEncryptedFileUseCase creates a new instance of the use case
func NewDeleteEncryptedFileUseCase(repository domain.Repository) DeleteEncryptedFileUseCase {
	return &deleteEncryptedFileUseCaseImpl{
		repository: repository,
	}
}

// Execute deletes an encrypted file - simplified to just repository operations
func (uc *deleteEncryptedFileUseCaseImpl) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	// Simply delegate to repository for deletion
	return uc.repository.DeleteByID(ctx, id)
}
