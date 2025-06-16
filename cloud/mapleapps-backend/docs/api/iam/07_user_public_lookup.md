# User Public Lookup API Documentation

## Overview

This API endpoint allows clients to retrieve a user's public key and verification information by email address. This is typically used for end-to-end encryption scenarios where you need another user's public key to encrypt data for them.

## Endpoint Details

**URL:** `GET /iam/api/v1/users/lookup`
**Authentication:** None required
**Content-Type:** Not applicable (GET request)

## Request Structure

### Query Parameters

| Parameter | Type | Required | Max Length | Description |
|-----------|------|----------|------------|-------------|
| `email` | string | Yes | 255 chars | Email address of the user to lookup. Will be automatically lowercased and whitespace trimmed. |

### URL Format

```
GET /iam/api/v1/users/lookup?email={email_address}
```

## Response Structure

### Success Response (HTTP 200)

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "public_key_in_base64": "base64_encoded_public_key",
  "verification_id": "verification_id_string"
}
```

### Field Descriptions (Response)

| Field | Type | Description |
|-------|------|-------------|
| `user_id` | string | User's unique identifier (UUID format) |
| `email` | string | User's email address (sanitized) |
| `name` | string | User's full name for display purposes |
| `public_key_in_base64` | string | Base64-encoded public key for encryption |
| `verification_id` | string | Public key verification identifier |

### Error Responses

#### Missing Email Parameter (HTTP 400)

```json
{
  "error": "email parameter required"
}
```

#### Invalid Email Format (HTTP 400)

```json
{
  "error": "invalid email format"
}
```

#### Validation Errors (HTTP 400)

```json
{
  "error": "Bad Request",
  "details": {
    "email": "Email is required"
  }
}
```

#### User Not Found (HTTP 400)

```json
{
  "error": "Bad Request",
  "details": {
    "email": "Email address does not exist: user@example.com"
  }
}
```

**Common validation errors:**
- `"Email is required"` - Empty email parameter
- `"Email is too long"` - Email exceeds 255 characters
- `"Email address does not exist: {email}"` - User not found in system

## Implementation Examples

### React.js/React Native

```javascript
const lookupUserPublicKey = async (email) => {
  try {
    // URL encode the email parameter
    const encodedEmail = encodeURIComponent(email.toLowerCase().trim());
    const url = `https://api.mapleapps.ca/iam/api/v1/users/lookup?email=${encodedEmail}`;

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      }
    });

    if (response.ok) {
      const result = await response.json();
      console.log('User found:', result.name);
      console.log('Public key available for encryption');

      // Decode the public key for use
      const publicKeyBytes = base64ToBytes(result.public_key_in_base64);

      return {
        userId: result.user_id,
        email: result.email,
        name: result.name,
        publicKey: publicKeyBytes,
        verificationId: result.verification_id
      };
    } else {
      const error = await response.text();
      throw new Error(error || 'Failed to lookup user');
    }
  } catch (error) {
    console.error('User lookup error:', error);
    throw error;
  }
};

// Utility function to convert base64 to bytes
const base64ToBytes = (base64String) => {
  const binaryString = atob(base64String);
  const bytes = new Uint8Array(binaryString.length);
  for (let i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes;
};

// Usage example
const encryptForUser = async (recipientEmail, message) => {
  try {
    // Lookup recipient's public key
    const recipient = await lookupUserPublicKey(recipientEmail);

    // Encrypt message with recipient's public key
    const encryptedMessage = await encryptWithPublicKey(message, recipient.publicKey);

    return {
      recipientId: recipient.userId,
      recipientName: recipient.name,
      encryptedData: encryptedMessage
    };
  } catch (error) {
    console.error('Failed to encrypt for user:', error);
    throw error;
  }
};
```

### Go

```go
package main

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
)

type UserPublicLookupResponse struct {
    UserID            string `json:"user_id"`
    Email             string `json:"email"`
    Name              string `json:"name"`
    PublicKeyInBase64 string `json:"public_key_in_base64"`
    VerificationID    string `json:"verification_id"`
}

type UserPublicInfo struct {
    UserID         string
    Email          string
    Name           string
    PublicKey      []byte
    VerificationID string
}

func lookupUserPublicKey(email string) (*UserPublicInfo, error) {
    // Sanitize and encode email
    email = strings.ToLower(strings.TrimSpace(email))
    encodedEmail := url.QueryEscape(email)

    apiURL := fmt.Sprintf("https://api.mapleapps.ca/iam/api/v1/users/lookup?email=%s", encodedEmail)

    resp, err := http.Get(apiURL)
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        var result UserPublicLookupResponse
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            return nil, fmt.Errorf("failed to decode response: %w", err)
        }

        // Decode the public key
        publicKey, err := base64.StdEncoding.DecodeString(result.PublicKeyInBase64)
        if err != nil {
            return nil, fmt.Errorf("failed to decode public key: %w", err)
        }

        return &UserPublicInfo{
            UserID:         result.UserID,
            Email:          result.Email,
            Name:           result.Name,
            PublicKey:      publicKey,
            VerificationID: result.VerificationID,
        }, nil
    }

    // Handle error response
    var errorBody map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&errorBody)
    return nil, fmt.Errorf("user lookup failed: %v", errorBody)
}

