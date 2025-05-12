// github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/service/me/delete.go
package me

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/federateduser"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/federateduser"
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
	userGetByIDUseCase    uc_user.FederatedUserGetByIDUseCase
	userDeleteByIDUseCase uc_user.FederatedUserDeleteByIDUseCase
}

func NewDeleteMeService(
	config *config.Configuration,
	logger *zap.Logger,
	passwordProvider password.Provider,
	userGetByIDUseCase uc_user.FederatedUserGetByIDUseCase,
	userDeleteByIDUseCase uc_user.FederatedUserDeleteByIDUseCase,
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

	sessionFederatedUserID, ok := sessCtx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting local federateduser id",
			zap.Any("error", "Not found in context: user_id"))
		return errors.New("federateduser id not found in context")
	}

	// Defend against admin deleting themselves
	sessionFederatedUserRole, _ := sessCtx.Value(constants.SessionFederatedUserRole).(int8)
	if sessionFederatedUserRole == dom_user.FederatedUserRoleRoot {
		svc.logger.Warn("admin is not allowed to delete themselves",
			zap.Any("error", ""))
		return httperror.NewForForbiddenWithSingleField("message", "admins do not have permission to delete themselves")
	}

	//
	// STEP 3: Get federateduser from database.
	//

	federateduser, err := svc.userGetByIDUseCase.Execute(sessCtx, sessionFederatedUserID)
	if err != nil {
		svc.logger.Error("Failed getting federateduser", zap.Any("error", err))
		return err
	}
	if federateduser == nil {
		errMsg := "FederatedUser does not exist"
		svc.logger.Error(errMsg, zap.Any("user_id", sessionFederatedUserID))
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

	// passwordMatch, _ := svc.passwordProvider.ComparePasswordAndHash(securePassword, federateduser.PasswordHash)
	// if !passwordMatch {
	// 	svc.logger.Warn("Password verification failed")
	// 	return httperror.NewForBadRequestWithSingleField("password", "Incorrect password")
	// }

	//
	// STEP 5: Delete federateduser.
	//

	err = svc.userDeleteByIDUseCase.Execute(sessCtx, sessionFederatedUserID)
	if err != nil {
		svc.logger.Error("Failed to delete federateduser", zap.Any("error", err))
		return err
	}

	svc.logger.Info("FederatedUser successfully deleted", zap.Any("user_id", sessionFederatedUserID))
	return nil
}
