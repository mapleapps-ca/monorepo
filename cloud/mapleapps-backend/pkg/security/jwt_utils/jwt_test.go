package jwt_utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testSecret = []byte("test-secret-key")

func TestGenerateJWTToken(t *testing.T) {
	uuid := "test-uuid"
	duration := time.Hour

	token, expiry, err := GenerateJWTToken(testSecret, uuid, duration)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiry.After(time.Now()))
	assert.True(t, expiry.Before(time.Now().Add(duration).Add(time.Second)))

	// Verify token can be processed
	processedUUID, err := ProcessJWTToken(testSecret, token)
	assert.NoError(t, err)
	assert.Equal(t, uuid, processedUUID)
}

func TestGenerateJWTTokenPair(t *testing.T) {
	uuid := "test-uuid"
	accessDuration := time.Hour
	refreshDuration := time.Hour * 24

	accessToken, accessExpiry, refreshToken, refreshExpiry, err := GenerateJWTTokenPair(
		testSecret,
		uuid,
		accessDuration,
		refreshDuration,
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.True(t, accessExpiry.After(time.Now()))
	assert.True(t, refreshExpiry.After(time.Now()))
	assert.True(t, accessExpiry.Before(time.Now().Add(accessDuration).Add(time.Second)))
	assert.True(t, refreshExpiry.Before(time.Now().Add(refreshDuration).Add(time.Second)))

	// Verify both tokens can be processed
	processedAccessUUID, err := ProcessJWTToken(testSecret, accessToken)
	assert.NoError(t, err)
	assert.Equal(t, uuid, processedAccessUUID)

	processedRefreshUUID, err := ProcessJWTToken(testSecret, refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, uuid, processedRefreshUUID)
}

func TestProcessJWTToken_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "not.a.token",
			wantErr: true,
		},
		{
			name:    "wrong signature",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzZXNzaW9uX3V1aWQiOiJ0ZXN0LXV1aWQiLCJleHAiOjE3MDQwNjc1NTF9.wrong",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid, err := ProcessJWTToken(testSecret, tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, uuid)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, uuid)
			}
		})
	}
}

func TestProcessJWTToken_Expired(t *testing.T) {
	uuid := "test-uuid"
	duration := -time.Hour // negative duration for expired token

	token, _, err := GenerateJWTToken(testSecret, uuid, duration)
	assert.NoError(t, err)

	processedUUID, err := ProcessJWTToken(testSecret, token)
	assert.Error(t, err)
	assert.Empty(t, processedUUID)
}
