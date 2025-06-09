// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/get.go
package user

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
)

func (impl userStorerImpl) GetByID(ctx context.Context, id gocql.UUID) (*dom_user.User, error) {
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
        FROM maplefile_users_by_id
        WHERE id = ?`

	err := impl.session.Query(query, id).WithContext(ctx).Scan(
		&email, &firstName, &lastName, &name, &lexicalName,
		&role, &status, &timezone, &createdAt, &modifiedAt,
		&profileData, &securityData, &metadata,
	)

	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		impl.logger.Error("Failed to get user by ID",
			zap.String("user_id", id.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Construct the user object
	user := &dom_user.User{
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
	if err := impl.deserializeUserData(profileData, securityData, metadata, user); err != nil {
		impl.logger.Error("Failed to deserialize user data",
			zap.String("user_id", id.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize user data: %w", err)
	}

	return user, nil
}

func (impl userStorerImpl) GetByEmail(ctx context.Context, email string) (*dom_user.User, error) {
	var (
		id                                     gocql.UUID
		emailResult                            string
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
        FROM maplefile_users_by_email
        WHERE email = ?`

	err := impl.session.Query(query, email).WithContext(ctx).Scan(
		&id, &emailResult, &firstName, &lastName, &name, &lexicalName, // 🔧 FIXED: Use emailResult variable
		&role, &status, &timezone, &createdAt, &modifiedAt,
		&profileData, &securityData, &metadata,
	)

	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		impl.logger.Error("Failed to get user by Email",
			zap.String("user_email", email),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Construct the user object
	user := &dom_user.User{
		ID:          id,
		Email:       emailResult,
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
	if err := impl.deserializeUserData(profileData, securityData, metadata, user); err != nil {
		impl.logger.Error("Failed to deserialize user data",
			zap.String("user_id", id.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize user data: %w", err)
	}

	return user, nil
}

func (impl userStorerImpl) GetByVerificationCode(ctx context.Context, verificationCode string) (*dom_user.User, error) {
	return nil, nil
}
