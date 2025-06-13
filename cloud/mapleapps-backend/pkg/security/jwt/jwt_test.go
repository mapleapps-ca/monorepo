package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/securebytes"
)

func setupTestProvider(t *testing.T) JWTProvider {
	hmacSecret, _ := securebytes.NewSecureBytes([]byte("test-secret"))
	cfg := &config.Configuration{
		App: config.AppConfig{
			AdministrationHMACSecret: hmacSecret,
		},
	}
	return NewJWTProvider(cfg)
}

func TestNewProvider(t *testing.T) {
	provider := setupTestProvider(t)
	assert.NotNil(t, provider)
}

func TestGenerateJWTToken(t *testing.T) {
	provider := setupTestProvider(t)
	uuid := "test-uuid"
	duration := time.Hour

	token, expiry, err := provider.GenerateJWTToken(uuid, duration)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiry.After(time.Now()))
	assert.True(t, expiry.Before(time.Now().Add(duration).Add(time.Second)))
}

func TestGenerateJWTTokenPair(t *testing.T) {
	provider := setupTestProvider(t)
	uuid := "test-uuid"
	accessDuration := time.Hour
	refreshDuration := time.Hour * 24

	accessToken, accessExpiry, refreshToken, refreshExpiry, err := provider.GenerateJWTTokenPair(uuid, accessDuration, refreshDuration)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.True(t, accessExpiry.After(time.Now()))
	assert.True(t, refreshExpiry.After(time.Now()))
	assert.True(t, accessExpiry.Before(time.Now().Add(accessDuration).Add(time.Second)))
	assert.True(t, refreshExpiry.Before(time.Now().Add(refreshDuration).Add(time.Second)))
}

func TestProcessJWTToken(t *testing.T) {
	provider := setupTestProvider(t)
	uuid := "test-uuid"
	duration := time.Hour

	// Generate a token first
	token, _, err := provider.GenerateJWTToken(uuid, duration)
	assert.NoError(t, err)

	// Process the generated token
	processedUUID, err := provider.ProcessJWTToken(token)
	assert.NoError(t, err)
	assert.Equal(t, uuid, processedUUID)
}

func TestProcessJWTToken_InvalidToken(t *testing.T) {
	provider := setupTestProvider(t)

	_, err := provider.ProcessJWTToken("invalid-token")
	assert.Error(t, err)
}

func TestProcessJWTToken_NilSecret(t *testing.T) {
	provider := jwtProvider{
		hmacSecret: nil,
	}

	_, err := provider.ProcessJWTToken("any-token")
	assert.Error(t, err)
	assert.Equal(t, "HMAC secret is required", err.Error())
}

func TestProcessJWTToken_ExpiredToken(t *testing.T) {
	provider := setupTestProvider(t)
	uuid := "test-uuid"
	duration := -time.Hour // negative duration for expired token

	token, _, err := provider.GenerateJWTToken(uuid, duration)
	assert.NoError(t, err)

	_, err = provider.ProcessJWTToken(token)
	assert.Error(t, err)
}
