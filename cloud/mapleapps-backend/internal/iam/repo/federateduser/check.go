// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/repo/federateduser/check.go
package federateduser

import (
	"context"

	"github.com/gocql/gocql"
)

func (impl userStorerImpl) CheckIfExistsByID(ctx context.Context, id gocql.UUID) (bool, error) {
	return false, nil
}

func (impl userStorerImpl) CheckIfExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}
