// github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/user/countbyfilter.go
package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/user"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type UserCountByFilterUseCase interface {
	Execute(ctx context.Context, filter *dom_user.UserFilter) (uint64, error)
}

type userCountByFilterUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.Repository
}

func NewUserCountByFilterUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_user.Repository,
) UserCountByFilterUseCase {
	return &userCountByFilterUseCaseImpl{config, logger, repo}
}

func (uc *userCountByFilterUseCaseImpl) Execute(ctx context.Context, filter *dom_user.UserFilter) (uint64, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if filter == nil {
		e["filter"] = "User filter is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating user count by filter",
			zap.Any("error", e))
		return 0, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Count in database.
	//

	uc.logger.Debug("Counting users by filter",
		zap.Any("role", filter.Role),
		zap.Any("status", filter.Status))

	return uc.repo.CountByFilter(ctx, filter)
}
