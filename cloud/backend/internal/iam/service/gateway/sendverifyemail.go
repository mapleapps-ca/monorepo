package gateway

import (
	"context"

	uc_emailer "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/emailer"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GatewaySendVerifyEmailService interface {
	Execute(sessCtx context.Context, req *GatewaySendVerifyEmailRequestIDO) error
}

type gatewaySendVerifyEmailServiceImpl struct {
	userGetByEmailUseCase                     uc_user.FederatedUserGetByEmailUseCase
	sendFederatedUserVerificationEmailUseCase uc_emailer.SendFederatedUserVerificationEmailUseCase
}

func NewGatewaySendVerifyEmailService(
	uc1 uc_user.FederatedUserGetByEmailUseCase,
	uc2 uc_emailer.SendFederatedUserVerificationEmailUseCase,
) GatewaySendVerifyEmailService {
	return &gatewaySendVerifyEmailServiceImpl{uc1, uc2}
}

type GatewaySendVerifyEmailRequestIDO struct {
	Email string `json:"email"`

	// Module refers to which module the user is registering for.
	Module int `json:"module,omitempty"`
}

func (s *gatewaySendVerifyEmailServiceImpl) Execute(sessCtx context.Context, req *GatewaySendVerifyEmailRequestIDO) error {
	// Extract from our session the following data.
	// sessionID := sessCtx.Value(constants.SessionID).(string)
	//
	// //
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
		return httperror.NewForBadRequest(&e)
	}

	// Lookup the user in our database, else return a `400 Bad Request` error.
	u, err := s.userGetByEmailUseCase.Execute(sessCtx, req.Email)
	if err != nil {
		return err
	}
	if u == nil {
		return httperror.NewForBadRequestWithSingleField("email", "does not exist")
	}

	if err := s.sendFederatedUserVerificationEmailUseCase.Execute(context.Background(), req.Module, u); err != nil {
		// Skip any error handling...
	}

	return nil
}
