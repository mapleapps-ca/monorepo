# Token Refresh API Documentation

## Overview

This API endpoint allows users to refresh their authentication tokens using a valid refresh token. The refresh process generates new access and refresh tokens while maintaining the user's session.

## Endpoint Details

**URL:** `POST /iam/api/v1/token/refresh`
**Authentication:** None required (refresh token provided in request body)
**Content-Type:** `application/json`

## Request Structure

### Request Body

```json
{
  "value": "string"
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `value` | string | Yes | Valid JWT refresh token received from login or previous refresh |

## Response Structure

### Success Response (HTTP 201)

```json
{
  "username": "user@example.com",
  "access_token": "string",
  "access_token_expiry_date": "2024-01-15T11:00:00Z",
  "refresh_token": "string",
  "refresh_token_expiry_date": "2024-01-29T10:30:00Z",
  "encrypted_tokens": "string",
  "token_nonce": "string"
}
```

### Field Descriptions (Response)

| Field | Type | Description |
|-------|------|-------------|
| `username` | string | User's email address |
| `access_token` | string | New JWT access token (legacy/fallback, may be empty) |
| `access_token_expiry_date` | string | ISO 8601 timestamp when access token expires (30 minutes) |
| `refresh_token` | string | New JWT refresh token (legacy/fallback, may be empty) |
| `refresh_token_expiry_date` | string | ISO 8601 timestamp when refresh token expires (14 days) |
| `encrypted_tokens` | string | Base64-encoded encrypted token payload (preferred method) |
| `token_nonce` | string | Base64-encoded nonce used for token encryption (when encrypted_tokens is present) |

### Token Types

The API returns tokens in two possible formats:

#### 1. Encrypted Tokens (Preferred)
When the user has a public key configured, tokens are encrypted with the user's public key:
- `encrypted_tokens`: Contains both access and refresh tokens encrypted together
- `token_nonce`: Nonce used for encryption
- `access_token` and `refresh_token`: Empty or omitted

#### 2. Plaintext Tokens (Legacy Fallback)
When the user doesn't have a public key or encryption fails:
- `access_token`: Plaintext JWT access token
- `refresh_token`: Plaintext JWT refresh token
- `encrypted_tokens` and `token_nonce`: Empty or omitted

### Error Responses

#### Invalid Token (HTTP 400/401)

```json
{
  "error": "Unauthorized",
  "message": "jwt refresh token failed"
}
```

#### Payload Structure Error (HTTP 400)

```json
{
  "error": "Bad Request",
  "details": {
    "non_field_error": "payload structure is wrong"
  }
}
```

#### Session Not Found (HTTP 500)

```json
{
  "error": "Internal Server Error",
  "message": "Session expired or not found"
}
```

**Common error scenarios:**
- `"jwt refresh token failed"` - Invalid, expired, or malformed refresh token
- `"payload structure is wrong"` - Invalid JSON format
- Session expired or not found in cache
- Token generation failures

## Implementation Examples

### React.js/React Native

```javascript
const refreshToken = async (refreshTokenValue) => {
  try {
    const response = await fetch('https://api.mapleapps.ca/iam/api/v1/token/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        value: refreshTokenValue
      })
    });

    if (response.status === 201) {
      const result = await response.json();
      console.log('Tokens refreshed successfully');

      // Handle encrypted tokens
      if (result.encrypted_tokens) {
        // Decrypt tokens using private key
        const tokens = await decryptTokens(result.encrypted_tokens, privateKey);
        localStorage.setItem('access_token', tokens.access_token);
        localStorage.setItem('refresh_token', tokens.refresh_token);
      } else {
        // Handle plaintext tokens (legacy)
        localStorage.setItem('access_token', result.access_token);
        localStorage.setItem('refresh_token', result.refresh_token);
      }

      // Store expiry times
      localStorage.setItem('access_token_expiry', result.access_token_expiry_date);
      localStorage.setItem('refresh_token_expiry', result.refresh_token_expiry_date);

      return result;
    } else {
      const error = await response.json();
      throw new Error(error.message || 'Failed to refresh token');
    }
  } catch (error) {
    console.error('Token refresh error:', error);
    // Clear tokens on refresh failure
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    throw error;
  }
};

