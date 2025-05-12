// cloud/backend/internal/vault/usecase/encryptedfile/create.go
package encryptedfile

import (
	"context"
	"io"

	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
)

// CreateEncryptedFileUseCase defines operations for creating a new encrypted file
type CreateEncryptedFileUseCase interface {
	Execute(ctx context.Context, file *domain.EncryptedFile, encryptedContent io.Reader) error
}

type createEncryptedFileUseCaseImpl struct {
	repository domain.Repository
}

// NewCreateEncryptedFileUseCase creates a new instance of the use case
func NewCreateEncryptedFileUseCase(repository domain.Repository) CreateEncryptedFileUseCase {
	return &createEncryptedFileUseCaseImpl{
		repository: repository,
	}
}

// Execute handles the creation of a new encrypted file - simplified to just repository operations
func (uc *createEncryptedFileUseCaseImpl) Execute(
	ctx context.Context,
	file *domain.EncryptedFile,
	encryptedContent io.Reader,
) error {
	// Simply delegate to repository for storage
	return uc.repository.Create(ctx, file, encryptedContent)
}
