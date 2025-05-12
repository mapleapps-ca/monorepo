package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	uc_emailer "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/emailer"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/random"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/password"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/database/mongodbcache"
)

type GatewayForgotPasswordService interface {
	Execute(sessCtx context.Context, req *GatewayForgotPasswordRequestIDO) (*GatewayForgotPasswordResponseIDO, error)
}

type gatewayForgotPasswordServiceImpl struct {
	passwordProvider                           password.Provider
	cache                                      mongodbcache.Cacher
	jwtProvider                                jwt.Provider
	userGetByEmailUseCase                      uc_user.FederatedUserGetByEmailUseCase
	userUpdateUseCase                          uc_user.FederatedUserUpdateUseCase
	sendFederatedUserPasswordResetEmailUseCase uc_emailer.SendFederatedUserPasswordResetEmailUseCase
}

func NewGatewayForgotPasswordService(
	pp password.Provider,
	cach mongodbcache.Cacher,
	jwtp jwt.Provider,
	uc1 uc_user.FederatedUserGetByEmailUseCase,
	uc2 uc_user.FederatedUserUpdateUseCase,
	uc3 uc_emailer.SendFederatedUserPasswordResetEmailUseCase,
) GatewayForgotPasswordService {
	return &gatewayForgotPasswordServiceImpl{pp, cach, jwtp, uc1, uc2, uc3}
}

type GatewayForgotPasswordRequestIDO struct {
	Email string `json:"email"`

	// Module refers to which module the user is reqesting this for.
	Module int `json:"module,omitempty"`
}

type GatewayForgotPasswordResponseIDO struct {
	Message string `json:"message"`
}

func (s *gatewayForgotPasswordServiceImpl) Execute(sessCtx context.Context, req *GatewayForgotPasswordRequestIDO) (*GatewayForgotPasswordResponseIDO, error) {
	//
	// STEP 1: Sanization of input.
	//

	// Defensive Code: For security purposes we need to perform some sanitization on the inputs.
	req.Email = strings.ToLower(req.Email)
	req.Email = strings.ReplaceAll(req.Email, " ", "")
	req.Email = strings.ReplaceAll(req.Email, "\t", "")
	req.Email = strings.TrimSpace(req.Email)

	//
	// STEP 2: Validation of input.
	//

	e := make(map[string]string)
	if req.Email == "" {
		e["email"] = "Email address is required"
	}
	if req.Module == 0 {
		e["module"] = "Module is required"
	}

	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 3:
	//

	// Lookup the federateduser in our database, else return a `400 Bad Request` error.
	u, err := s.userGetByEmailUseCase.Execute(sessCtx, req.Email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, httperror.NewForBadRequestWithSingleField("email", "Email address does not exist")
	}

	//
	// STEP 4:
	//

	passwordResetVerificationCode, err := random.GenerateSixDigitCode()
	if err != nil {
		return nil, err
	}

	u.PasswordResetVerificationCode = fmt.Sprintf("%s", passwordResetVerificationCode)
	u.PasswordResetVerificationExpiry = time.Now().Add(5 * time.Minute)
	u.ModifiedAt = time.Now()
	u.ModifiedByName = u.Name
	err = s.userUpdateUseCase.Execute(sessCtx, u)
	if err != nil {
		return nil, err
	}

	//
	// STEP 5: Send email
	//

	if err := s.sendFederatedUserPasswordResetEmailUseCase.Execute(sessCtx, req.Module, u); err != nil {
		// Skip any error handling...
	}

	//
	// STEP X: Done
	//

	// Return our auth keys.
	return &GatewayForgotPasswordResponseIDO{
		Message: "Password reset email has been sent",
	}, nil
}
