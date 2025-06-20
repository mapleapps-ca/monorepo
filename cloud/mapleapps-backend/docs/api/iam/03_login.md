# MapleApps Login API Documentation

## Login Flow (3-Step Process)

The login process uses a secure 3-step authentication flow with end-to-end encryption:

1. **Request OTT** - Request a one-time token via email
2. **Verify OTT** - Verify the token and receive encrypted challenge
3. **Complete Login** - Decrypt challenge and receive authentication tokens

---

## Step 1: Request One-Time Token (OTT)

**URL:** `POST /iam/api/v1/request-ott`
**Authentication:** None required
**Content-Type:** `application/json`

### Request Structure

```json
{
  "email": "string"
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | User's email address. Will be automatically lowercased and whitespace trimmed. |

### Response Structure

#### Success Response (HTTP 200)

```json
{
  "message": "A verification code has been sent to your email"
}
```

#### Error Responses (HTTP 400)

```json
{
  "error": "Bad Request",
  "details": {
    "email": "Email address is required"
  }
}
```

**Common validation errors:**
- `"Email address is required"` - Missing email field
- `"Email address does not exist"` - Email not found in system

### Implementation Examples

#### React.js/React Native
```javascript
const requestOTT = async (email) => {
  try {
    const response = await fetch('https://api.mapleapps.ca/iam/api/v1/request-ott', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email: email.toLowerCase().trim()
      })
    });

    if (response.ok) {
      const result = await response.json();
      console.log('OTT requested:', result.message);
      return result;
    } else {
      const error = await response.json();
      throw new Error(error.details?.email || 'Failed to request OTT');
    }
  } catch (error) {
    console.error('Network error:', error);
    throw error;
  }
};
```

#### Go
```go
type RequestOTTRequest struct {
    Email string `json:"email"`
}

type RequestOTTResponse struct {
    Message string `json:"message"`
}

