// internal/usecase/user/upsert_by_email.go
package user

import (
	"context"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// UpsertByEmailUseCase defines the interface for upserting a user by email
type UpsertByEmailUseCase interface {
	Execute(ctx context.Context, user *user.User) error
}

type upsertByEmailUseCase struct {
	userRepo  user.Repository
	txManager transaction.Manager
}

// NewUpsertByEmailUseCase creates a new UpsertByEmailUseCase
func NewUpsertByEmailUseCase(
	userRepo user.Repository,
	txManager transaction.Manager,
) UpsertByEmailUseCase {
	return &upsertByEmailUseCase{
		userRepo:  userRepo,
		txManager: txManager,
	}
}

// Execute upserts a user by email
func (uc *upsertByEmailUseCase) Execute(ctx context.Context, user *user.User) error {
	if err := uc.userRepo.UpsertByEmail(ctx, user); err != nil {
		return fmt.Errorf("failed to upsert user by email: %w", err)
	}
	return nil
}
