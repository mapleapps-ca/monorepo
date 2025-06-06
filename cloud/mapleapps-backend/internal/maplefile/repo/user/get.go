// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/get.go
package user

import (
	"context"

	"github.com/gocql/gocql"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
)

func (impl userStorerImpl) GetByID(ctx context.Context, id gocql.UUID) (*dom_user.User, error) {
	return nil, nil
}

func (impl userStorerImpl) GetByEmail(ctx context.Context, email string) (*dom_user.User, error) {
	return nil, nil
}

func (impl userStorerImpl) GetByVerificationCode(ctx context.Context, verificationCode string) (*dom_user.User, error) {
	return nil, nil
}