// Usage example
func encryptForUser(recipientEmail, message string) (*EncryptedMessage, error) {
    // Lookup recipient's public key
    recipient, err := lookupUserPublicKey(recipientEmail)
    if err != nil {
        return nil, fmt.Errorf("failed to lookup user: %w", err)
    }

    // Encrypt message with recipient's public key
    encryptedData, err := encryptWithPublicKey(message, recipient.PublicKey)
    if err != nil {
        return nil, fmt.Errorf("failed to encrypt message: %w", err)
    }

    return &EncryptedMessage{
        RecipientID:   recipient.UserID,
        RecipientName: recipient.Name,
        EncryptedData: encryptedData,
    }, nil
}
```

### Python

```python
import requests
import base64
from urllib.parse import quote
from typing import Optional, Dict

class UserPublicInfo:
    def __init__(self, user_id: str, email: str, name: str, public_key: bytes, verification_id: str):
        self.user_id = user_id
        self.email = email
        self.name = name
        self.public_key = public_key
        self.verification_id = verification_id

def lookup_user_public_key(email: str) -> Optional[UserPublicInfo]:
    """Lookup a user's public key information by email address."""
    # Sanitize and encode email
    email = email.lower().strip()
    encoded_email = quote(email)

    url = f"https://api.mapleapps.ca/iam/api/v1/users/lookup?email={encoded_email}"

    try:
        response = requests.get(url)

        if response.status_code == 200:
            data = response.json()

            # Decode the public key
            public_key = base64.b64decode(data['public_key_in_base64'])

            return UserPublicInfo(
                user_id=data['user_id'],
                email=data['email'],
                name=data['name'],
                public_key=public_key,
                verification_id=data['verification_id']
            )
        else:
            error_data = response.json() if response.headers.get('content-type') == 'application/json' else response.text
            raise Exception(f"User lookup failed: {error_data}")

    except requests.RequestException as e:
        raise Exception(f"Network error during user lookup: {str(e)}")

def encrypt_for_user(recipient_email: str, message: str) -> Dict:
    """Encrypt a message for a specific user."""
    try:
        # Lookup recipient's public key
        recipient = lookup_user_public_key(recipient_email)

        if not recipient:
            raise Exception("User not found")

        # Encrypt message with recipient's public key
        encrypted_data = encrypt_with_public_key(message, recipient.public_key)

        return {
            'recipient_id': recipient.user_id,
            'recipient_name': recipient.name,
            'encrypted_data': encrypted_data
        }

    except Exception as e:
        print(f"Failed to encrypt for user: {e}")
        raise

# Usage example
if __name__ == "__main__":
    try:
        user_info = lookup_user_public_key("john.doe@example.com")
        if user_info:
            print(f"Found user: {user_info.name}")
            print(f"Public key length: {len(user_info.public_key)} bytes")
        else:
            print("User not found")
    except Exception as e:
        print(f"Error: {e}")
```

### cURL Examples

#### Basic Lookup
```bash
curl -X GET "https://api.mapleapps.ca/iam/api/v1/users/lookup?email=user@example.com" \
  -H "Content-Type: application/json"
```

#### URL Encoded Email
```bash
curl -X GET "https://api.mapleapps.ca/iam/api/v1/users/lookup?email=user%40example.com" \
  -H "Content-Type: application/json"
```

#### With Complex Email
```bash
curl -X GET "https://api.mapleapps.ca/iam/api/v1/users/lookup?email=user.name%2Btag%40domain.com" \
  -H "Content-Type: application/json"
```

## Important Notes

### Security Considerations

1. **Public Information Only**: This endpoint only returns public information that's safe to expose
2. **No Authentication Required**: Anyone can lookup public keys (by design for encryption)
3. **Rate Limiting**: Consider implementing rate limiting to prevent abuse
4. **Email Enumeration**: The endpoint reveals whether an email exists in the system

### Use Cases

1. **End-to-End Encryption**: Getting recipient's public key before encrypting data
2. **Key Verification**: Verifying public key ownership using verification ID
3. **User Discovery**: Finding users for sharing encrypted content
4. **Contact Lists**: Building encrypted messaging contact lists

### Input Sanitization

The API automatically performs the following sanitization:
- Converts email to lowercase
- Removes leading/trailing whitespace
- Removes tabs and spaces within the email

### Best Practices

1. **URL Encoding**: Always URL-encode email parameters to handle special characters
2. **Public Key Validation**: Verify the public key format and length before use
3. **Verification ID**: Use verification ID to ensure public key authenticity
4. **Error Handling**: Handle "user not found" scenarios gracefully
5. **Caching**: Consider caching public key lookups for performance (with appropriate TTL)

### Integration Patterns

#### Batch User Lookup
```javascript
const lookupMultipleUsers = async (emails) => {
  const promises = emails.map(email => lookupUserPublicKey(email));
  const results = await Promise.allSettled(promises);

  return results.map((result, index) => ({
    email: emails[index],
    success: result.status === 'fulfilled',
    data: result.status === 'fulfilled' ? result.value : null,
    error: result.status === 'rejected' ? result.reason : null
  }));
};
```

#### Contact Validation
```javascript
const validateContacts = async (emailList) => {
  const validContacts = [];
  const invalidContacts = [];

  for (const email of emailList) {
    try {
      const user = await lookupUserPublicKey(email);
      validContacts.push(user);
    } catch (error) {
      invalidContacts.push({ email, error: error.message });
    }
  }

  return { validContacts, invalidContacts };
};
```

This API enables secure end-to-end encryption workflows by providing easy access to users' public key information while maintaining privacy and security standards.
