# User Registration API Documentation

## Overview

This API endpoint allows new users to register for MapleApps services (MapleFile or PaperCloud). The registration process includes end-to-end encryption (E2EE) setup and email verification.

## Endpoint Details

**URL:** `POST /iam/api/v1/register`
**Authentication:** None required
**Content-Type:** `application/json`

## Request Structure

### Request Body

```json
{
  "beta_access_code": "string",
  "first_name": "string",
  "last_name": "string",
  "email": "string",
  "phone": "string",
  "country": "string",
  "timezone": "string",
  "agree_terms_of_service": true,
  "agree_promotions": false,
  "agree_to_tracking_across_third_party_apps_and_services": false,
  "module": 1,
  "salt": "string",
  "publicKey": "string",
  "encryptedMasterKey": "string",
  "encryptedPrivateKey": "string",
  "encryptedRecoveryKey": "string",
  "masterKeyEncryptedWithRecoveryKey": "string",
  "verificationID": "string"
}
```

### Field Descriptions

#### Personal Information Fields (Required)

| Field | Type | Required | Max Length | Description |
|-------|------|----------|------------|-------------|
| `beta_access_code` | string | Yes | - | Temporary code for beta access. Must match server configuration. |
| `first_name` | string | Yes | - | User's first name |
| `last_name` | string | Yes | - | User's last name |
| `email` | string | Yes | 255 chars | User's email address. Will be automatically lowercased and whitespace trimmed. |
| `phone` | string | Yes | - | User's phone number |
| `country` | string | Yes | - | User's country |
| `timezone` | string | Yes | - | User's timezone (e.g., "America/Toronto") |

#### Agreement Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `agree_terms_of_service` | boolean | Yes | Must be `true`. User must agree to terms of service. |
| `agree_promotions` | boolean | No | Whether user agrees to receive promotional communications |
| `agree_to_tracking_across_third_party_apps_and_services` | boolean | No | Whether user agrees to cross-platform tracking |

#### Module Selection

| Field | Type | Required | Valid Values | Description |
|-------|------|----------|--------------|-------------|
| `module` | integer | Yes | `1` or `2` | Service module: `1` = MapleFile, `2` = PaperCloud |

#### End-to-End Encryption Fields (Required)

| Field | Type | Required | Encoding | Description |
|-------|------|----------|----------|-------------|
| `salt` | string | Yes | Base64 URL | Password salt for key derivation |
| `publicKey` | string | Yes | Base64 URL | User's public key for asymmetric encryption |
| `encryptedMasterKey` | string | Yes | Base64 URL | Master key encrypted with key encryption key (nonce + ciphertext) |
| `encryptedPrivateKey` | string | Yes | Base64 URL | Private key encrypted with master key (nonce + ciphertext) |
| `encryptedRecoveryKey` | string | Yes | Base64 URL | Recovery key encrypted with master key (nonce + ciphertext) |
| `masterKeyEncryptedWithRecoveryKey` | string | Yes | Base64 URL | Master key encrypted with recovery key (nonce + ciphertext) |
| `verificationID` | string | No | - | Public key verification ID. Auto-generated if not provided. |

### Cryptographic Requirements

The API uses **ChaCha20-Poly1305** encryption with the following specifications:

- **Nonce Size:** 12 bytes
- **Key Derivation:** Argon2ID algorithm
- **Encoding:** All encrypted fields must be Base64 URL-encoded
- **Format:** Encrypted fields contain `nonce + ciphertext` concatenated

#### Sample Encryption Flow for Client Implementation

```javascript
// 1. Generate keys
const masterKey = crypto.getRandomValues(new Uint8Array(32));
const keyPair = await crypto.subtle.generateKey(
  { name: "X25519" },
  true,
  ["deriveKey"]
);

// 2. Derive key encryption key from password
const salt = crypto.getRandomValues(new Uint8Array(32));
const kek = await deriveKeyWithArgon2ID(password, salt);

// 3. Encrypt master key with KEK
const encryptedMasterKey = await encryptWithChaCha20Poly1305(masterKey, kek);

// 4. Export and encode
const publicKeyBytes = await crypto.subtle.exportKey("raw", keyPair.publicKey);
const publicKeyB64 = base64UrlEncode(publicKeyBytes);
const saltB64 = base64UrlEncode(salt);
```

## Response Structure

### Success Response (HTTP 201)

```json
{
  "message": "Registration successful. Please check your email for verification.",
  "recovery_key_info": "IMPORTANT: Please ensure you have saved your recovery key. It cannot be retrieved later."
}
```

### Error Responses

#### Validation Error (HTTP 400)

```json
{
  "error": "Bad Request",
  "details": {
    "field_name": "Error message for specific field",
    "another_field": "Another error message"
  }
}
```

**Common validation errors:**

| Field | Error Message |
|-------|---------------|
| `beta_access_code` | "Beta access code is required" or "Invalid beta access code" |
| `first_name` | "First name is required" |
| `last_name` | "Last name is required" |
| `email` | "Email is required", "Email is too long", or "Email address already exists" |
| `phone` | "Phone number is required" |
| `country` | "Country is required" |
| `timezone` | "Timezone is required" |
| `agree_terms_of_service` | "Agreeing to terms of service is required and you must agree to the terms before proceeding" |
| `module` | "Module is required" or "Module is invalid" |
| `salt` | "Salt is required" |
| `publicKey` | "Public key is required" |
| `encryptedMasterKey` | "Encrypted master key is required" |
| `encryptedPrivateKey` | "Encrypted private key is required" |
| `encryptedRecoveryKey` | "Encrypted recovery key is required" |
| `masterKeyEncryptedWithRecoveryKey` | "Master key encrypted with recovery key is required" |

