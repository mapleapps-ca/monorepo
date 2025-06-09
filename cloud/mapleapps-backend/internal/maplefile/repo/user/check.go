// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/check.go
package user

import (
	"context"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

func (r *userStorerImpl) CheckIfExistsByID(ctx context.Context, id gocql.UUID) (bool, error) {
	query := `SELECT id FROM maplefile_users_by_id WHERE id = ? LIMIT 1`
	err := r.session.Query(query, id).WithContext(ctx).Scan(&id)

	if err == gocql.ErrNotFound {
		return false, nil
	}
	if err != nil {
		r.logger.Error("Failed to check if user exists by id",
			zap.String("id", id.String()),
			zap.Error(err))
		return false, err
	}

	return true, nil
}

func (r *userStorerImpl) CheckIfExistsByEmail(ctx context.Context, email string) (bool, error) {
	var id gocql.UUID

	query := `SELECT id FROM maplefile_users_by_email WHERE email = ? LIMIT 1`
	err := r.session.Query(query, email).WithContext(ctx).Scan(&id)

	if err == gocql.ErrNotFound {
		return false, nil
	}
	if err != nil {
		r.logger.Error("Failed to check if user exists by email",
			zap.String("email", email),
			zap.Error(err))
		return false, err
	}

	return true, nil
}
