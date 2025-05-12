package bannedipaddress

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_banip "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/bannedipaddress"
)

type BannedIPAddressListAllValuesUseCase interface {
	Execute(ctx context.Context) ([]string, error)
}

type bannedIPAddressListAllValuesUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_banip.Repository
}

func NewBannedIPAddressListAllValuesUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_banip.Repository,
) BannedIPAddressListAllValuesUseCase {
	return &bannedIPAddressListAllValuesUseCaseImpl{config, logger, repo}
}

func (uc *bannedIPAddressListAllValuesUseCaseImpl) Execute(ctx context.Context) ([]string, error) {
	return uc.repo.ListAllValues(ctx)
}