#### Server Error (HTTP 500)

```json
{
  "error": "Internal Server Error",
  "message": "An unexpected error occurred"
}
```

## Implementation Examples

### React.js/React Native Example

```javascript
const registerUser = async (userData) => {
  try {
    const response = await fetch('https://api.mapleapps.ca/iam/api/v1/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        beta_access_code: "BETA2024",
        first_name: userData.firstName,
        last_name: userData.lastName,
        email: userData.email,
        phone: userData.phone,
        country: userData.country,
        timezone: "America/Toronto",
        agree_terms_of_service: true,
        agree_promotions: userData.agreePromotions || false,
        agree_to_tracking_across_third_party_apps_and_services: false,
        module: 1, // 1 for MapleFile, 2 for PaperCloud
        salt: userData.saltB64,
        publicKey: userData.publicKeyB64,
        encryptedMasterKey: userData.encryptedMasterKeyB64,
        encryptedPrivateKey: userData.encryptedPrivateKeyB64,
        encryptedRecoveryKey: userData.encryptedRecoveryKeyB64,
        masterKeyEncryptedWithRecoveryKey: userData.masterKeyEncryptedWithRecoveryKeyB64,
        verificationID: userData.verificationID // Optional
      })
    });

    if (response.ok) {
      const result = await response.json();
      console.log('Registration successful:', result.message);
      // Show recovery key warning to user
      alert(result.recovery_key_info);
      return result;
    } else {
      const error = await response.json();
      console.error('Registration failed:', error);
      throw new Error(error.message || 'Registration failed');
    }
  } catch (error) {
    console.error('Network error:', error);
    throw error;
  }
};
```

### Go Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type RegisterRequest struct {
    BetaAccessCode                                 string `json:"beta_access_code"`
    FirstName                                      string `json:"first_name"`
    LastName                                       string `json:"last_name"`
    Email                                          string `json:"email"`
    Phone                                          string `json:"phone"`
    Country                                        string `json:"country"`
    Timezone                                       string `json:"timezone"`
    AgreeTermsOfService                            bool   `json:"agree_terms_of_service"`
    AgreePromotions                                bool   `json:"agree_promotions"`
    AgreeToTrackingAcrossThirdPartyAppsAndServices bool   `json:"agree_to_tracking_across_third_party_apps_and_services"`
    Module                                         int    `json:"module"`
    Salt                                           string `json:"salt"`
    PublicKey                                      string `json:"publicKey"`
    EncryptedMasterKey                             string `json:"encryptedMasterKey"`
    EncryptedPrivateKey                            string `json:"encryptedPrivateKey"`
    EncryptedRecoveryKey                           string `json:"encryptedRecoveryKey"`
    MasterKeyEncryptedWithRecoveryKey              string `json:"masterKeyEncryptedWithRecoveryKey"`
    VerificationID                                 string `json:"verificationID,omitempty"`
}

type RegisterResponse struct {
    Message         string `json:"message"`
    RecoveryKeyInfo string `json:"recovery_key_info"`
}

func registerUser(req RegisterRequest) (*RegisterResponse, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    resp, err := http.Post(
        "https://api.mapleapps.ca/iam/api/v1/register",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusCreated {
        var result RegisterResponse
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            return nil, fmt.Errorf("failed to decode response: %w", err)
        }
        return &result, nil
    }

    var errorResp map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&errorResp)
    return nil, fmt.Errorf("registration failed: %v", errorResp)
}
```

### cURL Example

```bash
curl -X POST https://api.mapleapps.ca/iam/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "beta_access_code": "BETA2024",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "country": "Canada",
    "timezone": "America/Toronto",
    "agree_terms_of_service": true,
    "agree_promotions": false,
    "agree_to_tracking_across_third_party_apps_and_services": false,
    "module": 1,
    "salt": "base64url_encoded_salt",
    "publicKey": "base64url_encoded_public_key",
    "encryptedMasterKey": "base64url_encoded_encrypted_master_key",
    "encryptedPrivateKey": "base64url_encoded_encrypted_private_key",
    "encryptedRecoveryKey": "base64url_encoded_encrypted_recovery_key",
    "masterKeyEncryptedWithRecoveryKey": "base64url_encoded_master_key_with_recovery",
    "verificationID": "optional_verification_id"
  }'
```

## Important Notes

### Security Considerations

1. **Beta Access Code**: Currently required for all registrations. Contact MapleApps support for beta access.

2. **Email Verification**: After successful registration, users will receive an email verification code that must be used to activate their account.

3. **Recovery Key**: The recovery key is crucial for account recovery. Users must save it securely as it cannot be retrieved later.

4. **Client-Side Encryption**: All cryptographic operations should be performed client-side. The server never sees unencrypted user data.

### Email Processing

- Email addresses are automatically converted to lowercase
- All whitespace and tabs are stripped
- Maximum length is 255 characters

### Post-Registration Flow

1. User submits registration
2. Server creates user account with encrypted data
3. Server sends verification email to user
4. User must verify email using code sent to complete activation
5. User can then login to the system

### Module Selection

- **Module 1 (MapleFile)**: File storage and sharing service
- **Module 2 (PaperCloud)**: Document management service

Choose the appropriate module based on the service your application integrates with.