// Utility function to automatically refresh tokens
const refreshTokenIfNeeded = async () => {
  const accessTokenExpiry = localStorage.getItem('access_token_expiry');
  const refreshTokenValue = localStorage.getItem('refresh_token');

  if (!accessTokenExpiry || !refreshTokenValue) {
    throw new Error('No tokens available');
  }

  // Check if access token expires in the next 5 minutes
  const expiryTime = new Date(accessTokenExpiry);
  const now = new Date();
  const timeUntilExpiry = expiryTime.getTime() - now.getTime();
  const fiveMinutesInMs = 5 * 60 * 1000;

  if (timeUntilExpiry < fiveMinutesInMs) {
    console.log('Access token expiring soon, refreshing...');
    return await refreshToken(refreshTokenValue);
  }

  return null; // No refresh needed
};
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type RefreshTokenRequest struct {
    Value string `json:"value"`
}

type RefreshTokenResponse struct {
    Username               string    `json:"username"`
    AccessToken            string    `json:"access_token,omitempty"`
    AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
    RefreshToken           string    `json:"refresh_token,omitempty"`
    RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
    EncryptedTokens        string    `json:"encrypted_tokens,omitempty"`
    TokenNonce             string    `json:"token_nonce,omitempty"`
}

func refreshToken(refreshTokenValue string) (*RefreshTokenResponse, error) {
    req := RefreshTokenRequest{
        Value: refreshTokenValue,
    }

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    resp, err := http.Post(
        "https://api.mapleapps.ca/iam/api/v1/token/refresh",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusCreated {
        var result RefreshTokenResponse
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            return nil, fmt.Errorf("failed to decode response: %w", err)
        }
        return &result, nil
    }

    var errorResp map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&errorResp)
    return nil, fmt.Errorf("token refresh failed: %v", errorResp)
}

// TokenManager handles automatic token refresh
type TokenManager struct {
    accessToken   string
    refreshToken  string
    accessExpiry  time.Time
    refreshExpiry time.Time
}

func (tm *TokenManager) RefreshIfNeeded() error {
    // Check if access token expires in the next 5 minutes
    if time.Until(tm.accessExpiry) < 5*time.Minute {
        fmt.Println("Access token expiring soon, refreshing...")

        result, err := refreshToken(tm.refreshToken)
        if err != nil {
            return fmt.Errorf("failed to refresh token: %w", err)
        }

        // Update stored tokens
        if result.EncryptedTokens != "" {
            // Handle encrypted tokens
            tokens, err := decryptTokens(result.EncryptedTokens, privateKey)
            if err != nil {
                return fmt.Errorf("failed to decrypt tokens: %w", err)
            }
            tm.accessToken = tokens.AccessToken
            tm.refreshToken = tokens.RefreshToken
        } else {
            // Handle plaintext tokens
            tm.accessToken = result.AccessToken
            tm.refreshToken = result.RefreshToken
        }

        tm.accessExpiry = result.AccessTokenExpiryDate
        tm.refreshExpiry = result.RefreshTokenExpiryDate

        fmt.Println("Tokens refreshed successfully")
    }

    return nil
}
```

### Python

```python
import requests
import json
from datetime import datetime, timedelta
from typing import Optional, Dict, Any

