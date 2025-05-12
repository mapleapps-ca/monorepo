// github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/federateduser/countbyfilter.go
package federateduser

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type FederatedUserCountByFilterUseCase interface {
	Execute(ctx context.Context, filter *dom_user.FederatedUserFilter) (uint64, error)
}

type userCountByFilterUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_user.Repository
}

func NewFederatedUserCountByFilterUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_user.Repository,
) FederatedUserCountByFilterUseCase {
	return &userCountByFilterUseCaseImpl{config, logger, repo}
}

func (uc *userCountByFilterUseCaseImpl) Execute(ctx context.Context, filter *dom_user.FederatedUserFilter) (uint64, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if filter == nil {
		e["filter"] = "FederatedUser filter is required"
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
