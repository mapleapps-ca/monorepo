package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	uc_emailer "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/emailer"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/random"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/password"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/cassandracache"
)

type GatewayForgotPasswordService interface {
	Execute(sessCtx context.Context, req *GatewayForgotPasswordRequestIDO) (*GatewayForgotPasswordResponseIDO, error)
}

type gatewayForgotPasswordServiceImpl struct {
	passwordProvider                           password.PasswordProvider
	cache                                      cassandracache.CassandraCacher
	jwtProvider                                jwt.JWTProvider
	userGetByEmailUseCase                      uc_user.FederatedUserGetByEmailUseCase
	userUpdateUseCase                          uc_user.FederatedUserUpdateUseCase
	sendFederatedUserPasswordResetEmailUseCase uc_emailer.SendFederatedUserPasswordResetEmailUseCase
}

func NewGatewayForgotPasswordService(
	pp password.PasswordProvider,
	cach cassandracache.CassandraCacher,
	jwtp jwt.JWTProvider,
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

	u.SecurityData.Code = fmt.Sprintf("%s", passwordResetVerificationCode)
	u.SecurityData.CodeExpiry = time.Now().Add(5 * time.Minute)
	u.Metadata.ModifiedAt = time.Now()
	u.Metadata.ModifiedByName = u.Name
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
