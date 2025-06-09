// repo/federateduser/delete.go
package federateduser

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
)

func (r *federatedUserRepository) DeleteByID(ctx context.Context, id gocql.UUID) error {
	// First, get the user to know all the data we need to delete
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user for deletion: %w", err)
	}
	if user == nil {
		return nil // User doesn't exist, nothing to delete
	}

	batch := r.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	// Delete from all tables
	batch.Query(`DELETE FROM iam_federated_users_by_id WHERE id = ?`, id)
	batch.Query(`DELETE FROM iam_federated_users_by_email WHERE email = ?`, user.Email)

	// Delete from status table
	// Skip

	// Delete from active users if applicable
	if user.Status == dom.FederatedUserStatusActive {
		// Skip
	}

	// Delete verification codes if any
	if user.SecurityData != nil && user.SecurityData.Code != "" {
		batch.Query(`
            DELETE FROM iam_federated_users_by_verification_code
            WHERE code = ? AND code_type = ?`,
			user.SecurityData.Code, user.SecurityData.CodeType,
		)
	}

	// Delete from search index
	// Skip

	// Execute the batch
	if err := r.session.ExecuteBatch(batch); err != nil {
		r.logger.Error("Failed to delete user",
			zap.String("user_id", id.String()),
			zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	r.logger.Info("User deleted successfully",
		zap.String("user_id", id.String()),
		zap.String("email", user.Email))

	return nil
}

func (r *federatedUserRepository) DeleteByEmail(ctx context.Context, email string) error {
	// First get the user by email to get the ID
	user, err := r.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user by email for deletion: %w", err)
	}
	if user == nil {
		return nil // User doesn't exist
	}

	// Delete by ID
	return r.DeleteByID(ctx, user.ID)
}
