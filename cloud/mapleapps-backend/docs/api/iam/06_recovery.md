# Account Recovery API Documentation

## Overview

The Account Recovery API provides a secure 3-step process for users to regain access to their accounts when they've forgotten their password but still have access to their recovery key. The entire process maintains end-to-end encryption.

## Recovery Flow (3-Step Process)

1. **Initiate Recovery** - Start recovery process and receive encrypted challenge
2. **Verify Recovery** - Decrypt challenge to prove recovery key ownership
3. **Complete Recovery** - Set new password and encryption keys

---

## Step 1: Initiate Account Recovery

**URL:** `POST /iam/api/v1/recovery/initiate`
**Authentication:** None required
**Content-Type:** `application/json`

### Request Structure

```json
{
  "email": "string",
  "method": "recovery_key"
}
```

### Field Descriptions

| Field | Type | Required | Valid Values | Description |
|-------|------|----------|--------------|-------------|
| `email` | string | Yes | - | User's email address. Will be automatically lowercased and whitespace trimmed. |
| `method` | string | No | `"recovery_key"` | Recovery method. Defaults to `"recovery_key"` if not specified. |

### Response Structure

#### Success Response (HTTP 200)

```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440001",
  "challenge_id": "550e8400-e29b-41d4-a716-446655440002",
  "encrypted_challenge": "base64_encoded_challenge",
  "expires_in": 600
}
```

### Field Descriptions (Response)

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | Unique session identifier for this recovery attempt |
| `challenge_id` | string | Unique challenge identifier |
| `encrypted_challenge` | string | Base64-encoded challenge encrypted with user's public key |
| `expires_in` | integer | Session expiry time in seconds (typically 600 = 10 minutes) |

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
- `"If this email exists, recovery instructions will be sent."` - Email not found (intentionally vague for security)
- `"Recovery key not configured for this account"` - User has no recovery key set up
- `"Too many recovery attempts. Please try again later."` - Rate limiting (5 attempts per 15 minutes)

### Implementation Examples

#### React.js/React Native
```javascript
const initiateRecovery = async (email) => {
  try {
    const response = await fetch('https://api.mapleapps.ca/iam/api/v1/recovery/initiate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email: email.toLowerCase().trim(),
        method: 'recovery_key'
      })
    });

    if (response.ok) {
      const result = await response.json();
      console.log('Recovery initiated:', result.session_id);
      return result;
    } else {
      const error = await response.json();
      throw new Error(error.details?.email || 'Failed to initiate recovery');
    }
  } catch (error) {
    console.error('Recovery initiation error:', error);
    throw error;
  }
};
```

#### Go
```go
type InitiateRecoveryRequest struct {
    Email  string `json:"email"`
    Method string `json:"method,omitempty"`
}

type InitiateRecoveryResponse struct {
    SessionID          string `json:"session_id"`
    ChallengeID        string `json:"challenge_id"`
    EncryptedChallenge string `json:"encrypted_challenge"`
    ExpiresIn          int    `json:"expires_in"`
}

func initiateRecovery(email string) (*InitiateRecoveryResponse, error) {
    req := InitiateRecoveryRequest{
        Email:  strings.ToLower(strings.TrimSpace(email)),
        Method: "recovery_key",
    }

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(
        "https://api.mapleapps.ca/iam/api/v1/recovery/initiate",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        var result InitiateRecoveryResponse
        json.NewDecoder(resp.Body).Decode(&result)
        return &result, nil
    }

    return nil, fmt.Errorf("failed to initiate recovery: %d", resp.StatusCode)
}
```

---

## Step 2: Verify Recovery Challenge

**URL:** `POST /iam/api/v1/recovery/verify`
**Authentication:** None required
**Content-Type:** `application/json`

### Request Structure

```json
{
  "session_id": "string",
  "decrypted_challenge": "string"
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `session_id` | string | Yes | Session ID from step 1 |
| `decrypted_challenge` | string | Yes | Base64-encoded challenge decrypted with recovery key |

### Response Structure

#### Success Response (HTTP 200)

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "recovery_token": "base64_encoded_token",
  "master_key_encrypted_with_recovery_key": "base64_encoded_data",
  "expires_in": 600
}
```

### Field Descriptions (Response)

| Field | Type | Description |
|-------|------|-------------|
| `user_id` | string | User's unique identifier |
| `email` | string | User's email address |
| `recovery_token` | string | Base64-encoded token for completing recovery |
| `master_key_encrypted_with_recovery_key` | string | Base64-encoded master key encrypted with recovery key |
| `expires_in` | integer | Token expiry time in seconds |

#### Error Responses (HTTP 400)

```json
{
  "error": "Bad Request",
  "details": {
    "session_id": "Invalid or expired session"
  }
}
```

