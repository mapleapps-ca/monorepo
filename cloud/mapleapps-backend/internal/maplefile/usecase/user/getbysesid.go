package user

import (
	"context"
	"encoding/json"
	"errors"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/cassandracache"
)

type UserGetBySessionIDUseCase interface {
	Execute(ctx context.Context, sessionID string) (*dom_user.User, error)
}

type userGetBySessionIDUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	cache  cassandracache.CassandraCacher
}

func NewUserGetBySessionIDUseCase(config *config.Configuration, logger *zap.Logger, ca cassandracache.CassandraCacher) UserGetBySessionIDUseCase {
	logger = logger.Named("UserGetBySessionIDUseCase")
	return &userGetBySessionIDUseCaseImpl{config, logger, ca}
}

func (uc *userGetBySessionIDUseCaseImpl) Execute(ctx context.Context, sessionID string) (*dom_user.User, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if sessionID == "" {
		e["session_id"] = "missing value"
	} else {
		//TODO: IMPL.
	}
	if len(e) != 0 {
		uc.logger.Warn("Validation failed for upsert",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2:
	//

	userBytes, err := uc.cache.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if userBytes == nil {
		uc.logger.Warn("record not found")
		return nil, errors.New("record not found")
	}
	var user dom_user.User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		uc.logger.Error("unmarshalling failed", zap.Any("err", err))
		return nil, err
	}

	return &user, nil
}
