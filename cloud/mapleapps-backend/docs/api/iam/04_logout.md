# MapleApps Logout API Documentation

## Logout API

**URL:** `POST /iam/api/v1/logout`
**Authentication:** Required (JWT token)
**Content-Type:** `application/json`

### Request Structure

No request body required.

### Headers

| Header | Value | Required | Description |
|--------|-------|----------|-------------|
| `Authorization` | `JWT <token>` | Yes | Access token received from login |
| `Content-Type` | `application/json` | Yes | Content type header |

### Response Structure

#### Success Response (HTTP 204)

No content returned. HTTP 204 No Content indicates successful logout.

#### Error Responses

**HTTP 401 Unauthorized:**
```json
{
  "error": "Unauthorized",
  "message": "attempting to access a protected endpoint and authorization not set"
}
```

**HTTP 400 Bad Request:**
```json
{
  "error": "Bad Request",
  "details": {
    "session_id": "not logged in"
  }
}
```

### Implementation Examples

#### React.js/React Native
```javascript
const logout = async () => {
  try {
    const token = localStorage.getItem('access_token');
    if (!token) {
      throw new Error('No access token found');
    }

    const response = await fetch('https://api.mapleapps.ca/iam/api/v1/logout', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `JWT ${token}`
      }
    });

    if (response.status === 204) {
      console.log('Logout successful');
      // Clear stored tokens
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      return true;
    } else {
      const error = await response.json();
      throw new Error(error.message || 'Failed to logout');
    }
  } catch (error) {
    console.error('Logout error:', error);
    // Clear tokens anyway on error
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    throw error;
  }
};
```

#### Go
```go
func logout(accessToken string) error {
    req, err := http.NewRequest("POST", "https://api.mapleapps.ca/iam/api/v1/logout", nil)
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("JWT %s", accessToken))

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNoContent {
        return nil // Success
    }

    return fmt.Errorf("logout failed with status: %d", resp.StatusCode)
}
```

#### cURL Example
```bash
curl -X POST https://api.mapleapps.ca/iam/api/v1/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: JWT your_access_token_here"
```
