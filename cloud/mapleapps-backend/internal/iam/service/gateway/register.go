package gateway

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/keys"
	uc_emailer "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/emailer"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/random"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/crypto"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/password"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/cassandracache"
)

type GatewayFederatedUserRegisterService interface {
	Execute(
		sessCtx context.Context,
		req *RegisterCustomerRequestIDO,
	) error
}

type gatewayFederatedUserRegisterServiceImpl struct {
	config                                    *config.Configuration
	logger                                    *zap.Logger
	passwordProvider                          password.Provider
	cache                                     cassandracache.Cacher
	jwtProvider                               jwt.Provider
	userGetByEmailUseCase                     uc_user.FederatedUserGetByEmailUseCase
	userCreateUseCase                         uc_user.FederatedUserCreateUseCase
	userUpdateUseCase                         uc_user.FederatedUserUpdateUseCase
	sendFederatedUserVerificationEmailUseCase uc_emailer.SendFederatedUserVerificationEmailUseCase
}

func NewGatewayFederatedUserRegisterService(
	cfg *config.Configuration,
	logger *zap.Logger,
	pp password.Provider,
	cach cassandracache.Cacher,
	jwtp jwt.Provider,
	uc1 uc_user.FederatedUserGetByEmailUseCase,
	uc2 uc_user.FederatedUserCreateUseCase,
	uc3 uc_user.FederatedUserUpdateUseCase,
	uc4 uc_emailer.SendFederatedUserVerificationEmailUseCase,
) GatewayFederatedUserRegisterService {
	logger = logger.Named("GatewayFederatedUserRegisterService")
	return &gatewayFederatedUserRegisterServiceImpl{cfg, logger, pp, cach, jwtp, uc1, uc2, uc3, uc4}
}

type RegisterCustomerRequestIDO struct {
	// --- Application and personal identiable information (PII) ---
	BetaAccessCode                                 string `json:"beta_access_code"` // Temporary code for beta access
	FirstName                                      string `json:"first_name"`
	LastName                                       string `json:"last_name"`
	Email                                          string `json:"email"`
	Phone                                          string `json:"phone,omitempty"`
	Country                                        string `json:"country,omitempty"`
	Timezone                                       string `bson:"timezone" json:"timezone"`
	AgreeTermsOfService                            bool   `json:"agree_terms_of_service,omitempty"`
	AgreePromotions                                bool   `json:"agree_promotions,omitempty"`
	AgreeToTrackingAcrossThirdPartyAppsAndServices bool   `json:"agree_to_tracking_across_third_party_apps_and_services,omitempty"`

	// Module refers to which module the user is registering for.
	Module int `json:"module,omitempty"`

	// --- E2EE Related ---
	Salt                              string `json:"salt"`
	PublicKey                         string `json:"publicKey"`
	EncryptedMasterKey                string `json:"encryptedMasterKey"`
	EncryptedPrivateKey               string `json:"encryptedPrivateKey"`
	EncryptedRecoveryKey              string `json:"encryptedRecoveryKey"`
	MasterKeyEncryptedWithRecoveryKey string `json:"masterKeyEncryptedWithRecoveryKey"`
	VerificationID                    string `json:"verificationID"`
}

func (svc *gatewayFederatedUserRegisterServiceImpl) Execute(
	sessCtx context.Context,
	req *RegisterCustomerRequestIDO,
) error {
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
	if req.BetaAccessCode == "" {
		e["beta_access_code"] = "Beta access code is required"
	} else {
		if req.BetaAccessCode != svc.config.App.BetaAccessCode {
			e["beta_access_code"] = "Invalid beta access code"
		}
	}
	if req.FirstName == "" {
		e["first_name"] = "First name is required"
	}
	if req.LastName == "" {
		e["last_name"] = "Last name is required"
	}
	if req.Email == "" {
		e["email"] = "Email is required"
	}
	if len(req.Email) > 255 {
		e["email"] = "Email is too long"
	}
	if req.Phone == "" {
		e["phone"] = "Phone number is required"
	}
	if req.Country == "" {
		e["country"] = "Country is required"
	}
	if req.Timezone == "" {
		e["timezone"] = "Timezone is required"
	}
	if req.AgreeTermsOfService == false {
		e["agree_terms_of_service"] = "Agreeing to terms of service is required and you must agree to the terms before proceeding"
	}
	if req.Module == 0 {
		e["module"] = "Module is required"
	} else {
		// Assuming MonolithModulePaperCloud is the only valid module for now
		if req.Module != int(constants.MonolithModuleMapleFile) && req.Module != int(constants.MonolithModulePaperCloud) {
			e["module"] = "Module is invalid"
		}
	}

	// --- E2EE Related Validation ---
	if req.Salt == "" {
		e["salt"] = "Salt is required"
	}
	if req.PublicKey == "" {
		e["publicKey"] = "Public key is required"
	}
	if req.EncryptedMasterKey == "" {
		e["encryptedMasterKey"] = "Encrypted master key is required"
	}
	if req.EncryptedPrivateKey == "" {
		e["encryptedPrivateKey"] = "Encrypted private key is required"
	}
	if req.EncryptedRecoveryKey == "" {
		e["encryptedRecoveryKey"] = "Encrypted recovery key is required"
	}
	if req.MasterKeyEncryptedWithRecoveryKey == "" {
		e["masterKeyEncryptedWithRecoveryKey"] = "Master key encrypted with recovery key is required"
	}

	// Developers Note: It's OK if user forgets, fall back to cloud service to generate it for the user.
	// if req.VerificationID == "" {
	// 	e["verificationID"] = "Verification ID is required"
	// }

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 3:
	//

	// Lookup the federateduser in our database, else return a `400 Bad Request` error.
	u, err := svc.userGetByEmailUseCase.Execute(sessCtx, req.Email)
	if err != nil {
		svc.logger.Error("failed getting user by email from database",
			zap.Any("error", err))
		return err
	}
	if u != nil {
		return httperror.NewForBadRequestWithSingleField("email", "Email address already exists")
	}
	// Create our federateduser.
	u, err = svc.createCustomerFederatedUserForRequest(sessCtx, req)
	if err != nil {
		return err
	}

	if err := svc.sendFederatedUserVerificationEmailUseCase.Execute(context.Background(), req.Module, u); err != nil {
		return err
	}

	return nil
}

