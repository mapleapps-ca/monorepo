// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/delete.go
package user

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
)

func (impl userStorerImpl) DeleteByID(ctx context.Context, id gocql.UUID) error {
	// First, get the user to know all the data we need to delete
	user, err := impl.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user for deletion: %w", err)
	}
	if user == nil {
		return nil // User doesn't exist, nothing to delete
	}

	batch := impl.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	// Delete from all tables
	batch.Query(`DELETE FROM maplefile_users_by_id WHERE id = ?`, id)
	batch.Query(`DELETE FROM maplefile_users_by_email WHERE email = ?`, user.Email)

	// Delete from status table
	// Skip

	// Delete from active users if applicable
	if user.Status == dom_user.UserStatusActive {
		// Skip
	}

	// Delete verification codes if any
	// Skip

	// Delete from search index
	// Skip

	// Execute the batch
	if err := impl.session.ExecuteBatch(batch); err != nil {
		impl.logger.Error("Failed to delete user",
			zap.String("user_id", id.String()),
			zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	impl.logger.Info("User deleted successfully",
		zap.String("user_id", id.String()),
		zap.String("email", user.Email))

	return nil
}

func (impl userStorerImpl) DeleteByEmail(ctx context.Context, email string) error {
	// First get the user by email to get the ID
	user, err := impl.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user by email for deletion: %w", err)
	}
	if user == nil {
		return nil // User doesn't exist
	}

	// Delete by ID
	return impl.DeleteByID(ctx, user.ID)
}