class TokenManager:
    def __init__(self, base_url: str = "https://api.mapleapps.ca"):
        self.base_url = base_url
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None
        self.access_expiry: Optional[datetime] = None
        self.refresh_expiry: Optional[datetime] = None

    def refresh_token_request(self, refresh_token_value: str) -> Dict[str, Any]:
        """Refresh authentication tokens using a refresh token."""
        url = f"{self.base_url}/iam/api/v1/token/refresh"

        payload = {
            "value": refresh_token_value
        }

        headers = {
            "Content-Type": "application/json"
        }

        try:
            response = requests.post(url, json=payload, headers=headers)

            if response.status_code == 201:
                result = response.json()

                # Update stored tokens
                if result.get('encrypted_tokens'):
                    # Handle encrypted tokens
                    tokens = self.decrypt_tokens(result['encrypted_tokens'], self.private_key)
                    self.access_token = tokens['access_token']
                    self.refresh_token = tokens['refresh_token']
                else:
                    # Handle plaintext tokens
                    self.access_token = result.get('access_token')
                    self.refresh_token = result.get('refresh_token')

                # Parse expiry dates
                self.access_expiry = datetime.fromisoformat(
                    result['access_token_expiry_date'].replace('Z', '+00:00')
                )
                self.refresh_expiry = datetime.fromisoformat(
                    result['refresh_token_expiry_date'].replace('Z', '+00:00')
                )

                return result
            else:
                error_data = response.json()
                raise Exception(f"Token refresh failed: {error_data.get('message', 'Unknown error')}")

        except requests.RequestException as e:
            raise Exception(f"Network error during token refresh: {str(e)}")

    def refresh_if_needed(self) -> bool:
        """Automatically refresh tokens if access token expires soon."""
        if not self.access_expiry or not self.refresh_token:
            return False

        # Check if access token expires in the next 5 minutes
        time_until_expiry = self.access_expiry - datetime.now()
        if time_until_expiry < timedelta(minutes=5):
            print("Access token expiring soon, refreshing...")
            try:
                self.refresh_token_request(self.refresh_token)
                print("Tokens refreshed successfully")
                return True
            except Exception as e:
                print(f"Failed to refresh tokens: {e}")
                # Clear tokens on failure
                self.access_token = None
                self.refresh_token = None
                raise

        return False
```

### cURL Example

```bash
curl -X POST https://api.mapleapps.ca/iam/api/v1/token/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "value": "your_refresh_token_here"
  }'
```

## Important Notes

### Token Lifecycle

1. **Access Token Expiry**: 30 minutes from issuance
2. **Refresh Token Expiry**: 14 days from issuance
3. **Session Rotation**: Each refresh creates a new session with new tokens
4. **Old Session Cleanup**: Previous session is replaced with new session

### Security Considerations

1. **Session Rotation**: Each token refresh generates entirely new access and refresh tokens
2. **Encrypted Tokens**: When available, tokens are encrypted with the user's public key
3. **Secure Storage**: Store refresh tokens securely on the client side
4. **Automatic Cleanup**: Failed refresh attempts should clear stored tokens

### Best Practices

1. **Proactive Refresh**: Refresh tokens before access tokens expire (5 minutes before expiry recommended)
2. **Error Handling**: Always handle refresh failures gracefully and redirect to login
3. **Token Validation**: Validate token expiry dates before making API calls
4. **Fallback Strategy**: Support both encrypted and plaintext token formats

### Common Integration Patterns

#### Automatic Token Refresh Middleware

```javascript
// Axios interceptor for automatic token refresh
axios.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401) {
      try {
        await refreshTokenIfNeeded();
        // Retry original request with new token
        return axios.request(error.config);
      } catch (refreshError) {
        // Redirect to login
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }
    return Promise.reject(error);
  }
);
```

#### Token Refresh on App Start

```javascript
// Check and refresh tokens when app starts
const initializeAuth = async () => {
  try {
    await refreshTokenIfNeeded();
    console.log('Authentication initialized');
  } catch (error) {
    console.log('No valid session, redirecting to login');
    window.location.href = '/login';
  }
};
```

This API enables seamless token management and maintains user sessions without requiring frequent re-authentication.