**Common validation errors:**
- `"Session ID is required"` - Missing session_id
- `"Decrypted challenge is required"` - Missing decrypted_challenge
- `"Invalid or expired session"` - Session not found or expired
- `"Session has expired"` - Session timeout (10 minutes)
- `"Session already verified"` - Session already used
- `"Invalid challenge format"` - Malformed decrypted challenge
- `"Invalid challenge response"` - Challenge verification failed

### Implementation Examples

#### React.js/React Native
```javascript
const verifyRecovery = async (sessionId, decryptedChallenge) => {
  try {
    const response = await fetch('https://api.mapleapps.ca/iam/api/v1/recovery/verify', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        session_id: sessionId,
        decrypted_challenge: decryptedChallenge
      })
    });

    if (response.ok) {
      const result = await response.json();
      console.log('Recovery verified, received recovery token');
      return result;
    } else {
      const error = await response.json();
      throw new Error(error.details?.session_id || error.details?.decrypted_challenge || 'Failed to verify recovery');
    }
  } catch (error) {
    console.error('Recovery verification error:', error);
    throw error;
  }
};
```

#### Go
```go
type VerifyRecoveryRequest struct {
    SessionID          string `json:"session_id"`
    DecryptedChallenge string `json:"decrypted_challenge"`
}

type VerifyRecoveryResponse struct {
    UserID                   string `json:"user_id"`
    Email                    string `json:"email"`
    RecoveryToken            string `json:"recovery_token"`
    MasterKeyWithRecoveryKey string `json:"master_key_encrypted_with_recovery_key"`
    ExpiresIn                int    `json:"expires_in"`
}

func verifyRecovery(sessionID, decryptedChallenge string) (*VerifyRecoveryResponse, error) {
    req := VerifyRecoveryRequest{
        SessionID:          sessionID,
        DecryptedChallenge: decryptedChallenge,
    }

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(
        "https://api.mapleapps.ca/iam/api/v1/recovery/verify",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        var result VerifyRecoveryResponse
        json.NewDecoder(resp.Body).Decode(&result)
        return &result, nil
    }

    return nil, fmt.Errorf("failed to verify recovery: %d", resp.StatusCode)
}
```

---

## Step 3: Complete Account Recovery

**URL:** `POST /iam/api/v1/recovery/complete`
**Authentication:** None required
**Content-Type:** `application/json`

### Request Structure

```json
{
  "recovery_token": "string",
  "new_salt": "string",
  "new_encrypted_master_key": "string",
  "new_encrypted_private_key": "string",
  "new_encrypted_recovery_key": "string",
  "new_master_key_encrypted_with_recovery_key": "string"
}
```

### Field Descriptions

| Field | Type | Required | Encoding | Description |
|-------|------|----------|----------|-------------|
| `recovery_token` | string | Yes | Base64 URL | Recovery token from step 2 |
| `new_salt` | string | Yes | Base64 URL | New password salt for key derivation |
| `new_encrypted_master_key` | string | Yes | Base64 URL | New master key encrypted with new KEK (nonce + ciphertext) |
| `new_encrypted_private_key` | string | Yes | Base64 URL | New private key encrypted with new master key (nonce + ciphertext) |
| `new_encrypted_recovery_key` | string | Yes | Base64 URL | New recovery key encrypted with new master key (nonce + ciphertext) |
| `new_master_key_encrypted_with_recovery_key` | string | Yes | Base64 URL | New master key encrypted with new recovery key (nonce + ciphertext) |

### Cryptographic Requirements

All encrypted fields use **ChaCha20-Poly1305** encryption with:
- **Nonce Size:** 12 bytes
- **Format:** `nonce + ciphertext` concatenated
- **Encoding:** Base64 URL encoding

### Response Structure

#### Success Response (HTTP 200)

```json
{
  "success": true,
  "message": "Account recovery completed successfully. You can now log in with your new password."
}
```

### Field Descriptions (Response)

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | Whether recovery was completed successfully |
| `message` | string | Human-readable success message |

#### Error Responses (HTTP 400)

```json
{
  "error": "Bad Request",
  "details": {
    "recovery_token": "Recovery token is required"
  }
}
```

**Common validation errors:**
- `"Recovery token is required"` - Missing recovery_token
- `"New salt is required"` - Missing new_salt
- `"New encrypted master key is required"` - Missing new_encrypted_master_key
- `"New encrypted private key is required"` - Missing new_encrypted_private_key
- `"New encrypted recovery key is required"` - Missing new_encrypted_recovery_key
- `"New master key encrypted with recovery key is required"` - Missing new_master_key_encrypted_with_recovery_key
- `"Invalid recovery token"` - Token invalid or expired
- `"Recovery session not verified"` - Session not verified in step 2
- `"Recovery session expired"` - Session timeout
- `"Invalid [field] format"` - Malformed Base64 encoding
- `"[Field] too short"` - Encrypted data shorter than nonce size

