// repo/federateduser/update.go
package federateduser

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
)

func (r *federatedUserRepository) UpdateByID(ctx context.Context, user *dom.FederatedUser) error {
	// First, get the existing user to check what changed
	existingUser, err := r.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing user: %w", err)
	}
	if existingUser == nil {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	// Update modified timestamp
	user.ModifiedAt = time.Now()

	// Serialize data
	profileDataJSON, err := r.serializeProfileData(user.ProfileData)
	if err != nil {
		return fmt.Errorf("failed to serialize profile data: %w", err)
	}

	securityDataJSON, err := r.serializeSecurityData(user.SecurityData)
	if err != nil {
		return fmt.Errorf("failed to serialize security data: %w", err)
	}

	metadataJSON, err := r.serializeMetadata(user.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	batch := r.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	// 1. Update main table
	batch.Query(`
        UPDATE users_by_id
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
		batch.Query(`DELETE FROM users_by_email WHERE email = ?`, existingUser.Email)

		// Insert new email entry
		batch.Query(`
            INSERT INTO users_by_email (
                email, id, first_name, last_name, status, created_at
            ) VALUES (?, ?, ?, ?, ?, ?)`,
			user.Email, user.ID, user.FirstName, user.LastName,
			user.Status, user.CreatedAt,
		)
	} else {
		// Just update the existing email entry
		batch.Query(`
            UPDATE users_by_email
            SET first_name = ?, last_name = ?, status = ?
            WHERE email = ?`,
			user.FirstName, user.LastName, user.Status, user.Email,
		)
	}

	// 3. Handle status change
	if existingUser.Status != user.Status {
		// Remove from old status table
		oldCreatedDate := existingUser.CreatedAt.Format("2006-01-02")
		batch.Query(`
            DELETE FROM users_by_status_and_date
            WHERE status = ? AND created_date = ? AND created_at = ? AND id = ?`,
			existingUser.Status, oldCreatedDate, existingUser.CreatedAt, user.ID,
		)

		// Add to new status table
		newCreatedDate := user.CreatedAt.Format("2006-01-02")
		batch.Query(`
            INSERT INTO users_by_status_and_date (
                status, created_date, created_at, id, email, name, lexical_name, role
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			user.Status, newCreatedDate, user.CreatedAt, user.ID,
			user.Email, user.Name, user.LexicalName, user.Role,
		)

		// Handle active users table
		if existingUser.Status == dom.FederatedUserStatusActive {
			// Remove from active users
			batch.Query(`
                DELETE FROM active_users_by_date
                WHERE created_date = ? AND created_at = ? AND id = ?`,
				oldCreatedDate, existingUser.CreatedAt, user.ID,
			)
		}
		if user.Status == dom.FederatedUserStatusActive {
			// Add to active users
			batch.Query(`
                INSERT INTO active_users_by_date (
                    created_date, created_at, id, email, name, role
                ) VALUES (?, ?, ?, ?, ?, ?)`,
				newCreatedDate, user.CreatedAt, user.ID,
				user.Email, user.Name, user.Role,
			)
		}
	} else {
		// Just update the existing status entry
		createdDate := user.CreatedAt.Format("2006-01-02")
		batch.Query(`
            UPDATE users_by_status_and_date
            SET email = ?, name = ?, lexical_name = ?, role = ?
            WHERE status = ? AND created_date = ? AND created_at = ? AND id = ?`,
			user.Email, user.Name, user.LexicalName, user.Role,
			user.Status, createdDate, user.CreatedAt, user.ID,
		)

		if user.Status == dom.FederatedUserStatusActive {
			batch.Query(`
                UPDATE active_users_by_date
                SET email = ?, name = ?, role = ?
                WHERE created_date = ? AND created_at = ? AND id = ?`,
				user.Email, user.Name, user.Role,
				createdDate, user.CreatedAt, user.ID,
			)
		}
	}

	// 4. Handle verification code changes
	if existingUser.SecurityData != nil && existingUser.SecurityData.Code != "" {
		// Remove old code
		batch.Query(`
            DELETE FROM users_by_verification_code
            WHERE code = ? AND code_type = ?`,
			existingUser.SecurityData.Code, existingUser.SecurityData.CodeType,
		)
	}

	if user.SecurityData != nil && user.SecurityData.Code != "" {
		// Add new code with TTL
		ttl := int(time.Until(user.SecurityData.CodeExpiry).Seconds())
		if ttl > 0 {
			batch.Query(`
                INSERT INTO users_by_verification_code (
                    code, code_type, user_id, email, created_at, expires_at
                ) VALUES (?, ?, ?, ?, ?, ?) USING TTL ?`,
				user.SecurityData.Code, user.SecurityData.CodeType,
				user.ID, user.Email, time.Now(), user.SecurityData.CodeExpiry, ttl,
			)
		}
	}

	// Execute the batch
	if err := r.session.ExecuteBatch(batch); err != nil {
		r.logger.Error("Failed to update user",
			zap.String("user_id", user.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.Info("User updated successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email))

	return nil
}
