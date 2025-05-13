// monorepo/native/desktop/papercloud-cli/internal/domain/user/interface.go
package user

import (
	"context"
)

// Repository Interface for federatediam.
type Repository interface {
	Create(ctx context.Context, m *User) error
	// GetByID(ctx context.Context, id primitive.ObjectID) (*User, error)
	// GetByEmail(ctx context.Context, email string) (*User, error)
	// GetByVerificationCode(ctx context.Context, verificationCode string) (*User, error)
	// DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// DeleteByEmail(ctx context.Context, email string) error
	// CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	// UpdateByID(ctx context.Context, m *User) error
	// ListAll(ctx context.Context) ([]*User, error)
	// CountByFilter(ctx context.Context, filter *UserFilter) (uint64, error)
	// ListByFilter(ctx context.Context, filter *UserFilter) (*UserFilterResult, error)
}
