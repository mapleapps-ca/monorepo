// internal/usecase/user/get_by_email.go
package user

import (
	"context"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// GetByEmailUseCase defines the interface for retrieving a user by email
type GetByEmailUseCase interface {
	Execute(ctx context.Context, email string) (*user.User, error)
}

type getByEmailUseCase struct {
	userRepo user.Repository
}

// NewGetByEmailUseCase creates a new GetByEmailUseCase
func NewGetByEmailUseCase(userRepo user.Repository) GetByEmailUseCase {
	return &getByEmailUseCase{
		userRepo: userRepo,
	}
}

// Execute retrieves a user by email
func (uc *getByEmailUseCase) Execute(ctx context.Context, email string) (*user.User, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}
