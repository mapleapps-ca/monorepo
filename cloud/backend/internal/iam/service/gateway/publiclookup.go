package gateway

import (
	"context"
	"encoding/base64"
	"strings"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/password"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/database/mongodbcache"
)

type GatewayFederatedUserPublicLookupRequestDTO struct {
	Email string `json:"email"`
}

type GatewayFederatedUserPublicLookupResponseDTO struct {
	UserID            string `json:"user_id"`
	Email             string `json:"email"`
	Name              string `json:"name"`                 // Optional: for display
	PublicKeyInBase64 string `json:"public_key_in_base64"` // Base64 encoded
	VerificationID    string `json:"verification_id"`
}

type GatewayFederatedUserPublicLookupService interface {
	Execute(sessCtx context.Context, req *GatewayFederatedUserPublicLookupRequestDTO) (*GatewayFederatedUserPublicLookupResponseDTO, error)
}

type gatewayFederatedUserPublicLookupServiceImpl struct {
	config                *config.Configuration
	logger                *zap.Logger
	passwordProvider      password.Provider
	cache                 mongodbcache.Cacher
	jwtProvider           jwt.Provider
	userGetByEmailUseCase uc_user.FederatedUserGetByEmailUseCase
}

func NewGatewayFederatedUserPublicLookupService(
	cfg *config.Configuration,
	logger *zap.Logger,
	pp password.Provider,
	cach mongodbcache.Cacher,
	jwtp jwt.Provider,
	uc1 uc_user.FederatedUserGetByEmailUseCase,
) GatewayFederatedUserPublicLookupService {
	return &gatewayFederatedUserPublicLookupServiceImpl{cfg, logger, pp, cach, jwtp, uc1}
}

func (svc *gatewayFederatedUserPublicLookupServiceImpl) Execute(sessCtx context.Context, req *GatewayFederatedUserPublicLookupRequestDTO) (*GatewayFederatedUserPublicLookupResponseDTO, error) {
	//
	// STEP 1: Sanitization of the input.
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
		e["email"] = "Email is required"
	}
	if len(req.Email) > 255 {
		e["email"] = "Email is too long"
	}
	// if req.Module == 0 {
	// 	e["module"] = "Module is required"
	// } else {
	// 	// Assuming MonolithModulePaperCloud is the only valid module for now
	// 	if req.Module != int(constants.MonolithModuleMapleFile) && req.Module != int(constants.MonolithModulePaperCloud) {
	// 		e["module"] = "Module is invalid"
	// 	}
	// }

	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 3:
	//

	// Lookup the federateduser in our database, else return a `400 Bad Request` error.
	u, err := svc.userGetByEmailUseCase.Execute(sessCtx, req.Email)
	if err != nil {
		svc.logger.Error("failed getting user by email from database",
			zap.Any("error", err))
		return nil, err
	}
	if u == nil {
		return nil, httperror.NewForBadRequestWithSingleField("email", "Email address does not exists")
	}

	dto := &GatewayFederatedUserPublicLookupResponseDTO{
		UserID:            u.ID.Hex(),
		Email:             u.Email,
		Name:              u.Name,
		PublicKeyInBase64: base64.StdEncoding.EncodeToString(u.PublicKey.Key),
		VerificationID:    u.VerificationID,
	}

	return dto, nil
}