func (s *gatewayFederatedUserRegisterServiceImpl) createCustomerFederatedUserForRequest(sessCtx context.Context, req *RegisterCustomerRequestIDO) (*dom_user.FederatedUser, error) {
	// Get the IP address from the session context
	ipAddress, _ := sessCtx.Value(constants.SessionIPAddress).(string)

	// Generate email verification code
	emailVerificationCode, err := random.GenerateSixDigitCode()
	if err != nil {
		return nil, err
	}

	// Decode base64 encoded crypto values
	saltBytes, err := base64.RawURLEncoding.DecodeString(req.Salt)
	if err != nil {
		s.logger.Error("failed decoding salt", zap.Error(err))
		return nil, fmt.Errorf("invalid salt format: %w", err)
	}

	publicKeyBytes, err := base64.RawURLEncoding.DecodeString(req.PublicKey)
	if err != nil {
		s.logger.Error("failed decoding public key", zap.Error(err))
		return nil, fmt.Errorf("invalid public key format: %w", err)
	}

	// Process EncryptedMasterKey - expect base64 encoded data with nonce and ciphertext
	encMasterKeyBytes, err := base64.RawURLEncoding.DecodeString(req.EncryptedMasterKey)
	if err != nil {
		s.logger.Error("failed decoding encrypted master key", zap.Error(err))
		return nil, fmt.Errorf("invalid encrypted master key format: %w", err)
	}

	// Split into nonce and ciphertext based on crypto.NonceSize (now 12 bytes for ChaCha20-Poly1305)
	if len(encMasterKeyBytes) < crypto.NonceSize {
		return nil, fmt.Errorf("encrypted master key data too short")
	}

	// Store current key in history
	currentTime := time.Now() // Capture the current time once for initial user registration time.
	historicalKey := keys.EncryptedHistoricalKey{
		KeyVersion:    1,
		Nonce:         encMasterKeyBytes[:crypto.NonceSize],
		Ciphertext:    encMasterKeyBytes[crypto.NonceSize:],
		RotatedAt:     currentTime,
		RotatedReason: "Initial user registration",
		Algorithm:     crypto.ChaCha20Poly1305Algorithm,
	}

	encryptedMasterKey := keys.EncryptedMasterKey{
		Nonce:        encMasterKeyBytes[:crypto.NonceSize],
		Ciphertext:   encMasterKeyBytes[crypto.NonceSize:],
		KeyVersion:   1,
		RotatedAt:    &currentTime, // Pass the address of the captured time
		PreviousKeys: []keys.EncryptedHistoricalKey{historicalKey},
	}

	// Process EncryptedPrivateKey
	encPrivateKeyBytes, err := base64.RawURLEncoding.DecodeString(req.EncryptedPrivateKey)
	if err != nil {
		s.logger.Error("failed decoding encrypted private key", zap.Error(err))
		return nil, fmt.Errorf("invalid encrypted private key format: %w", err)
	}

	if len(encPrivateKeyBytes) < crypto.NonceSize {
		return nil, fmt.Errorf("encrypted private key data too short")
	}
	encryptedPrivateKey := keys.EncryptedPrivateKey{
		Nonce:      encPrivateKeyBytes[:crypto.NonceSize],
		Ciphertext: encPrivateKeyBytes[crypto.NonceSize:],
	}

	// Process EncryptedRecoveryKey
	encRecoveryKeyBytes, err := base64.RawURLEncoding.DecodeString(req.EncryptedRecoveryKey)
	if err != nil {
		s.logger.Error("failed decoding encrypted recovery key", zap.Error(err))
		return nil, fmt.Errorf("invalid encrypted recovery key format: %w", err)
	}

	if len(encRecoveryKeyBytes) < crypto.NonceSize {
		return nil, fmt.Errorf("encrypted recovery key data too short")
	}
	encryptedRecoveryKey := keys.EncryptedRecoveryKey{
		Nonce:      encRecoveryKeyBytes[:crypto.NonceSize],
		Ciphertext: encRecoveryKeyBytes[crypto.NonceSize:],
	}

	// Process MasterKeyEncryptedWithRecoveryKey
	encMasterWithRecoveryBytes, err := base64.RawURLEncoding.DecodeString(req.MasterKeyEncryptedWithRecoveryKey)
	if err != nil {
		s.logger.Error("failed decoding master key encrypted with recovery key", zap.Error(err))
		return nil, fmt.Errorf("invalid master key encrypted with recovery key format: %w", err)
	}

	if len(encMasterWithRecoveryBytes) < crypto.NonceSize {
		return nil, fmt.Errorf("master key encrypted with recovery key data too short")
	}
	masterKeyEncryptedWithRecoveryKey := keys.MasterKeyEncryptedWithRecoveryKey{
		Nonce:      encMasterWithRecoveryBytes[:crypto.NonceSize],
		Ciphertext: encMasterWithRecoveryBytes[crypto.NonceSize:],
	}

	// Generate VerificationID if not provided via server side.
	if req.VerificationID == "" {
		verificationID, err := crypto.GenerateVerificationID(publicKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to generate verification id from given public key via server-side and no verification id provided client-side")
		}
		req.VerificationID = verificationID
	}

	// Defensive Coding: Verify the `verificationID` to make sure we are enforcing the fact that `verificationID` are derived from public keys.
	if !crypto.VerifyVerificationID(publicKeyBytes, req.VerificationID) {
		return nil, fmt.Errorf("failed to verify the user verification id, was it generated from the public key?")
	}

	// Create the PublicKey object
	publicKey := keys.PublicKey{
		Key:            publicKeyBytes,
		VerificationID: req.VerificationID,
	}

	userID := gocql.TimeUUID()

	profiledata := &dom_user.FederatedUserProfileData{
		Phone:                req.Phone,
		Country:              req.Country,
		Timezone:             req.Timezone,
		Region:               "",
		City:                 "",
		PostalCode:           "",
		AddressLine1:         "",
		AddressLine2:         "",
		HasShippingAddress:   false,
		ShippingName:         "",
		ShippingPhone:        "",
		ShippingCountry:      "",
		ShippingRegion:       "",
		ShippingCity:         "",
		ShippingPostalCode:   "",
		ShippingAddressLine1: "",
		ShippingAddressLine2: "",
		AgreeTermsOfService:  req.AgreeTermsOfService,
		AgreePromotions:      req.AgreePromotions,
		AgreeToTrackingAcrossThirdPartyAppsAndServices: req.AgreeToTrackingAcrossThirdPartyAppsAndServices,
	}
	securitydata := &dom_user.FederatedUserSecurityData{
		// --- E2EE ---
		PasswordSalt:                      saltBytes,
		KDFParams:                         keys.DefaultKDFParams(),
		PublicKey:                         publicKey,
		EncryptedMasterKey:                encryptedMasterKey,
		EncryptedPrivateKey:               encryptedPrivateKey,
		EncryptedRecoveryKey:              encryptedRecoveryKey,
		MasterKeyEncryptedWithRecoveryKey: masterKeyEncryptedWithRecoveryKey,
		VerificationID:                    req.VerificationID,
		LastPasswordChange:                time.Now(),
		KDFParamsNeedUpgrade:              false,
		CurrentKeyVersion:                 1,
		LastKeyRotation:                   &currentTime,
		KeyRotationPolicy:                 nil,

		// --- Quality of Assurance on Email intake
		WasEmailVerified: false,
		Code:             fmt.Sprintf("%s", emailVerificationCode),
		CodeType:         dom_user.FederatedUserCodeTypeEmailVerification,
		CodeExpiry:       time.Now().Add(72 * time.Hour),
	}
	metadata := &dom_user.FederatedUserMetadata{
		CreatedByUserID:       userID,
		CreatedAt:             time.Now(),
		CreatedByName:         fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		CreatedFromIPAddress:  ipAddress,
		ModifiedByUserID:      userID,
		ModifiedAt:            time.Now(),
		ModifiedByName:        fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		ModifiedFromIPAddress: ipAddress,
	}

	u := &dom_user.FederatedUser{
		ID:           userID,
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Name:         fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		LexicalName:  fmt.Sprintf("%s, %s", req.LastName, req.FirstName),
		Role:         dom_user.FederatedUserRoleIndividual,
		Status:       dom_user.FederatedUserStatusActive,
		Timezone:     req.Timezone,
		ProfileData:  profiledata,
		SecurityData: securitydata,
		Metadata:     metadata,
		CreatedAt:    metadata.CreatedAt,
		ModifiedAt:   metadata.ModifiedAt,
	}
	err = s.userCreateUseCase.Execute(sessCtx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}
