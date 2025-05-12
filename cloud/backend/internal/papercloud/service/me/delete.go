// github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/service/me/delete.go
package me

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/user"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/usecase/user"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/password"
	sstring "github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/securestring"
)

type DeleteMeRequestDTO struct {
	Password string `json:"password"`
}

type DeleteMeService interface {
	Execute(sessCtx context.Context, req *DeleteMeRequestDTO) error
}

type deleteMeServiceImpl struct {
	config                *config.Configuration
	logger                *zap.Logger
	passwordProvider      password.Provider
	userGetByIDUseCase    uc_user.UserGetByIDUseCase
	userDeleteByIDUseCase uc_user.UserDeleteByIDUseCase
}

func NewDeleteMeService(
	config *config.Configuration,
	logger *zap.Logger,
	passwordProvider password.Provider,
	userGetByIDUseCase uc_user.UserGetByIDUseCase,
	userDeleteByIDUseCase uc_user.UserDeleteByIDUseCase,
) DeleteMeService {
	return &deleteMeServiceImpl{
		config:                config,
		logger:                logger,
		passwordProvider:      passwordProvider,
		userGetByIDUseCase:    userGetByIDUseCase,
		userDeleteByIDUseCase: userDeleteByIDUseCase,
	}
}

func (svc *deleteMeServiceImpl) Execute(sessCtx context.Context, req *DeleteMeRequestDTO) error {
	//
	// STEP 1: Validation
	//

	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return httperror.NewForBadRequestWithSingleField("non_field_error", "Password is required")
	}

	e := make(map[string]string)
	if req.Password == "" {
		e["password"] = "Password is required"
	}
	if len(e) != 0 {
		svc.logger.Warn("Failed validation",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get required from context.
	//

	sessionUserID, ok := sessCtx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting local user id",
			zap.Any("error", "Not found in context: user_id"))
		return errors.New("user id not found in context")
	}

	// Defend against admin deleting themselves
	sessionUserRole, _ := sessCtx.Value(constants.SessionFederatedUserRole).(int8)
	if sessionUserRole == dom_user.UserRoleRoot {
		svc.logger.Warn("admin is not allowed to delete themselves",
			zap.Any("error", ""))
		return httperror.NewForForbiddenWithSingleField("message", "admins do not have permission to delete themselves")
	}

	//
	// STEP 3: Get user from database.
	//

	user, err := svc.userGetByIDUseCase.Execute(sessCtx, sessionUserID)
	if err != nil {
		svc.logger.Error("Failed getting user", zap.Any("error", err))
		return err
	}
	if user == nil {
		errMsg := "User does not exist"
		svc.logger.Error(errMsg, zap.Any("user_id", sessionUserID))
		return httperror.NewForBadRequestWithSingleField("message", errMsg)
	}

	//
	// STEP 4: Verify password.
	//

	securePassword, err := sstring.NewSecureString(req.Password)
	if err != nil {
		svc.logger.Error("Failed to create secure string", zap.Any("error", err))
		return err
	}
	defer securePassword.Wipe()

	passwordMatch, _ := svc.passwordProvider.ComparePasswordAndHash(securePassword, user.PasswordHash)
	if !passwordMatch {
		svc.logger.Warn("Password verification failed")
		return httperror.NewForBadRequestWithSingleField("password", "Incorrect password")
	}

	//
	// STEP 5: Delete user.
	//

	err = svc.userDeleteByIDUseCase.Execute(sessCtx, sessionUserID)
	if err != nil {
		svc.logger.Error("Failed to delete user", zap.Any("error", err))
		return err
	}

	svc.logger.Info("User successfully deleted", zap.Any("user_id", sessionUserID))
	return nil
}
