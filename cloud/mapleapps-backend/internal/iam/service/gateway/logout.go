package gateway

import (
	"context"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/cassandracache"
)

type GatewayLogoutService interface {
	Execute(ctx context.Context) error
}

type gatewayLogoutServiceImpl struct {
	cache cassandracache.Cacher
}

func NewGatewayLogoutService(
	cach cassandracache.Cacher,
) GatewayLogoutService {
	return &gatewayLogoutServiceImpl{cach}
}

func (s *gatewayLogoutServiceImpl) Execute(ctx context.Context) error {
	// Extract from our session the following data.
	sessionID, ok := ctx.Value(constants.SessionID).(string)
	if !ok {
		return httperror.NewForBadRequestWithSingleField("session_id", "not logged in")
	}

	if err := s.cache.Delete(ctx, sessionID); err != nil {
		return err
	}
	return nil
}
