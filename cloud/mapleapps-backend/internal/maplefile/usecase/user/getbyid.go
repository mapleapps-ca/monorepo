// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/user/getbyid.go
package user

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type UserGetByIDUseCase interface {
	Execute(ctx context.Context, id gocql.UUID) (*dom_user.User, error)
}

type userGetByIDUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.Repository
}

func NewUserGetByIDUseCase(config *config.Configuration, logger *zap.Logger, repo dom_user.Repository) UserGetByIDUseCase {
	logger = logger.Named("UserGetByIDUseCase")

	// Defensive check: ensure dependencies are not nil
	if config == nil {
		panic("config cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}
	if repo == nil {
		panic("repository cannot be nil")
	}

	return &userGetByIDUseCaseImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (uc *userGetByIDUseCaseImpl) Execute(ctx context.Context, id gocql.UUID) (*dom_user.User, error) {
	// Defensive check: ensure use case was properly initialized
	if uc.repo == nil {
		uc.logger.Error("repository is nil - use case was not properly initialized")
		return nil, errors.New("internal error: repository not available")
	}

	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.String() == "" {
		e["id"] = "missing value"
	}
	if len(e) != 0 {
		uc.logger.Warn("Validation failed for get by ID",
			zap.Any("error", e),
			zap.String("id", id.String()))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	uc.logger.Debug("Getting user by ID",
		zap.String("user_id", id.String()))

	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get user from repository",
			zap.String("user_id", id.String()),
			zap.Any("error", err))
		return nil, err
	}

	if user != nil {
		uc.logger.Debug("Successfully retrieved user",
			zap.String("user_id", id.String()),
			zap.String("email", user.Email))
	} else {
		uc.logger.Debug("User not found",
			zap.String("user_id", id.String()))
	}

	return user, nil
}
