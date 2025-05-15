// internal/usecase/user/list_all.go
package user

import (
	"context"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
)

// ListAllUseCase defines the interface for listing all users
type ListAllUseCase interface {
	Execute(ctx context.Context) ([]*user.User, error)
}

type listAllUseCase struct {
	userRepo user.Repository
}

// NewListAllUseCase creates a new ListAllUseCase
func NewListAllUseCase(userRepo user.Repository) ListAllUseCase {
	return &listAllUseCase{
		userRepo: userRepo,
	}
}

// Execute lists all users
func (uc *listAllUseCase) Execute(ctx context.Context) ([]*user.User, error) {
	users, err := uc.userRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all users: %w", err)
	}
	return users, nil
}
