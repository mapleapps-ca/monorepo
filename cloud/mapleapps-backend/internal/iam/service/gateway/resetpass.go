package gateway

import (
	"context"

	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/password"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/cassandracache"
)

type GatewayResetPasswordService interface {
	Execute(sessCtx context.Context, req *GatewayResetPasswordRequestIDO) (*GatewayResetPasswordResponseIDO, error)
}

type gatewayResetPasswordServiceImpl struct {
	passwordProvider      password.PasswordProvider
	cache                 cassandracache.Cacher
	jwtProvider           jwt.JWTProvider
	userGetByEmailUseCase uc_user.FederatedUserGetByEmailUseCase
	userUpdateUseCase     uc_user.FederatedUserUpdateUseCase
}

func NewGatewayResetPasswordService(
	pp password.PasswordProvider,
	cach cassandracache.Cacher,
	jwtp jwt.JWTProvider,
	uc1 uc_user.FederatedUserGetByEmailUseCase,
	uc2 uc_user.FederatedUserUpdateUseCase,
) GatewayResetPasswordService {
	return &gatewayResetPasswordServiceImpl{pp, cach, jwtp, uc1, uc2}
}

type GatewayResetPasswordRequestIDO struct {
	Code            string `json:"code"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type GatewayResetPasswordResponseIDO struct {
	Message string `json:"message"`
}

func (s *gatewayResetPasswordServiceImpl) Execute(sessCtx context.Context, req *GatewayResetPasswordRequestIDO) (*GatewayResetPasswordResponseIDO, error) {
	return nil, nil
	// ipAddress, _ := sessCtx.Value(constants.SessionIPAddress).(string)

	// //
	// // STEP 1: Sanization of input.
	// //

	// // Defensive Code: For security purposes we need to perform some sanitization on the inputs.
	// req.Email = strings.ToLower(req.Email)
	// req.Email = strings.ReplaceAll(req.Email, " ", "")
	// req.Email = strings.ReplaceAll(req.Email, "\t", "")
	// req.Email = strings.TrimSpace(req.Email)

	// //
	// // STEP 2: Validation of input.
	// //

	// e := make(map[string]string)
	// if req.Email == "" {
	// 	e["email"] = "Email address is required"
	// }
	// if len(req.Email) > 255 {
	// 	e["email"] = "too long"
	// }
	// if req.Password == "" {
	// 	e["password"] = "missing value"
	// }
	// if req.PasswordConfirm == "" {
	// 	e["password_confirm"] = "missing value"
	// }
	// if req.PasswordConfirm != req.Password {
	// 	e["password"] = "does not match"
	// 	e["password_confirm"] = "does not match"
	// }

	// if len(e) != 0 {
	// 	return nil, httperror.NewForBadRequest(&e)
	// }

	// //
	// // STEP 3:
	// //

	// // Lookup the user in our database, else return a `400 Bad Request` error.
	// u, err := s.userGetByEmailUseCase.Execute(sessCtx, req.Email)
	// if err != nil {
	// 	return nil, err
	// }
	// if u == nil {
	// 	return nil, httperror.NewForBadRequestWithSingleField("email", "Email address does not exist")
	// }

	// //
	// // STEP 4:
	// //

	// if req.Code != u.PasswordResetVerificationCode {
	// 	return nil, httperror.NewForBadRequestWithSingleField("code", "Verification code is incorrect")

	// }
	// if time.Now().After(u.PasswordResetVerificationExpiry) {
	// 	return nil, httperror.NewForBadRequestWithSingleField("code", "Verification code has expired")
	// }

	// //
	// // STEP 4: Hash the password and update the user's password in the database.
	// //

	// password, err := sstring.NewSecureString(req.Password)
	// if err != nil {
	// 	return nil, err
	// }
	// defer password.Wipe()

	// passwordHash, err := s.passwordProvider.GenerateHashFromPassword(password)
	// if err != nil {
	// 	return nil, err
	// }

	// // u.PasswordHash = passwordHash
	// // u.PasswordHashAlgorithm = s.passwordProvider.AlgorithmName()
	// u.PasswordResetVerificationCode = ""
	// u.PasswordResetVerificationExpiry = time.Time{} // This is equivalent to not-set time
	// u.ModifiedAt = time.Now()
	// u.ModifiedByName = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
	// u.ModifiedFromIPAddress = ipAddress
	// err = s.userUpdateUseCase.Execute(sessCtx, u)
	// if err != nil {
	// 	return nil, err
	// }

	// //
	// // STEP 5: Done
	// //

	// // Return our auth keys.
	// return &GatewayResetPasswordResponseIDO{
	// 	Message: "Password reset email has been sent",
	// }, nil
}