func requestOTT(email string) (*RequestOTTResponse, error) {
    req := RequestOTTRequest{Email: strings.ToLower(strings.TrimSpace(email))}

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(
        "https://api.mapleapps.ca/iam/api/v1/request-ott",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        var result RequestOTTResponse
        json.NewDecoder(resp.Body).Decode(&result)
        return &result, nil
    }

    return nil, fmt.Errorf("failed to request OTT: %d", resp.StatusCode)
}
```

---

## Step 2: Verify One-Time Token

**URL:** `POST /iam/api/v1/verify-ott`
**Authentication:** None required
**Content-Type:** `application/json`

### Request Structure

```json
{
  "email": "string",
  "ott": "string"
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | User's email address (same as step 1) |
| `ott` | string | Yes | 6-digit verification code received via email |

### Response Structure

#### Success Response (HTTP 200)

```json
{
  "salt": "string",
  "kdf_params": {
    "algorithm": "argon2id",
    "version": "1.0",
    "iterations": 3,
    "memory": 65536,
    "parallelism": 4,
    "salt_length": 32,
    "key_length": 32
  },
  "publicKey": "string",
  "encryptedMasterKey": "string",
  "encryptedPrivateKey": "string",
  "encryptedChallenge": "string",
  "challengeId": "string",
  "last_password_change": "2024-01-15T10:30:00Z",
  "kdf_params_need_upgrade": false,
  "current_key_version": 1,
  "last_key_rotation": "2024-01-15T10:30:00Z",
  "key_rotation_policy": null
}
```

### Field Descriptions (Response)

| Field | Type | Description |
|-------|------|-------------|
| `salt` | string | Base64-encoded password salt for key derivation |
| `kdf_params` | object | Key derivation function parameters |
| `publicKey` | string | Base64-encoded user's public key |
| `encryptedMasterKey` | string | Base64-encoded encrypted master key (nonce + ciphertext) |
| `encryptedPrivateKey` | string | Base64-encoded encrypted private key (nonce + ciphertext) |
| `encryptedChallenge` | string | Base64-encoded challenge encrypted with user's public key |
| `challengeId` | string | Unique identifier for this challenge |
| `last_password_change` | string | ISO 8601 timestamp of last password change |
| `kdf_params_need_upgrade` | boolean | Whether KDF parameters need upgrading |
| `current_key_version` | integer | Current encryption key version |
| `last_key_rotation` | string | ISO 8601 timestamp of last key rotation (nullable) |
| `key_rotation_policy` | object | Key rotation policy (nullable) |

#### Error Responses (HTTP 400)

```json
{
  "error": "Bad Request",
  "details": {
    "ott": "Invalid verification code"
  }
}
```

**Common validation errors:**
- `"Email address is required"` - Missing email
- `"Verification code is required"` - Missing OTT
- `"Invalid verification code"` - Wrong OTT
- `"Verification code has expired"` - OTT expired (10 minutes)
- `"Verification code has already been used"` - OTT already consumed

### Implementation Examples

#### React.js/React Native
```javascript
const verifyOTT = async (email, ott) => {
  try {
    const response = await fetch('https://api.mapleapps.ca/iam/api/v1/verify-ott', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email: email.toLowerCase().trim(),
        ott: ott.trim()
      })
    });

    if (response.ok) {
      const result = await response.json();
      console.log('OTT verified, received encrypted keys');
      return result;
    } else {
      const error = await response.json();
      throw new Error(error.details?.ott || 'Failed to verify OTT');
    }
  } catch (error) {
    console.error('Network error:', error);
    throw error;
  }
};
```

#### Go
```go
type VerifyOTTRequest struct {
    Email string `json:"email"`
    OTT   string `json:"ott"`
}

type KDFParams struct {
    Algorithm   string `json:"algorithm"`
    Version     string `json:"version"`
    Iterations  uint32 `json:"iterations"`
    Memory      uint32 `json:"memory"`
    Parallelism uint8  `json:"parallelism"`
    SaltLength  uint32 `json:"salt_length"`
    KeyLength   uint32 `json:"key_length"`
}

type VerifyOTTResponse struct {
    Salt                     string     `json:"salt"`
    KDFParams               KDFParams  `json:"kdf_params"`
    PublicKey               string     `json:"publicKey"`
    EncryptedMasterKey      string     `json:"encryptedMasterKey"`
    EncryptedPrivateKey     string     `json:"encryptedPrivateKey"`
    EncryptedChallenge      string     `json:"encryptedChallenge"`
    ChallengeID             string     `json:"challengeId"`
    LastPasswordChange      time.Time  `json:"last_password_change"`
    KDFParamsNeedUpgrade    bool       `json:"kdf_params_need_upgrade"`
    CurrentKeyVersion       int        `json:"current_key_version"`
    LastKeyRotation         *time.Time `json:"last_key_rotation,omitempty"`
}

func verifyOTT(email, ott string) (*VerifyOTTResponse, error) {
    req := VerifyOTTRequest{
        Email: strings.ToLower(strings.TrimSpace(email)),
        OTT:   strings.TrimSpace(ott),
    }

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(
        "https://api.mapleapps.ca/iam/api/v1/verify-ott",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        var result VerifyOTTResponse
        json.NewDecoder(resp.Body).Decode(&result)
        return &result, nil
    }

    return nil, fmt.Errorf("failed to verify OTT: %d", resp.StatusCode)
}
```

---

## Step 3: Complete Login

**URL:** `POST /iam/api/v1/complete-login`
**Authentication:** None required
**Content-Type:** `application/json`

### Request Structure

```json
{
  "email": "string",
  "challengeId": "string",
  "decryptedData": "string"
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | User's email address (same as previous steps) |
| `challengeId` | string | Yes | Challenge ID received from step 2 |
| `decryptedData` | string | Yes | Base64-encoded decrypted challenge data |

### Response Structure

#### Success Response (HTTP 200)

```json
{
  "access_token": "string",
  "access_token_expiry_date": "2024-01-15T11:00:00Z",
  "refresh_token": "string",
  "refresh_token_expiry_date": "2024-01-29T10:30:00Z",
  "encrypted_access_token": "string",
  "encrypted_refresh_token": "string",
  "token_nonce": "string",
  "username": "user@example.com"
}
```

### Field Descriptions (Response)

| Field | Type | Description |
|-------|------|-------------|
| `access_token` | string | JWT access token (legacy/fallback, may be empty) |
| `access_token_expiry_date` | string | ISO 8601 timestamp when access token expires (30 minutes) |
| `refresh_token` | string | JWT refresh token (legacy/fallback, may be empty) |
| `refresh_token_expiry_date` | string | ISO 8601 timestamp when refresh token expires (14 days) |
| `encrypted_access_token` | string | Base64-encoded encrypted access token (preferred method) |
| `encrypted_refresh_token` | string | Base64-encoded encrypted refresh token (preferred method) |
| `token_nonce` | string | Base64-encoded nonce used for token encryption |
| `username` | string | User's email address |

### Token Storage

The response provides tokens in both encrypted and plaintext formats for backward compatibility. The encrypted tokens are preferred and should be stored using the following localStorage keys:

- `mapleapps_encrypted_access_token` - Encrypted access token
- `mapleapps_encrypted_refresh_token` - Encrypted refresh token
- `mapleapps_token_nonce` - Token encryption nonce
- `mapleapps_access_token_expiry` - Access token expiry date
- `mapleapps_refresh_token_expiry` - Refresh token expiry date
- `mapleapps_user_email` - User's email address

#### Error Responses (HTTP 400)

```json
{
  "error": "Bad Request",
  "details": {
    "challengeId": "Invalid or expired challenge"
  }
}
```

**Common validation errors:**
- `"Email address is required"` - Missing email
- `"Challenge ID is required"` - Missing challengeId
- `"Decrypted data is required"` - Missing decryptedData
- `"Invalid or expired challenge"` - Challenge not found or expired
- `"Email address does not match challenge"` - Email mismatch
- `"Challenge has expired"` - Challenge expired (5 minutes)
- `"Challenge has already been used"` - Challenge already consumed
- `"Invalid challenge response"` - Decrypted data doesn't match

### Implementation Examples

#### React.js/React Native
```javascript
const completeLogin = async (email, challengeId, decryptedChallenge) => {
  try {
    const response = await fetch('https://api.mapleapps.ca/iam/api/v1/complete-login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email: email.toLowerCase().trim(),
        challengeId: challengeId,
        decryptedData: decryptedChallenge
      })
    });

    if (response.ok) {
      const result = await response.json();
      console.log('Login completed successfully');

      // Store encrypted tokens (preferred method)
      if (result.encrypted_access_token && result.encrypted_refresh_token) {
        localStorage.setItem('mapleapps_encrypted_access_token', result.encrypted_access_token);
        localStorage.setItem('mapleapps_encrypted_refresh_token', result.encrypted_refresh_token);
        localStorage.setItem('mapleapps_token_nonce', result.token_nonce);
        localStorage.setItem('mapleapps_access_token_expiry', result.access_token_expiry_date);
        localStorage.setItem('mapleapps_refresh_token_expiry', result.refresh_token_expiry_date);
        localStorage.setItem('mapleapps_user_email', result.username);
      } else {
        // Fallback to plaintext tokens
        localStorage.setItem('mapleapps_access_token', result.access_token);
        localStorage.setItem('mapleapps_refresh_token', result.refresh_token);
        localStorage.setItem('mapleapps_access_token_expiry', result.access_token_expiry_date);
        localStorage.setItem('mapleapps_refresh_token_expiry', result.refresh_token_expiry_date);
        localStorage.setItem('mapleapps_user_email', email);
      }

      return result;
    } else {
      const error = await response.json();
      throw new Error(error.details?.challengeId || 'Failed to complete login');
    }
  } catch (error) {
    console.error('Network error:', error);
    throw error;
  }
};
```

#### Go
```go
type CompleteLoginRequest struct {
    Email         string `json:"email"`
    ChallengeID   string `json:"challengeId"`
    DecryptedData string `json:"decryptedData"`
}

