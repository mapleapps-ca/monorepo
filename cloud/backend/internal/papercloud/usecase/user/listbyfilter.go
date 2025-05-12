// github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/usecase/user/listbyfilter.go
package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/user"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type UserListByFilterUseCase interface {
	Execute(ctx context.Context, filter *dom_user.UserFilter) (*dom_user.UserFilterResult, error)
}

type userListByFilterUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.Repository
}

func NewUserListByFilterUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_user.Repository,
) UserListByFilterUseCase {
	return &userListByFilterUseCaseImpl{config, logger, repo}
}

func (uc *userListByFilterUseCaseImpl) Execute(ctx context.Context, filter *dom_user.UserFilter) (*dom_user.UserFilterResult, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if filter == nil {
		e["filter"] = "User filter is required"
	} else {
		// Validate limit to prevent excessive data loads
		if filter.Limit > 1000 {
			filter.Limit = 1000
		}
	}

	if len(e) != 0 {
		uc.logger.Warn("Failed validating user list by filter",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: List from database.
	//

	uc.logger.Debug("Listing users by filter",
		zap.Any("role", filter.Role),
		zap.Any("status", filter.Status),
		zap.Any("limit", filter.Limit))

	return uc.repo.ListByFilter(ctx, filter)
}
