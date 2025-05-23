// cloud/backend/internal/iam/service/gateway/verifyott.go
package gateway

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/nacl/box"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/keys"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/crypto"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/security/jwt"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/database/mongodbcache"
)

// Data structures for OTT verification
type GatewayVerifyLoginOTTRequestIDO struct {
	Email string `json:"email"`
	OTT   string `json:"ott"`
}

type GatewayVerifyLoginOTTResponseIDO struct {
	Salt                string         `json:"salt"`
	KDFParams           keys.KDFParams `json:"kdf_params" bson:"kdf_params"`
	PublicKey           string         `json:"publicKey"`
	EncryptedMasterKey  string         `json:"encryptedMasterKey"`
	EncryptedPrivateKey string         `json:"encryptedPrivateKey"`
	EncryptedChallenge  string         `json:"encryptedChallenge"`
	ChallengeID         string         `json:"challengeId"`

	// KDF upgrade and key rotation fields.
	LastPasswordChange   time.Time               `json:"last_password_change" bson:"last_password_change"`
	KDFParamsNeedUpgrade bool                    `json:"kdf_params_need_upgrade" bson:"kdf_params_need_upgrade"`
	CurrentKeyVersion    int                     `json:"current_key_version" bson:"current_key_version"`
	LastKeyRotation      *time.Time              `json:"last_key_rotation,omitempty" bson:"last_key_rotation,omitempty"`
	KeyRotationPolicy    *keys.KeyRotationPolicy `json:"key_rotation_policy,omitempty" bson:"key_rotation_policy,omitempty"`
}