type CompleteLoginResponse struct {
    AccessToken            string    `json:"access_token,omitempty"`
    AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
    RefreshToken           string    `json:"refresh_token,omitempty"`
    RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
    EncryptedAccessToken   string    `json:"encrypted_access_token"`
    EncryptedRefreshToken  string    `json:"encrypted_refresh_token"`
    TokenNonce             string    `json:"token_nonce"`
    Username               string    `json:"username"`
}

func completeLogin(email, challengeId, decryptedData string) (*CompleteLoginResponse, error) {
    req := CompleteLoginRequest{
        Email:         strings.ToLower(strings.TrimSpace(email)),
        ChallengeID:   challengeId,
        DecryptedData: decryptedData,
    }

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(
        "https://api.mapleapps.ca/iam/api/v1/complete-login",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        var result CompleteLoginResponse
        json.NewDecoder(resp.Body).Decode(&result)
        return &result, nil
    }

    return nil, fmt.Errorf("failed to complete login: %d", resp.StatusCode)
}
```

## Background Token Refresh Worker

The implementation includes a background worker (`auth-worker.js`) that automatically monitors and refreshes tokens before they expire. The worker:

- Checks token status every 30 seconds
- Automatically refreshes tokens 5 minutes before expiry
- Synchronizes token state across multiple browser tabs
- Handles token refresh failures gracefully

### Worker Configuration

Place the worker file in your public directory:
```
/public/auth-worker.js
```

### Worker Integration Example

```javascript
// Initialize the worker manager
import workerManager from './services/workerManager';

