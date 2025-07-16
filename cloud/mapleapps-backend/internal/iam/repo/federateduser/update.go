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
        UPDATE iam_federated_users_by_id
        SET email = ?, first_name = ?, last_name = ?, name = ?, lexical_name = ?,
            role = ?, status = ?, modified_at = ?,
            profile_data = ?, security_data = ?, metadata = ?,
            user_plan = ?, storage_limit_bytes = ?, storage_used_bytes = ?
        WHERE id = ?`,
		user.Email, user.FirstName, user.LastName, user.Name, user.LexicalName,
		user.Role, user.Status, user.ModifiedAt,
		profileDataJSON, securityDataJSON, metadataJSON,
		user.UserPlan, user.StorageLimitBytes, user.StorageUsedBytes,
		user.ID,
	)

	// 2. Handle email change
	if existingUser.Email != user.Email {
		// Delete old email entry
		batch.Query(`DELETE FROM iam_federated_users_by_email WHERE email = ?`, existingUser.Email)

		// Insert new email entry
		batch.Query(`
            INSERT INTO iam_federated_users_by_email (
                email, id, first_name, last_name, name, lexical_name,
                role, status, timezone, created_at, modified_at,
                profile_data, security_data, metadata,
                user_plan, storage_limit_bytes, storage_used_bytes
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			user.Email, user.ID, user.FirstName, user.LastName, user.Name, user.LexicalName,
			user.Role, user.Status, user.Timezone, user.CreatedAt, user.ModifiedAt,
			profileDataJSON, securityDataJSON, metadataJSON,
			user.UserPlan, user.StorageLimitBytes, user.StorageUsedBytes,
		)
	} else {
		// Just update the existing email entry
		batch.Query(`
            UPDATE iam_federated_users_by_email
            SET first_name = ?, last_name = ?, name = ?, lexical_name = ?,
                role = ?, status = ?, modified_at = ?,
                profile_data = ?, security_data = ?, metadata = ?,
                user_plan = ?, storage_limit_bytes = ?, storage_used_bytes = ?
            WHERE email = ?`,
			user.FirstName, user.LastName, user.Name, user.LexicalName,
			user.Role, user.Status, user.ModifiedAt,
			profileDataJSON, securityDataJSON, metadataJSON,
			user.UserPlan, user.StorageLimitBytes, user.StorageUsedBytes,
			user.Email,
		)
	}

	// 3. Handle status change
	if existingUser.Status != user.Status {
		// Remove from old status table
		// Skip

		// Add to new status table
		// Skip

		// Handle active users table
		if existingUser.Status == dom.FederatedUserStatusActive {
			// Skip
		}
		if user.Status == dom.FederatedUserStatusActive {
			// Skip
		} else {
			// Just update the existing status entry
			// Skip

			if user.Status == dom.FederatedUserStatusActive {
				// Skip
			}
		}
	}

	// 4. Handle verification code changes
	if existingUser.SecurityData != nil && existingUser.SecurityData.Code != "" {
		// Remove old code
		batch.Query(`
            DELETE FROM iam_federated_users_by_verification_code
            WHERE code = ? AND code_type = ?`,
			existingUser.SecurityData.Code, existingUser.SecurityData.CodeType,
		)
	}

	if user.SecurityData != nil && user.SecurityData.Code != "" {
		// Add new code with TTL
		ttl := int(time.Until(user.SecurityData.CodeExpiry).Seconds())
		if ttl > 0 {
			batch.Query(`
                INSERT INTO iam_federated_users_by_verification_code (
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
