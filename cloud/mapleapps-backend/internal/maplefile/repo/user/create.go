// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/create.go
package user

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
)

func (impl userStorerImpl) Create(ctx context.Context, user *dom_user.User) error {
	// Ensure we have a valid UUID
	if user.ID == (gocql.UUID{}) {
		user.ID = gocql.TimeUUID()
	}

	// Set timestamps if not set
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.ModifiedAt.IsZero() {
		user.ModifiedAt = now
	}

	// Serialize complex data to JSON
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

	// Use a batch for atomic writes across multiple tables
	batch := impl.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	// 1. Insert into users_by_id (primary table)
	batch.Query(`
        INSERT INTO maplefile_users_by_id (
            id, email, first_name, last_name, name, lexical_name,
            role, status, timezone, created_at, modified_at,
            profile_data, security_data, metadata
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.FirstName, user.LastName, user.Name, user.LexicalName,
		user.Role, user.Status, user.Timezone, user.CreatedAt, user.ModifiedAt,
		profileDataJSON, securityDataJSON, metadataJSON,
	)

	// 2. Insert into users_by_email
	batch.Query(`
    INSERT INTO maplefile_users_by_email (
        email, id,  first_name, last_name, name, lexical_name,
        role, status, timezone, created_at, modified_at,
        profile_data, security_data, metadata
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Email, user.ID, user.FirstName, user.LastName, user.Name, user.LexicalName,
		user.Role, user.Status, user.Timezone, user.CreatedAt, user.ModifiedAt,
		profileDataJSON, securityDataJSON, metadataJSON,
	)

	// 3. Insert into users_by_status_and_date for listing
	// Skip

	// 4. If status is active, also insert into active users table
	if user.Status == dom_user.UserStatusActive {
		// Skip
	}

	// 5. Handle verification codes with TTL
	// Skip

	// 6. Add to search index (simplified - you might want to use external search)
	if user.Name != "" || user.Email != "" {
		// Skip
	}

	// Execute the batch
	if err := impl.session.ExecuteBatch(batch); err != nil {
		impl.logger.Error("Failed to create user",
			zap.String("user_id", user.ID.String()),
			zap.String("email", user.Email),
			zap.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}

	impl.logger.Info("User created successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email))

	return nil
}