// Start the worker when user is authenticated
const initializeAuth = async () => {
  try {
    await workerManager.initialize();

    if (localStorage.getItem('mapleapps_encrypted_access_token')) {
      workerManager.startMonitoring();
      console.log('Background token monitoring started');
    }
  } catch (error) {
    console.error('Worker initialization failed:', error);
    // Continue without background monitoring
  }
};

// Listen for worker events
workerManager.addAuthStateChangeListener((type, data) => {
  switch (type) {
    case 'token_refresh_success':
      console.log('Tokens refreshed automatically');
      break;
    case 'token_refresh_failed':
      console.error('Token refresh failed, redirecting to login');
      window.location.href = '/login';
      break;
    case 'force_logout':
      console.log('Session expired, redirecting to login');
      window.location.href = '/login';
      break;
  }
});
```

## Important Notes

### Security Considerations

1. **Login Flow Security**: The 3-step login process ensures that passwords never leave the client device and all authentication data is end-to-end encrypted.

2. **Token Expiration**:
   - Access tokens expire in 30 minutes
   - Refresh tokens expire in 14 days
   - OTT codes expire in 10 minutes
   - Challenges expire in 5 minutes

3. **Rate Limiting**: Implement appropriate rate limiting on the client side to prevent abuse.

### Error Handling

All APIs return structured error responses with specific field-level validation messages. Always check the `details` object for field-specific errors.

### Token Management

- Store tokens securely using the specified localStorage keys
- Implement automatic token refresh using the background worker
- Clear all tokens on logout or authentication errors
- Encrypted tokens are preferred over plaintext tokens

### Client-Side Cryptography

The login flow requires client-side implementation of:
- PBKDF2 key derivation (100,000 iterations, SHA-256)
- ChaCha20-Poly1305 encryption/decryption (libsodium secretbox)
- X25519 key exchange (libsodium box seal)
- Base64 encoding/decoding

Libraries used in the implementation:
- `libsodium-wrappers-sumo` for encryption operations
- `@scure/bip39` for mnemonic generation
- Web Crypto API for PBKDF2

### LocalStorage Keys Reference

| Key | Description |
|-----|-------------|
| `mapleapps_encrypted_access_token` | Encrypted access token |
| `mapleapps_encrypted_refresh_token` | Encrypted refresh token |
| `mapleapps_token_nonce` | Token encryption nonce |
| `mapleapps_access_token_expiry` | Access token expiry timestamp |
| `mapleapps_refresh_token_expiry` | Refresh token expiry timestamp |
| `mapleapps_user_email` | User's email address |