### Implementation Examples

#### React.js/React Native
```javascript
const completeRecovery = async (recoveryToken, newCryptoData) => {
  try {
    const response = await fetch('https://api.mapleapps.ca/iam/api/v1/recovery/complete', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        recovery_token: recoveryToken,
        new_salt: newCryptoData.saltB64,
        new_encrypted_master_key: newCryptoData.encryptedMasterKeyB64,
        new_encrypted_private_key: newCryptoData.encryptedPrivateKeyB64,
        new_encrypted_recovery_key: newCryptoData.encryptedRecoveryKeyB64,
        new_master_key_encrypted_with_recovery_key: newCryptoData.masterKeyWithRecoveryKeyB64
      })
    });

    if (response.ok) {
      const result = await response.json();
      console.log('Recovery completed:', result.message);
      return result;
    } else {
      const error = await response.json();
      throw new Error(error.details?.recovery_token || 'Failed to complete recovery');
    }
  } catch (error) {
    console.error('Recovery completion error:', error);
    throw error;
  }
};
```

#### Go
```go
type CompleteRecoveryRequest struct {
    RecoveryToken               string `json:"recovery_token"`
    NewSalt                     string `json:"new_salt"`
    NewEncryptedMasterKey       string `json:"new_encrypted_master_key"`
    NewEncryptedPrivateKey      string `json:"new_encrypted_private_key"`
    NewEncryptedRecoveryKey     string `json:"new_encrypted_recovery_key"`
    NewMasterKeyWithRecoveryKey string `json:"new_master_key_encrypted_with_recovery_key"`
}

type CompleteRecoveryResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

func completeRecovery(req CompleteRecoveryRequest) (*CompleteRecoveryResponse, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(
        "https://api.mapleapps.ca/iam/api/v1/recovery/complete",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        var result CompleteRecoveryResponse
        json.NewDecoder(resp.Body).Decode(&result)
        return &result, nil
    }

    return nil, fmt.Errorf("failed to complete recovery: %d", resp.StatusCode)
}
```

### cURL Examples

#### Step 1: Initiate Recovery
```bash
curl -X POST https://api.mapleapps.ca/iam/api/v1/recovery/initiate \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "method": "recovery_key"
  }'
```

#### Step 2: Verify Recovery
```bash
curl -X POST https://api.mapleapps.ca/iam/api/v1/recovery/verify \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440001",
    "decrypted_challenge": "base64_url_encoded_decrypted_challenge"
  }'
```

#### Step 3: Complete Recovery
```bash
curl -X POST https://api.mapleapps.ca/iam/api/v1/recovery/complete \
  -H "Content-Type: application/json" \
  -d '{
    "recovery_token": "base64_url_encoded_token",
    "new_salt": "base64_url_encoded_salt",
    "new_encrypted_master_key": "base64_url_encoded_master_key",
    "new_encrypted_private_key": "base64_url_encoded_private_key",
    "new_encrypted_recovery_key": "base64_url_encoded_recovery_key",
    "new_master_key_encrypted_with_recovery_key": "base64_url_encoded_master_with_recovery"
  }'
```

## Important Notes

### Security Considerations

1. **Rate Limiting**: Maximum 5 recovery attempts per email address within 15 minutes
2. **Session Timeout**: Recovery sessions expire after 10 minutes of inactivity
3. **End-to-End Encryption**: All cryptographic operations performed client-side
4. **Audit Trail**: All recovery attempts are logged for security monitoring
5. **Key Rotation**: Successful recovery increments key version and records rotation

### Prerequisites

1. **Recovery Key Setup**: User must have configured recovery key during registration
2. **Recovery Key Access**: User must have access to their recovery key to decrypt challenge
3. **Client-Side Crypto**: Application must support ChaCha20-Poly1305 and X25519 operations

### Recovery Process Flow

1. **User initiates recovery** → Server generates encrypted challenge
2. **User decrypts challenge** with recovery key → Proves recovery key ownership
3. **User creates new password** → Generates new encryption keys
4. **Server updates account** → New keys replace old ones, user can login

### Error Handling

- Failed attempts are logged and count toward rate limiting
- Generic error messages prevent user enumeration attacks
- Sessions automatically expire to prevent abuse
- Clear validation messages help with legitimate recovery attempts

### Post-Recovery

- User can immediately log in with new password
- All previous sessions are invalidated
- Key version is incremented for forward security
- Recovery key may need to be re-saved securely

This recovery system ensures users can regain access to their accounts while maintaining the security and privacy guarantees of end-to-end encryption.