// ChallengeData structure to be stored in cache
type ChallengeData struct {
	Email           string    `json:"email"`
	ChallengeID     string    `json:"challenge_id"`
	Challenge       string    `json:"challenge"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	IsVerified      bool      `json:"is_verified"`
	FederatedUserID string    `json:"federated_user_id"`
}

// Service interface for OTT verification
type GatewayVerifyLoginOTTService interface {
	Execute(sessCtx context.Context, req *GatewayVerifyLoginOTTRequestIDO) (*GatewayVerifyLoginOTTResponseIDO, error)
}

// Implementation of OTT verification service
type gatewayVerifyLoginOTTServiceImpl struct {
	config                *config.Configuration
	logger                *zap.Logger
	cache                 mongodbcache.Cacher
	jwtProvider           jwt.Provider
	userGetByEmailUseCase uc_user.FederatedUserGetByEmailUseCase
}

func NewGatewayVerifyLoginOTTService(
	config *config.Configuration,
	logger *zap.Logger,
	cache mongodbcache.Cacher,
	jwtProvider jwt.Provider,
	userGetByEmailUseCase uc_user.FederatedUserGetByEmailUseCase,
) GatewayVerifyLoginOTTService {
	return &gatewayVerifyLoginOTTServiceImpl{
		config:                config,
		logger:                logger,
		cache:                 cache,
		jwtProvider:           jwtProvider,
		userGetByEmailUseCase: userGetByEmailUseCase,
	}
}

func (s *gatewayVerifyLoginOTTServiceImpl) Execute(sessCtx context.Context, req *GatewayVerifyLoginOTTRequestIDO) (*GatewayVerifyLoginOTTResponseIDO, error) {
	// Validate input
	e := make(map[string]string)
	if req.Email == "" {
		e["email"] = "Email address is required"
	}
	if req.OTT == "" {
		e["ott"] = "Verification code is required"
	}
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	// Sanitize input
	req.Email = strings.ToLower(req.Email)
	req.Email = strings.ReplaceAll(req.Email, " ", "")
	req.OTT = strings.TrimSpace(req.OTT)

	// Retrieve OTT data from cache
	cacheKey := fmt.Sprintf("login_ott:%s", req.Email)
	ottDataJSON, err := s.cache.Get(sessCtx, cacheKey)
	if err != nil {
		s.logger.Error("Failed to retrieve OTT data",
			zap.Error(err))
		return nil, httperror.NewForBadRequestWithSingleField("ott", "Invalid or expired verification code")
	}

	if ottDataJSON == nil {
		s.logger.Error("OTT data not found in cache")
		return nil, httperror.NewForBadRequestWithSingleField("ott", "Invalid or expired verification code")
	}

	// Unmarshal the data from JSON
	var ottData LoginOTTData
	if err := json.Unmarshal(ottDataJSON, &ottData); err != nil {
		s.logger.Error("Failed to unmarshal OTT data",
			zap.Error(err))
		return nil, httperror.NewForBadRequestWithSingleField("ott", "Invalid verification code")
	}

	// Verify OTT
	if ottData.OTT != req.OTT {
		return nil, httperror.NewForBadRequestWithSingleField("ott", "Invalid verification code")
	}

	// Check expiry
	if time.Now().After(ottData.ExpiresAt) {
		return nil, httperror.NewForBadRequestWithSingleField("ott", "Verification code has expired")
	}

	// Check if already verified
	if ottData.IsVerified {
		return nil, httperror.NewForBadRequestWithSingleField("ott", "Verification code has already been used")
	}

	// Get user from database
	user, err := s.userGetByEmailUseCase.Execute(sessCtx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, httperror.NewForBadRequestWithSingleField("email", "Email address does not exist")
	}

	// Generate a challenge for final verification
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		s.logger.Error("Failed to generate challenge", zap.Error(err))
		return nil, fmt.Errorf("failed to process login: %w", err)
	}

	// Base64 encode the challenge for storage
	challengeBase64 := base64.StdEncoding.EncodeToString(challenge)

	// Generate a unique challenge ID
	challengeID := uuid.New().String()

	// Log the generated challenge details
	s.logger.Info("VerifyOTT: Generated challenge for caching",
		zap.String("challenge_id", challengeID),
		zap.String("challenge_bytes_preview_base64_for_cache", challengeBase64),
		zap.Int("challenge_bytes_length", len(challenge)),
	)

	// Store challenge in cache
	challengeData := ChallengeData{
		Email:           req.Email,
		ChallengeID:     challengeID,
		Challenge:       challengeBase64,
		CreatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(5 * time.Minute), // Challenge valid for 5 minutes
		IsVerified:      false,
		FederatedUserID: user.ID.Hex(),
	}

	// Generate a unique cache key for this challenge
	challengeCacheKey := fmt.Sprintf("login_challenge:%s", challengeID)

	// Marshal the challenge data to JSON
	challengeDataJSON, err := json.Marshal(challengeData)
	if err != nil {
		s.logger.Error("Failed to marshal challenge data", zap.Error(err))
		return nil, fmt.Errorf("failed to process login verification: %w", err)
	}

	// Store in cache with expiry
	if err := s.cache.SetWithExpiry(sessCtx, challengeCacheKey, challengeDataJSON, 5*time.Minute); err != nil {
		s.logger.Error("Failed to store challenge in cache", zap.Error(err))
		return nil, fmt.Errorf("failed to process login verification: %w", err)
	}

	// Mark OTT as verified
	ottData.IsVerified = true
	ottData.ChallengeID = challengeID

	// Marshal the updated OTT data to JSON
	updatedOTTDataJSON, err := json.Marshal(ottData)
	if err != nil {
		s.logger.Error("Failed to marshal updated OTT data", zap.Error(err))
		// Continue anyway, as the challenge is already stored
	} else {
		if err := s.cache.SetWithExpiry(sessCtx, cacheKey, updatedOTTDataJSON, 10*time.Minute); err != nil {
			s.logger.Error("Failed to update OTT in cache", zap.Error(err))
			// Continue anyway, as the challenge is already stored
		}
	}

	encryptedChallenge, err := getEncryptedChallenge(challenge, user)
	if err != nil {
		s.logger.Error("Failed to encrypt challenge", zap.Error(err))
		return nil, fmt.Errorf("failed to process login: %w", err)
	}

	// Convert structured keys to string representations for the response
	saltBase64 := base64.StdEncoding.EncodeToString(user.PasswordSalt)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(user.PublicKey.Key)

	// Combine nonce and ciphertext for encryptedMasterKey
	encryptedMasterKeyBytes := make([]byte, len(user.EncryptedMasterKey.Nonce)+len(user.EncryptedMasterKey.Ciphertext))
	copy(encryptedMasterKeyBytes, user.EncryptedMasterKey.Nonce)
	copy(encryptedMasterKeyBytes[len(user.EncryptedMasterKey.Nonce):], user.EncryptedMasterKey.Ciphertext)
	encryptedMasterKeyBase64 := base64.StdEncoding.EncodeToString(encryptedMasterKeyBytes)

	// Combine nonce and ciphertext for encryptedPrivateKey
	encryptedPrivateKeyBytes := make([]byte, len(user.EncryptedPrivateKey.Nonce)+len(user.EncryptedPrivateKey.Ciphertext))
	copy(encryptedPrivateKeyBytes, user.EncryptedPrivateKey.Nonce)
	copy(encryptedPrivateKeyBytes[len(user.EncryptedPrivateKey.Nonce):], user.EncryptedPrivateKey.Ciphertext)
	encryptedPrivateKeyBase64 := base64.StdEncoding.EncodeToString(encryptedPrivateKeyBytes)

	// Return encrypted keys and challenge for client-side password verification
	return &GatewayVerifyLoginOTTResponseIDO{
		// Base64 encoded encrypted keys and challenge
		Salt:                saltBase64,
		KDFParams:           user.KDFParams,
		PublicKey:           publicKeyBase64,
		EncryptedMasterKey:  encryptedMasterKeyBase64,
		EncryptedPrivateKey: encryptedPrivateKeyBase64,
		EncryptedChallenge:  encryptedChallenge,
		ChallengeID:         challengeID,
		// KDF upgrade and key rotation fields.
		LastPasswordChange:   user.LastPasswordChange,
		KDFParamsNeedUpgrade: user.KDFParamsNeedUpgrade,
		CurrentKeyVersion:    user.CurrentKeyVersion,
		LastKeyRotation:      user.LastKeyRotation,
		KeyRotationPolicy:    user.KeyRotationPolicy,
	}, nil
}

// getEncryptedChallenge encrypts the challenge using the user's public key
// in a way that is compatible with libsodium's crypto_box_seal_open.
func getEncryptedChallenge(challengeBytes []byte, user *domain.FederatedUser) (string, error) {
	publicKeyBytes := user.PublicKey.Key
	if len(publicKeyBytes) != crypto.PublicKeySize { // crypto.PublicKeySize is 32
		return "", fmt.Errorf("invalid public key length: got %d, want %d", len(publicKeyBytes), crypto.PublicKeySize)
	}

	var userPubKeyFixed [crypto.PublicKeySize]byte
	copy(userPubKeyFixed[:], publicKeyBytes)

	// box.SealAnonymous handles ephemeral key generation, nonce derivation,
	// encryption, and prepends the ephemeral public key to the output.
	// rand.Reader is used for generating the ephemeral key pair.
	sealedChallenge, err := box.SealAnonymous(nil, challengeBytes, &userPubKeyFixed, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to seal challenge anonymously: %w", err)
	}

	// The output of box.SealAnonymous is: ephemeral_public_key (32 bytes) || ciphertext (message + MAC)
	// This is exactly what crypto_box_seal_open expects.
	return base64.StdEncoding.EncodeToString(sealedChallenge), nil
}
