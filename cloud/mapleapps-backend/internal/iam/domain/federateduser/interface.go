// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser/interface.go
package federateduser

import (
	"context"

	"github.com/gocql/gocql"
)

// FederatedUserRepository Interface for federatediam.
type FederatedUserRepository interface {
	Create(ctx context.Context, m *FederatedUser) error
	GetByID(ctx context.Context, id gocql.UUID) (*FederatedUser, error)
	GetByEmail(ctx context.Context, email string) (*FederatedUser, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*FederatedUser, error)
	DeleteByID(ctx context.Context, id gocql.UUID) error
	DeleteByEmail(ctx context.Context, email string) error
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *FederatedUser) error
}
