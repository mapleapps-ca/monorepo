// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/repo/federateduser/get.go
package federateduser

import (
	"context"

	"github.com/gocql/gocql"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
)

func (impl userStorerImpl) GetByID(ctx context.Context, id gocql.UUID) (*dom_user.FederatedUser, error) {
	return nil, nil
}

func (impl userStorerImpl) GetByEmail(ctx context.Context, email string) (*dom_user.FederatedUser, error) {
	return nil, nil
}

func (impl userStorerImpl) GetByVerificationCode(ctx context.Context, verificationCode string) (*dom_user.FederatedUser, error) {
	return nil, nil
}
