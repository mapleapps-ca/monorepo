// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/delete.go
package user

import (
	"context"

	"github.com/gocql/gocql"
)

func (impl userStorerImpl) DeleteByID(ctx context.Context, id gocql.UUID) error {

	return nil
}

func (impl userStorerImpl) DeleteByEmail(ctx context.Context, email string) error {

	return nil
}
