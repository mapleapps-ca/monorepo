package gateway

import (
	"context"
	"time"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	domain "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GatewayVerifyEmailService interface {
	Execute(sessCtx context.Context, req *GatewayVerifyEmailRequestIDO) (*GatwayVerifyEmailResponseIDO, error)
}

type gatewayVerifyEmailServiceImpl struct {
	userGetByVerificationCodeUseCase uc_user.FederatedUserGetByVerificationCodeUseCase
	userUpdateUseCase                uc_user.FederatedUserUpdateUseCase
}

func NewGatewayVerifyEmailService(
	uc1 uc_user.FederatedUserGetByVerificationCodeUseCase,
	uc2 uc_user.FederatedUserUpdateUseCase,
) GatewayVerifyEmailService {
	return &gatewayVerifyEmailServiceImpl{uc1, uc2}
}

type GatewayVerifyEmailRequestIDO struct {
	Code string `json:"code"`
}

type GatwayVerifyEmailResponseIDO struct {
	Message           string `json:"message"`
	FederatedUserRole int8   `bson:"user_role" json:"user_role"`
}

func (s *gatewayVerifyEmailServiceImpl) Execute(sessCtx context.Context, req *GatewayVerifyEmailRequestIDO) (*GatwayVerifyEmailResponseIDO, error) {
	// Extract from our session the following data.
	// sessionID := sessCtx.Value(constants.SessionID).(string)

	res := &GatwayVerifyEmailResponseIDO{}

	// Lookup the user in our database, else return a `400 Bad Request` error.
	u, err := s.userGetByVerificationCodeUseCase.Execute(sessCtx, req.Code)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, httperror.NewForBadRequestWithSingleField("code", "does not exist")
	}

	//TODO: Handle expiry dates.

	// Extract from our session the following data.
	// userID := sessCtx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	ipAddress, _ := sessCtx.Value(constants.SessionIPAddress).(string)

	// Verify the user.
	u.SecurityData.WasEmailVerified = true
	// ou.ModifiedByFederatedUserID = userID
	u.ModifiedAt = time.Now()
	// ou.ModifiedByName = fmt.Sprintf("%s %s", ou.FirstName, ou.LastName)
	u.Metadata.ModifiedAt = u.ModifiedAt
	u.Metadata.ModifiedFromIPAddress = ipAddress
	if err := s.userUpdateUseCase.Execute(sessCtx, u); err != nil {
		return nil, err
	}

	//
	// Send notification based on user role
	//

	switch u.Role {
	case domain.FederatedUserRoleIndividual:
		{
			res.Message = "Thank you for verifying. You may log in now to get started!"
			break
		}
	default:
		{
			res.Message = "Thank you for verifying. You may log in now to get started!"
			break
		}
	}
	res.FederatedUserRole = u.Role

	return res, nil
}
