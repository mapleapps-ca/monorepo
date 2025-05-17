// monorepo/native/desktop/maplefile-cli/internal/domain/user/interface.go
package user

import (
	"context"
)

// Repository Interface for federatediam.
type Repository interface {
	UpsertByEmail(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	DeleteByEmail(ctx context.Context, email string) error
	// CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	// UpdateByID(ctx context.Context, m *User) error
	ListAll(ctx context.Context) ([]*User, error)
	// CountByFilter(ctx context.Context, filter *UserFilter) (uint64, error)
	// ListByFilter(ctx context.Context, filter *UserFilter) (*UserFilterResult, error)
	UpdateVerificationStatus(ctx context.Context, email string, verified bool, role int8, status int8) error
	OpenTransaction() error
	CommitTransaction() error
	DiscardTransaction()
}
