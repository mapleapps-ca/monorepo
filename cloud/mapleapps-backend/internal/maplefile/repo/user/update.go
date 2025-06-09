// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/update.go
package user

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
)

func (impl userStorerImpl) UpdateByID(ctx context.Context, user *dom_user.User) error {
	// First, get the existing user to check what changed
	existingUser, err := impl.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing user: %w", err)
	}
	if existingUser == nil {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	// Update modified timestamp
	user.ModifiedAt = time.Now()

	// Serialize data
	profileDataJSON, err := impl.serializeProfileData(user.ProfileData)
	if err != nil {
		return fmt.Errorf("failed to serialize profile data: %w", err)
	}

	securityDataJSON, err := impl.serializeSecurityData(user.SecurityData)
	if err != nil {
		return fmt.Errorf("failed to serialize security data: %w", err)
	}

	metadataJSON, err := impl.serializeMetadata(user.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	batch := impl.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	// 1. Update main table
	batch.Query(`
        UPDATE maplefile_users_by_id
        SET email = ?, first_name = ?, last_name = ?, name = ?, lexical_name = ?,
            role = ?, status = ?, modified_at = ?,
            profile_data = ?, security_data = ?, metadata = ?
        WHERE id = ?`,
		user.Email, user.FirstName, user.LastName, user.Name, user.LexicalName,
		user.Role, user.Status, user.ModifiedAt,
		profileDataJSON, securityDataJSON, metadataJSON,
		user.ID,
	)

	// 2. Handle email change
	if existingUser.Email != user.Email {
		// Delete old email entry
		batch.Query(`DELETE FROM maplefile_users_by_email WHERE email = ?`, existingUser.Email)

		// Insert new email entry
		batch.Query(`
            INSERT INTO maplefile_users_by_email (
                email, id, first_name, last_name, status, created_at
            ) VALUES (?, ?, ?, ?, ?, ?)`,
			user.Email, user.ID, user.FirstName, user.LastName,
			user.Status, user.CreatedAt,
		)
	} else {
		// Just update the existing email entry
		batch.Query(`
            UPDATE maplefile_users_by_email
            SET first_name = ?, last_name = ?, status = ?
            WHERE email = ?`,
			user.FirstName, user.LastName, user.Status, user.Email,
		)
	}

	// 3. Handle status change
	if existingUser.Status != user.Status {
		// Remove from old status table
		// kip

		// Add to new status table
		// Skip

		// Handle active users table
		if existingUser.Status == dom_user.UserStatusActive {
			// Skip
		}
		if user.Status == dom_user.UserStatusActive {
			// Skip
		} else {
			// Just update the existing status entry
			// Skip

			if user.Status == dom_user.UserStatusActive {
				// Skip
			}
		}
	}

	// 4. Handle verification code changes
	// Skip

	if user.SecurityData != nil && user.SecurityData.Code != "" {
		// Skip
	}

	// Execute the batch
	if err := impl.session.ExecuteBatch(batch); err != nil {
		impl.logger.Error("Failed to update user",
			zap.String("user_id", user.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	impl.logger.Info("User updated successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email))

	return nil
}
