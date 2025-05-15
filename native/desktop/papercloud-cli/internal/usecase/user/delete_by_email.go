// internal/usecase/user/delete_by_email.go
package user

import (
	"context"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/transaction"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
)

// DeleteByEmailUseCase defines the interface for deleting a user by email
type DeleteByEmailUseCase interface {
	Execute(ctx context.Context, email string) error
}

type deleteByEmailUseCase struct {
	userRepo  user.Repository
	txManager transaction.Manager
}

// NewDeleteByEmailUseCase creates a new DeleteByEmailUseCase
func NewDeleteByEmailUseCase(
	userRepo user.Repository,
	txManager transaction.Manager,
) DeleteByEmailUseCase {
	return &deleteByEmailUseCase{
		userRepo:  userRepo,
		txManager: txManager,
	}
}

// Execute deletes a user by email
func (uc *deleteByEmailUseCase) Execute(ctx context.Context, email string) error {
	if err := uc.userRepo.DeleteByEmail(ctx, email); err != nil {
		return fmt.Errorf("failed to delete user by email: %w", err)
	}
	return nil
}
