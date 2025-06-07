// repo/federateduser/check.go
package federateduser

import (
	"context"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

func (r *federatedUserRepository) CheckIfExistsByEmail(ctx context.Context, email string) (bool, error) {
	var id gocql.UUID

	query := `SELECT id FROM users_by_email WHERE email = ? LIMIT 1`
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
