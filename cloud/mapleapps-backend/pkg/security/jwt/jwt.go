package jwt

import (
	"errors"
	"time"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/jwt_utils"
	sbytes "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/securebytes"
)

// Provider provides interface for abstracting JWT generation.
type Provider interface {
	GenerateJWTToken(uuid string, ad time.Duration) (string, time.Time, error)
	GenerateJWTTokenPair(uuid string, ad time.Duration, rd time.Duration) (string, time.Time, string, time.Time, error)
	ProcessJWTToken(reqToken string) (string, error)
}

type jwtProvider struct {
	hmacSecret *sbytes.SecureBytes
}

// NewProvider Constructor that returns the JWT generator.
func NewProvider(cfg *config.Configuration) Provider {
	return jwtProvider{
		hmacSecret: cfg.App.AdministrationHMACSecret,
	}
}

// GenerateJWTToken generates a single JWT token.
func (p jwtProvider) GenerateJWTToken(uuid string, ad time.Duration) (string, time.Time, error) {
	return jwt_utils.GenerateJWTToken(p.hmacSecret.Bytes(), uuid, ad)
}

// GenerateJWTTokenPair Generate the `access token` and `refresh token` for the secret key.
func (p jwtProvider) GenerateJWTTokenPair(uuid string, ad time.Duration, rd time.Duration) (string, time.Time, string, time.Time, error) {
	return jwt_utils.GenerateJWTTokenPair(p.hmacSecret.Bytes(), uuid, ad, rd)
}

func (p jwtProvider) ProcessJWTToken(reqToken string) (string, error) {
	if p.hmacSecret == nil {
		return "", errors.New("HMAC secret is required")
	}
	return jwt_utils.ProcessJWTToken(p.hmacSecret.Bytes(), reqToken)
}
