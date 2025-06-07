// repo/federateduser/get.go
package federateduser

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
)

func (r *federatedUserRepository) GetByID(ctx context.Context, id gocql.UUID) (*dom.FederatedUser, error) {
	var (
		email, firstName, lastName, name, lexicalName string
		role, status                                  int8
		timezone                                      string
		createdAt, modifiedAt                         time.Time
		profileData, securityData, metadata           string
	)

	query := `
        SELECT email, first_name, last_name, name, lexical_name,
               role, status, timezone, created_at, modified_at,
               profile_data, security_data, metadata
        FROM users_by_id
        WHERE id = ?`

	err := r.session.Query(query, id).WithContext(ctx).Scan(
		&email, &firstName, &lastName, &name, &lexicalName,
		&role, &status, &timezone, &createdAt, &modifiedAt,
		&profileData, &securityData, &metadata,
	)

	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("Failed to get user by ID",
			zap.String("user_id", id.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Construct the user object
	user := &dom.FederatedUser{
		ID:          id,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Name:        name,
		LexicalName: lexicalName,
		Role:        role,
		Status:      status,
		Timezone:    timezone,
		CreatedAt:   createdAt,
		ModifiedAt:  modifiedAt,
	}

	// Deserialize JSON fields
	if err := r.deserializeUserData(profileData, securityData, metadata, user); err != nil {
		r.logger.Error("Failed to deserialize user data",
			zap.String("user_id", id.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize user data: %w", err)
	}

	return user, nil
}

func (r *federatedUserRepository) GetByEmail(ctx context.Context, email string) (*dom.FederatedUser, error) {
	var (
		id                                     gocql.UUID
		firstName, lastName, name, lexicalName string
		role, status                           int8
		timezone                               string
		createdAt, modifiedAt                  time.Time
		profileData, securityData, metadata    string
	)

	query := `
        SELECT id, email, first_name, last_name, name, lexical_name,
               role, status, timezone, created_at, modified_at,
               profile_data, security_data, metadata
        FROM users_by_email
        WHERE email = ?`

	err := r.session.Query(query, email).WithContext(ctx).Scan(
		&id, &email, &firstName, &lastName, &name, &lexicalName,
		&role, &status, &timezone, &createdAt, &modifiedAt,
		&profileData, &securityData, &metadata,
	)

	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("Failed to get user by Email",
			zap.String("user_email", email),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Construct the user object
	user := &dom.FederatedUser{
		ID:          id,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Name:        name,
		LexicalName: lexicalName,
		Role:        role,
		Status:      status,
		Timezone:    timezone,
		CreatedAt:   createdAt,
		ModifiedAt:  modifiedAt,
	}

	// Deserialize JSON fields
	if err := r.deserializeUserData(profileData, securityData, metadata, user); err != nil {
		r.logger.Error("Failed to deserialize user data",
			zap.String("user_id", id.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize user data: %w", err)
	}

	return user, nil
}

func (r *federatedUserRepository) GetByVerificationCode(ctx context.Context, verificationCode string) (*dom.FederatedUser, error) {
	// We need to check both code types (email_verification and password_reset)
	codeTypes := []string{"email_verification", "password_reset"}

	for _, codeType := range codeTypes {
		var userID gocql.UUID
		var expiresAt time.Time

		query := `
            SELECT user_id, expires_at
            FROM users_by_verification_code
            WHERE code = ? AND code_type = ?`

		err := r.session.Query(query, verificationCode, codeType).
			WithContext(ctx).Scan(&userID, &expiresAt)

		if err == gocql.ErrNotFound {
			continue // Try next code type
		}
		if err != nil {
			r.logger.Error("Failed to get user by verification code",
				zap.String("code_type", codeType),
				zap.Error(err))
			return nil, fmt.Errorf("failed to get user by verification code: %w", err)
		}

		// Check if code is expired
		if time.Now().After(expiresAt) {
			r.logger.Warn("Verification code expired",
				zap.String("user_id", userID.String()),
				zap.String("code_type", codeType))
			return nil, nil
		}

		// Get the full user data
		user, err := r.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		// Set the verification code data in the security data
		if user != nil && user.SecurityData != nil {
			user.SecurityData.Code = verificationCode
			user.SecurityData.CodeType = codeType
			user.SecurityData.CodeExpiry = expiresAt
		}

		return user, nil
	}

	return nil, nil // Code not found in any table
}
