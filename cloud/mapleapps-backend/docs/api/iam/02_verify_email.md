# MapleApps Post-Registration Verification API Documentation
## Verify Email API

**URL:** `POST /iam/api/v1/verify-email-code`
**Authentication:** None required
**Content-Type:** `application/json`

### Request Structure

```json
{
  "code": "string"
}
```

### Field Descriptions

| Field | Type | Required | Max Length | Description |
|-------|------|----------|------------|-------------|
| `code` | string | Yes | - | 6-digit verification code received via email during registration |

### Response Structure

#### Success Response (HTTP 201)

```json
{
  "message": "Thank you for verifying. You may log in now to get started!",
  "user_role": 3
}
```

### Field Descriptions (Response)

| Field | Type | Description |
|-------|------|-------------|
| `message` | string | Success message confirming email verification |
| `user_role` | integer | User's role: `1` = Root, `2` = Company, `3` = Individual |

#### Error Responses (HTTP 400)

```json
{
  "error": "Bad Request",
  "details": {
    "code": "does not exist"
  }
}
```

**Common validation errors:**
- `"payload structure is wrong"` - Invalid JSON format
- `"does not exist"` - Verification code not found or expired

### Implementation Examples

#### React.js/React Native
```javascript
const verifyEmail = async (verificationCode) => {
  try {
    const response = await fetch('https://api.mapleapps.ca/iam/api/v1/verify-email-code', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        code: verificationCode.trim()
      })
    });

    if (response.status === 201) {
      const result = await response.json();
      console.log('Email verified:', result.message);
      console.log('User role:', result.user_role);
      return result;
    } else {
      const error = await response.json();
      throw new Error(error.details?.code || 'Failed to verify email');
    }
  } catch (error) {
    console.error('Email verification error:', error);
    throw error;
  }
};
```

#### Go
```go
type VerifyEmailRequest struct {
    Code string `json:"code"`
}

type VerifyEmailResponse struct {
    Message  string `json:"message"`
    UserRole int8   `json:"user_role"`
}

func verifyEmail(code string) (*VerifyEmailResponse, error) {
    req := VerifyEmailRequest{
        Code: strings.TrimSpace(code),
    }

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(
        "https://api.mapleapps.ca/iam/api/v1/verify-email-code",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusCreated {
        var result VerifyEmailResponse
        json.NewDecoder(resp.Body).Decode(&result)
        return &result, nil
    }

    return nil, fmt.Errorf("failed to verify email: %d", resp.StatusCode)
}
```

#### cURL Example
```bash
curl -X POST https://api.mapleapps.ca/iam/api/v1/verify-email-code \
  -H "Content-Type: application/json" \
  -d '{
    "code": "123456"
  }'
```
