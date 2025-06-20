# Token Refresh API Documentation

## Overview

This API endpoint allows users to refresh their authentication tokens using a valid refresh token. The refresh process generates new access and refresh tokens while maintaining the user's session. The implementation includes automatic background token refresh via a web worker.

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
| `value` | string | Yes | Valid encrypted refresh token received from login or previous refresh |

## Response Structure

### Success Response (HTTP 201)

```json
{
  "username": "user@example.com",
  "encrypted_access_token": "base64_encoded_encrypted_access_token",
  "encrypted_refresh_token": "base64_encoded_encrypted_refresh_token",
  "access_token_expiry_date": "2024-01-15T11:00:00Z",
  "refresh_token_expiry_date": "2024-01-29T10:30:00Z",
  "token_nonce": "base64_encoded_nonce"
}
```

### Field Descriptions (Response)

| Field | Type | Description |
|-------|------|----------|-------------|
| `username` | string | User's email address |
| `encrypted_access_token` | string | Base64-encoded encrypted access token |
| `encrypted_refresh_token` | string | Base64-encoded encrypted refresh token |
| `access_token_expiry_date` | string | ISO 8601 timestamp when access token expires (30 minutes) |
| `refresh_token_expiry_date` | string | ISO 8601 timestamp when refresh token expires (14 days) |
| `token_nonce` | string | Base64-encoded nonce used for token encryption |

### Token Security

All tokens are encrypted using the user's public key with NaCl box encryption. Users must have a properly configured public key in their security data for authentication to succeed. The access and refresh tokens are encrypted separately for enhanced security.

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

#### User Not Properly Configured (HTTP 500)

```json
{
  "error": "Internal Server Error",
  "message": "user account not properly configured for secure authentication"
}
```

**Common error scenarios:**
- `"jwt refresh token failed"` - Invalid, expired, or malformed refresh token
- `"payload structure is wrong"` - Invalid JSON format
- Session expired or not found in cache
- `"user account not properly configured for secure authentication"` - User doesn't have required public key for encryption
- `"failed to encrypt authentication tokens"` - Token encryption failed
- Token generation failures

## Implementation Examples

### React.js/React Native with Background Worker

```javascript
// Manual token refresh function
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

      // Store encrypted tokens
      localStorage.setItem('mapleapps_encrypted_access_token', result.encrypted_access_token);
      localStorage.setItem('mapleapps_encrypted_refresh_token', result.encrypted_refresh_token);
      localStorage.setItem('mapleapps_token_nonce', result.token_nonce);
      localStorage.setItem('mapleapps_access_token_expiry', result.access_token_expiry_date);
      localStorage.setItem('mapleapps_refresh_token_expiry', result.refresh_token_expiry_date);
      localStorage.setItem('mapleapps_user_email', result.username);

      return result;
    } else {
      const error = await response.json();
      throw new Error(error.message || 'Failed to refresh token');
    }
  } catch (error) {
    console.error('Token refresh error:', error);
    // Clear tokens on refresh failure
    localStorage.removeItem('mapleapps_encrypted_access_token');
    localStorage.removeItem('mapleapps_encrypted_refresh_token');
    localStorage.removeItem('mapleapps_token_nonce');
    localStorage.removeItem('mapleapps_access_token_expiry');
    localStorage.removeItem('mapleapps_refresh_token_expiry');
    localStorage.removeItem('mapleapps_user_email');
    throw error;
  }
};

// Check if tokens need refresh
const isAccessTokenExpiringSoon = (minutesBeforeExpiry = 5) => {
  const expiryTime = localStorage.getItem('mapleapps_access_token_expiry');
  if (!expiryTime) return true;

  const expiry = new Date(expiryTime);
  const now = new Date();
  const timeUntilExpiry = expiry.getTime() - now.getTime();
  const warningThreshold = minutesBeforeExpiry * 60 * 1000;

  return timeUntilExpiry <= warningThreshold;
};

// Manual refresh if needed
const refreshTokenIfNeeded = async () => {
  const refreshTokenValue = localStorage.getItem('mapleapps_encrypted_refresh_token');

  if (!refreshTokenValue) {
    throw new Error('No refresh token available');
  }

  if (isAccessTokenExpiringSoon()) {
    console.log('Access token expiring soon, refreshing...');
    return await refreshToken(refreshTokenValue);
  }

  return null; // No refresh needed
};
```

### Background Worker Integration

The implementation includes an automatic background worker that monitors and refreshes tokens:

```javascript
// auth-worker.js configuration
const STORAGE_KEYS = {
  ENCRYPTED_ACCESS_TOKEN: "mapleapps_encrypted_access_token",
  ENCRYPTED_REFRESH_TOKEN: "mapleapps_encrypted_refresh_token",
  TOKEN_NONCE: "mapleapps_token_nonce",
  ACCESS_TOKEN_EXPIRY: "mapleapps_access_token_expiry",
  REFRESH_TOKEN_EXPIRY: "mapleapps_refresh_token_expiry",
  USER_EMAIL: "mapleapps_user_email",
};

const CHECK_INTERVAL = 30000; // Check every 30 seconds
const REFRESH_THRESHOLD = 5 * 60 * 1000; // Refresh 5 minutes before expiry

// Worker Manager usage
import workerManager from './services/workerManager';

// Initialize and start monitoring
const initializeBackgroundRefresh = async () => {
  try {
    await workerManager.initialize();

    if (localStorage.getItem('mapleapps_encrypted_access_token')) {
      workerManager.startMonitoring();
      console.log('Background token monitoring started');
    }

    // Listen for worker events
    workerManager.addAuthStateChangeListener((type, data) => {
      switch (type) {
        case 'token_refresh_success':
          console.log('Tokens refreshed automatically by worker');
          break;

        case 'token_refresh_failed':
          console.error('Worker token refresh failed:', data.error);
          // Clear tokens and redirect to login
          localStorage.clear();
          window.location.href = '/login';
          break;

        case 'force_logout':
          console.log('Refresh token expired, forcing logout');
          localStorage.clear();
          window.location.href = '/login';
          break;

        case 'token_status_update':
          console.log('Token status:', data.tokenInfo);
          break;
      }
    });
  } catch (error) {
    console.warn('Worker initialization failed, falling back to manual refresh:', error);
  }
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
    AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
    RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
    EncryptedAccessToken   string    `json:"encrypted_access_token"`
    EncryptedRefreshToken  string    `json:"encrypted_refresh_token"`
    TokenNonce             string    `json:"token_nonce"`
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
    encryptedAccessToken  string
    encryptedRefreshToken string
    tokenNonce           string
    accessExpiry         time.Time
    refreshExpiry        time.Time
}

func (tm *TokenManager) RefreshIfNeeded() error {
    // Check if access token expires in the next 5 minutes
    if time.Until(tm.accessExpiry) < 5*time.Minute {
        fmt.Println("Access token expiring soon, refreshing...")

        result, err := refreshToken(tm.encryptedRefreshToken)
        if err != nil {
            return fmt.Errorf("failed to refresh token: %w", err)
        }

        // Update stored tokens
        tm.encryptedAccessToken = result.EncryptedAccessToken
        tm.encryptedRefreshToken = result.EncryptedRefreshToken
        tm.tokenNonce = result.TokenNonce
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
        self.encrypted_access_token: Optional[str] = None
        self.encrypted_refresh_token: Optional[str] = None
        self.token_nonce: Optional[str] = None
        self.access_expiry: Optional[datetime] = None
        self.refresh_expiry: Optional[datetime] = None
        self.username: Optional[str] = None

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
                self.encrypted_access_token = result['encrypted_access_token']
                self.encrypted_refresh_token = result['encrypted_refresh_token']
                self.token_nonce = result['token_nonce']
                self.username = result['username']

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

    def is_access_token_expiring_soon(self, minutes_before: int = 5) -> bool:
        """Check if access token expires within specified minutes."""
        if not self.access_expiry:
            return True

        time_until_expiry = self.access_expiry - datetime.now()
        return time_until_expiry < timedelta(minutes=minutes_before)

    def refresh_if_needed(self) -> bool:
        """Automatically refresh tokens if access token expires soon."""
        if not self.refresh_expiry or not self.encrypted_refresh_token:
            return False

        # Check if refresh token is still valid
        if datetime.now() >= self.refresh_expiry:
            raise Exception("Refresh token has expired")

        # Check if access token expires in the next 5 minutes
        if self.is_access_token_expiring_soon():
            print("Access token expiring soon, refreshing...")
            try:
                self.refresh_token_request(self.encrypted_refresh_token)
                print("Tokens refreshed successfully")
                return True
            except Exception as e:
                print(f"Failed to refresh tokens: {e}")
                # Clear tokens on failure
                self.encrypted_access_token = None
                self.encrypted_refresh_token = None
                self.token_nonce = None
                raise

        return False
```

### cURL Example

```bash
# Using encrypted refresh token
curl -X POST https://api.mapleapps.ca/iam/api/v1/token/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "value": "your_encrypted_refresh_token_here"
  }'
```

## Worker Features

The background worker (`auth-worker.js`) provides:

### Automatic Token Monitoring
- Checks token status every 30 seconds
- Refreshes tokens 5 minutes before access token expiry
- Handles refresh failures gracefully

### Cross-Tab Synchronization
- Uses BroadcastChannel API for multi-tab communication
- Synchronizes token updates across all open tabs
- Maintains consistent authentication state

### Worker Events

| Event | Description |
|-------|-------------|
| `worker_ready` | Worker initialized and ready |
| `token_status_update` | Token status checked |
| `token_refresh_success` | Tokens refreshed successfully |
| `token_refresh_failed` | Token refresh failed |
| `force_logout` | Refresh token expired, logout required |
| `worker_error` | Worker encountered an error |

### Worker Message Protocol

```javascript
// Start monitoring
worker.postMessage({ type: 'start_monitoring' });

// Stop monitoring
worker.postMessage({ type: 'stop_monitoring' });

// Force token check
worker.postMessage({ type: 'force_token_check' });

// Manual refresh
worker.postMessage({
  type: 'manual_refresh',
  data: {
    refreshToken: 'encrypted_refresh_token',
    storageData: { /* current localStorage data */ }
  }
});

// Get worker status
worker.postMessage({ type: 'get_worker_status' });
```

## Important Notes

### Token Lifecycle

1. **Access Token Expiry**: 30 minutes from issuance
2. **Refresh Token Expiry**: 14 days from issuance
3. **Session Rotation**: Each refresh creates a new session with new tokens
4. **Old Session Cleanup**: Previous session is replaced with new session

### Security Considerations

1. **Session Rotation**: Each token refresh generates entirely new access and refresh tokens
2. **Separate Encryption**: Access and refresh tokens are encrypted separately for enhanced security
3. **Required Encryption**: All tokens are encrypted with the user's public key - no plaintext fallback
4. **Secure Storage**: Store refresh tokens securely on the client side using the specified localStorage keys
5. **Automatic Cleanup**: Failed refresh attempts should clear all stored tokens

### LocalStorage Keys

| Key | Description |
|-----|-------------|
| `mapleapps_encrypted_access_token` | Encrypted access token |
| `mapleapps_encrypted_refresh_token` | Encrypted refresh token |
| `mapleapps_token_nonce` | Token encryption nonce |
| `mapleapps_access_token_expiry` | Access token expiry timestamp |
| `mapleapps_refresh_token_expiry` | Refresh token expiry timestamp |
| `mapleapps_user_email` | User's email address |

### Best Practices

1. **Proactive Refresh**: The worker automatically refreshes tokens 5 minutes before expiry
2. **Error Handling**: Always handle refresh failures gracefully and redirect to login
3. **Token Validation**: Validate token expiry dates before making API calls
4. **Public Key Requirement**: Ensure users have properly configured public keys for encryption
5. **Worker Fallback**: Implement manual refresh as fallback if worker initialization fails

### Common Integration Patterns

#### Axios Interceptor with Worker

```javascript
// Axios interceptor that works with the background worker
axios.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401) {
      // Check if worker is handling the refresh
      const workerStatus = await workerManager.getWorkerStatus();

      if (workerStatus.isRefreshing) {
        // Wait for worker to complete refresh
        await new Promise(resolve => setTimeout(resolve, 1000));
        return axios.request(error.config);
      }

      // Manual refresh fallback
      try {
        await refreshTokenIfNeeded();
        return axios.request(error.config);
      } catch (refreshError) {
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }
    return Promise.reject(error);
  }
);
```

#### Token Health Check

```javascript
// Check token health and status
const getTokenHealth = () => {
  const accessExpiry = localStorage.getItem('mapleapps_access_token_expiry');
  const refreshExpiry = localStorage.getItem('mapleapps_refresh_token_expiry');

  if (!accessExpiry || !refreshExpiry) {
    return { status: 'no_tokens', canRefresh: false };
  }

  const now = new Date();
  const accessExpiryDate = new Date(accessExpiry);
  const refreshExpiryDate = new Date(refreshExpiry);

  if (now >= refreshExpiryDate) {
    return { status: 'refresh_expired', canRefresh: false };
  }

  if (now >= accessExpiryDate) {
    return { status: 'access_expired', canRefresh: true };
  }

  const timeUntilExpiry = accessExpiryDate - now;
  if (timeUntilExpiry < 5 * 60 * 1000) {
    return { status: 'expiring_soon', canRefresh: true };
  }

  return { status: 'healthy', canRefresh: false };
};
```

This API enables seamless token management with mandatory encryption and automatic background refresh for enhanced security. The implementation ensures continuous authentication without requiring frequent user interaction.
