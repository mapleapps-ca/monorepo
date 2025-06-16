# MapleFile Collections API Documentation

## Overview

The MapleFile Collections API provides a comprehensive interface for managing hierarchical, encrypted file collections (folders and albums) with advanced sharing capabilities. This API supports end-to-end encryption, user access control, and real-time synchronization.

### Base URL
```
https://maplefile.ca/maplefile/api/v1
```

### Authentication
All endpoints require authentication via session-based authentication. The user ID is automatically extracted from the session context.

### Content Type
All requests and responses use `application/json` content type.

---

## Data Models

### Collection Types
- `folder` - Standard folder for organizing files
- `album` - Photo/media album with specialized features

### Collection States
- `active` - Collection is available for use
- `deleted` - Collection is soft-deleted (recoverable)
- `archived` - Collection is archived (read-only)

### Permission Levels
- `read_only` - Can view collection and files
- `read_write` - Can add/modify files and subcollections
- `admin` - Full control including sharing and deletion

---

## API Endpoints

### 1. Create Collection

**Endpoint:** `POST /collections`

Creates a new collection (folder or album) with optional hierarchy and sharing settings.

**Request Body:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "owner_id": "550e8400-e29b-41d4-a716-446655440001",
  "encrypted_name": "base64_encrypted_collection_name",
  "collection_type": "folder",
  "encrypted_collection_key": {
    "ciphertext": [1, 2, 3, 4, 5],
    "nonce": [6, 7, 8, 9, 10],
    "key_version": 1,
    "rotated_at": "2023-01-01T00:00:00Z",
    "previous_keys": []
  },
  "members": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "collection_id": "550e8400-e29b-41d4-a716-446655440000",
      "recipient_id": "550e8400-e29b-41d4-a716-446655440003",
      "recipient_email": "user@example.com",
      "granted_by_id": "550e8400-e29b-41d4-a716-446655440001",
      "encrypted_collection_key": [11, 12, 13, 14, 15],
      "permission_level": "read_write",
      "created_at": "2023-01-01T00:00:00Z",
      "is_inherited": false,
      "inherited_from_id": "00000000-0000-0000-0000-000000000000"
    }
  ],
  "parent_id": "550e8400-e29b-41d4-a716-446655440004",
  "ancestor_ids": [
    "550e8400-e29b-41d4-a716-446655440005",
    "550e8400-e29b-41d4-a716-446655440004"
  ],
  "created_at": "2023-01-01T00:00:00Z",
  "created_by_user_id": "550e8400-e29b-41d4-a716-446655440001",
  "modified_at": "2023-01-01T00:00:00Z",
  "modified_by_user_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Request Fields:**
- `id` (UUID, optional): Client-generated ID (server will override)
- `owner_id` (UUID, optional): Owner ID (server will override with authenticated user)
- `encrypted_name` (string, required): Base64 encrypted collection name
- `collection_type` (string, required): Either "folder" or "album"
- `encrypted_collection_key` (object, required): Encrypted key for collection data
  - `ciphertext` (byte array, required): Encrypted key data
  - `nonce` (byte array, required): Encryption nonce
  - `key_version` (integer): Key version for rotation
  - `rotated_at` (timestamp): When key was last rotated
  - `previous_keys` (array): Historical keys for decryption
- `members` (array, optional): Initial sharing configuration
- `parent_id` (UUID, optional): Parent collection ID for hierarchy
- `ancestor_ids` (array, optional): Array of ancestor collection IDs
- `created_at` (timestamp, optional): Client timestamp (server will override)
- `created_by_user_id` (UUID, optional): Creator ID (server will override)
- `modified_at` (timestamp, optional): Modification timestamp (server will override)
- `modified_by_user_id` (UUID, optional): Modifier ID (server will override)

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "owner_id": "550e8400-e29b-41d4-a716-446655440001",
  "encrypted_name": "base64_encrypted_collection_name",
  "collection_type": "folder",
  "parent_id": "550e8400-e29b-41d4-a716-446655440004",
  "ancestor_ids": [
    "550e8400-e29b-41d4-a716-446655440005",
    "550e8400-e29b-41d4-a716-446655440004"
  ],
  "encrypted_collection_key": {
    "ciphertext": [1, 2, 3, 4, 5],
    "nonce": [6, 7, 8, 9, 10],
    "key_version": 1,
    "rotated_at": "2023-01-01T00:00:00Z",
    "previous_keys": []
  },
  "created_at": "2023-01-01T00:00:00Z",
  "modified_at": "2023-01-01T00:00:00Z",
  "members": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "collection_id": "550e8400-e29b-41d4-a716-446655440000",
      "recipient_id": "550e8400-e29b-41d4-a716-446655440001",
      "recipient_email": "owner@example.com",
      "granted_by_id": "550e8400-e29b-41d4-a716-446655440001",
      "encrypted_collection_key": [1, 2, 3, 4, 5],
      "permission_level": "admin",
      "created_at": "2023-01-01T00:00:00Z",
      "is_inherited": false,
      "inherited_from_id": "00000000-0000-0000-0000-000000000000"
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Collection created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required
- `500 Internal Server Error` - Server error

---

### 2. Get Collection

**Endpoint:** `GET /collections/{collection_id}`

Retrieves a specific collection by ID if the user has access.

**Path Parameters:**
- `collection_id` (UUID, required): The collection ID

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "owner_id": "550e8400-e29b-41d4-a716-446655440001",
  "encrypted_name": "base64_encrypted_collection_name",
  "collection_type": "folder",
  "parent_id": "550e8400-e29b-41d4-a716-446655440004",
  "ancestor_ids": [
    "550e8400-e29b-41d4-a716-446655440005",
    "550e8400-e29b-41d4-a716-446655440004"
  ],
  "encrypted_collection_key": {
    "ciphertext": [1, 2, 3, 4, 5],
    "nonce": [6, 7, 8, 9, 10],
    "key_version": 1,
    "rotated_at": "2023-01-01T00:00:00Z",
    "previous_keys": []
  },
  "created_at": "2023-01-01T00:00:00Z",
  "modified_at": "2023-01-01T00:00:00Z",
  "members": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "collection_id": "550e8400-e29b-41d4-a716-446655440000",
      "recipient_id": "550e8400-e29b-41d4-a716-446655440001",
      "recipient_email": "owner@example.com",
      "granted_by_id": "550e8400-e29b-41d4-a716-446655440001",
      "encrypted_collection_key": [1, 2, 3, 4, 5],
      "permission_level": "admin",
      "created_at": "2023-01-01T00:00:00Z",
      "is_inherited": false,
      "inherited_from_id": "00000000-0000-0000-0000-000000000000"
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Collection retrieved successfully
- `403 Forbidden` - User doesn't have access to this collection
- `404 Not Found` - Collection not found
- `401 Unauthorized` - Authentication required

---

### 3. Update Collection

**Endpoint:** `PUT /collections/{collection_id}`

Updates collection metadata. Requires admin permissions.

**Path Parameters:**
- `collection_id` (UUID, required): The collection ID

**Request Body:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "encrypted_name": "new_base64_encrypted_name",
  "collection_type": "album",
  "encrypted_collection_key": {
    "ciphertext": [1, 2, 3, 4, 5],
    "nonce": [6, 7, 8, 9, 10],
    "key_version": 2,
    "rotated_at": "2023-01-01T00:00:00Z",
    "previous_keys": []
  },
  "version": 1
}
```

**Request Fields:**
- `id` (UUID, required): Collection ID (must match path parameter)
- `encrypted_name` (string, required): Updated encrypted collection name
- `collection_type` (string, optional): Updated collection type
- `encrypted_collection_key` (object, optional): Updated encryption key
- `version` (integer, required): Current version for optimistic locking

**Response:** Same format as Get Collection response

**Status Codes:**
- `200 OK` - Collection updated successfully
- `400 Bad Request` - Invalid request data or version conflict
- `403 Forbidden` - User doesn't have admin permission
- `404 Not Found` - Collection not found

---

### 4. Delete Collection (Soft Delete)

**Endpoint:** `DELETE /collections/{collection_id}`

Soft deletes a collection, making it recoverable for 30 days.

**Path Parameters:**
- `collection_id` (UUID, required): The collection ID

**Response:**
```json
{
  "success": true,
  "message": "Collection, descendants, and all associated files soft-deleted successfully"
}
```

**Status Codes:**
- `200 OK` - Collection deleted successfully
- `403 Forbidden` - Only collection owner can delete
- `404 Not Found` - Collection not found

---

### 5. Archive Collection

**Endpoint:** `POST /collections/{collection_id}/archive`

Archives a collection, making it read-only.

**Path Parameters:**
- `collection_id` (UUID, required): The collection ID

**Response:**
```json
{
  "success": true,
  "message": "Collection archived successfully"
}
```

**Status Codes:**
- `200 OK` - Collection archived successfully
- `403 Forbidden` - Only collection owner can archive
- `404 Not Found` - Collection not found

---

### 6. Restore Collection

**Endpoint:** `POST /collections/{collection_id}/restore`

Restores an archived collection to active state.

**Path Parameters:**
- `collection_id` (UUID, required): The collection ID

**Response:**
```json
{
  "success": true,
  "message": "Collection restored successfully"
}
```

**Status Codes:**
- `200 OK` - Collection restored successfully
- `403 Forbidden` - Only collection owner can restore
- `404 Not Found` - Collection not found

---

### 7. List User Collections

**Endpoint:** `GET /collections`

Retrieves all collections owned by the authenticated user.

**Response:**
```json
{
  "collections": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "owner_id": "550e8400-e29b-41d4-a716-446655440001",
      "encrypted_name": "base64_encrypted_collection_name",
      "collection_type": "folder",
      "parent_id": "550e8400-e29b-41d4-a716-446655440004",
      "ancestor_ids": [],
      "created_at": "2023-01-01T00:00:00Z",
      "modified_at": "2023-01-01T00:00:00Z",
      "members": []
    }
  ]
}
```

---

### 8. List Shared Collections

**Endpoint:** `GET /collections/shared`

Retrieves all collections shared with the authenticated user.

**Response:** Same format as List User Collections

---

### 9. Get Filtered Collections

**Endpoint:** `GET /collections/filtered`

Retrieves collections based on ownership and sharing filters.

**Query Parameters:**
- `include_owned` (boolean, optional): Include owned collections (default: true)
- `include_shared` (boolean, optional): Include shared collections (default: false)

**Example:** `GET /collections/filtered?include_owned=true&include_shared=true`

**Response:**
```json
{
  "owned_collections": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "owner_id": "550e8400-e29b-41d4-a716-446655440001",
      "encrypted_name": "base64_encrypted_collection_name",
      "collection_type": "folder",
      "created_at": "2023-01-01T00:00:00Z",
      "modified_at": "2023-01-01T00:00:00Z",
      "members": []
    }
  ],
  "shared_collections": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "owner_id": "550e8400-e29b-41d4-a716-446655440003",
      "encrypted_name": "shared_collection_name",
      "collection_type": "album",
      "created_at": "2023-01-01T00:00:00Z",
      "modified_at": "2023-01-01T00:00:00Z",
      "members": []
    }
  ],
  "total_count": 2
}
```

---

### 10. Find Root Collections

**Endpoint:** `GET /collections/root`

Retrieves all root-level collections (no parent) owned by the user.

**Response:** Same format as List User Collections

---

### 11. Find Collections by Parent

**Endpoint:** `GET /collections-by-parent/{parent_id}`

Retrieves all direct child collections of a parent collection.

**Path Parameters:**
- `parent_id` (UUID, required): The parent collection ID

**Response:** Same format as List User Collections

---

### 12. Share Collection

**Endpoint:** `POST /collections/{collection_id}/share`

Shares a collection with another user.

**Path Parameters:**
- `collection_id` (UUID, required): The collection ID

**Request Body:**
```json
{
  "collection_id": "550e8400-e29b-41d4-a716-446655440000",
  "recipient_id": "550e8400-e29b-41d4-a716-446655440001",
  "recipient_email": "user@example.com",
  "permission_level": "read_write",
  "encrypted_collection_key": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
  "share_with_descendants": true
}
```

**Request Fields:**
- `collection_id` (UUID, required): Collection ID (automatically set from URL)
- `recipient_id` (UUID, required): ID of user receiving access
- `recipient_email` (string, required): Email of recipient for display
- `permission_level` (string, required): "read_only", "read_write", or "admin"
- `encrypted_collection_key` (byte array, required): Collection key encrypted with recipient's public key
- `share_with_descendants` (boolean, required): Whether to share all child collections

**Response:**
```json
{
  "success": true,
  "message": "Collection shared successfully",
  "memberships_created": 5
}
```

**Status Codes:**
- `200 OK` - Collection shared successfully
- `403 Forbidden` - User doesn't have admin permission
- `404 Not Found` - Collection not found

---

### 13. Remove Member

**Endpoint:** `DELETE /collections/{collection_id}/members`

Removes a user's access to a collection.

**Path Parameters:**
- `collection_id` (UUID, required): The collection ID

**Request Body:**
```json
{
  "collection_id": "550e8400-e29b-41d4-a716-446655440000",
  "recipient_id": "550e8400-e29b-41d4-a716-446655440001",
  "remove_from_descendants": true
}
```

**Request Fields:**
- `collection_id` (UUID, required): Collection ID (automatically set from URL)
- `recipient_id` (UUID, required): ID of user to remove
- `remove_from_descendants` (boolean, required): Whether to remove from child collections

**Response:**
```json
{
  "success": true,
  "message": "Member removed successfully"
}
```

---

### 14. Move Collection

**Endpoint:** `POST /collections/{collection_id}/move`

Moves a collection to a new parent location.

**Path Parameters:**
- `collection_id` (UUID, required): The collection ID

**Request Body:**
```json
{
  "collection_id": "550e8400-e29b-41d4-a716-446655440000",
  "new_parent_id": "550e8400-e29b-41d4-a716-446655440001",
  "updated_ancestors": [
    "550e8400-e29b-41d4-a716-446655440002",
    "550e8400-e29b-41d4-a716-446655440001"
  ],
  "updated_path_segments": [
    "root_folder",
    "parent_folder"
  ]
}
```

**Request Fields:**
- `collection_id` (UUID, required): Collection ID (automatically set from URL)
- `new_parent_id` (UUID, required): New parent collection ID
- `updated_ancestors` (array, required): New ancestor chain
- `updated_path_segments` (array, required): Path segments for display

**Response:**
```json
{
  "success": true,
  "message": "Collection moved successfully"
}
```

---

### 15. Sync Collections

**Endpoint:** `GET /sync/collections`

Retrieves collection synchronization data for offline clients.

**Query Parameters:**
- `limit` (integer, optional): Maximum number of results (default: 1000, max: 5000)
- `cursor` (string, optional): Base64 encoded cursor for pagination

**Example:** `GET /sync/collections?limit=100&cursor=eyJsYXN0X21vZGlmaWVkIjoiMjAyMy0wMS0wMVQwMDowMDowMFoiLCJsYXN0X2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIn0=`

**Response:**
```json
{
  "collections": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "version": 1,
      "modified_at": "2023-01-01T00:00:00Z",
      "state": "active",
      "parent_id": "550e8400-e29b-41d4-a716-446655440001",
      "tombstone_version": 0,
      "tombstone_expiry": "0001-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "eyJsYXN0X21vZGlmaWVkIjoiMjAyMy0wMS0wMVQwMDowMDowMFoiLCJsYXN0X2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIn0=",
  "has_more": true
}
```

**Response Fields:**
- `collections` (array): Array of minimal collection sync data
  - `id` (UUID): Collection ID
  - `version` (integer): Current version number
  - `modified_at` (timestamp): Last modification time
  - `state` (string): Current state (active, deleted, archived)
  - `parent_id` (UUID, optional): Parent collection ID
  - `tombstone_version` (integer): Version when deleted (0 if not deleted)
  - `tombstone_expiry` (timestamp): When tombstone expires
- `next_cursor` (string, optional): Cursor for next page
- `has_more` (boolean): Whether more results are available

---

## Error Responses

All endpoints may return these standard error responses:

### 400 Bad Request
```json
{
  "error": {
    "code": 400,
    "message": "Bad Request",
    "details": {
      "field_name": "Field specific error message"
    }
  }
}
```

### 401 Unauthorized
```json
{
  "error": {
    "code": 401,
    "message": "Authentication required"
  }
}
```

### 403 Forbidden
```json
{
  "error": {
    "code": 403,
    "message": "You don't have permission to access this resource"
  }
}
```

### 404 Not Found
```json
{
  "error": {
    "code": 404,
    "message": "Resource not found"
  }
}
```

### 500 Internal Server Error
```json
{
  "error": {
    "code": 500,
    "message": "Internal server error"
  }
}
```

---

## Implementation Examples

### React Native Example

```javascript
class MapleFileAPI {
  constructor(baseURL, sessionToken) {
    this.baseURL = baseURL;
    this.sessionToken = sessionToken;
  }

  async createCollection(collectionData) {
    const response = await fetch(`${this.baseURL}/collections`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.sessionToken}`
      },
      body: JSON.stringify(collectionData)
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  }

  async getCollection(collectionId) {
    const response = await fetch(`${this.baseURL}/collections/${collectionId}`, {
      headers: {
        'Authorization': `Bearer ${this.sessionToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  }

  async shareCollection(collectionId, shareData) {
    const response = await fetch(`${this.baseURL}/collections/${collectionId}/share`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.sessionToken}`
      },
      body: JSON.stringify(shareData)
    });

    return await response.json();
  }
}
```

### React.js Example

```javascript
import axios from 'axios';

class MapleFileService {
  constructor(baseURL) {
    this.api = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    // Add session token to all requests
    this.api.interceptors.request.use(config => {
      const token = localStorage.getItem('session_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });
  }

  async syncCollections(cursor = null, limit = 1000) {
    const params = { limit };
    if (cursor) params.cursor = cursor;

    const response = await this.api.get('/sync/collections', { params });
    return response.data;
  }

  async getFilteredCollections(includeOwned = true, includeShared = false) {
    const response = await this.api.get('/collections/filtered', {
      params: {
        include_owned: includeOwned,
        include_shared: includeShared
      }
    });
    return response.data;
  }
}
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

type MapleFileClient struct {
    BaseURL     string
    SessionToken string
    HTTPClient  *http.Client
}

type CreateCollectionRequest struct {
    ID                     string                 `json:"id"`
    EncryptedName          string                 `json:"encrypted_name"`
    CollectionType         string                 `json:"collection_type"`
    EncryptedCollectionKey map[string]interface{} `json:"encrypted_collection_key"`
    ParentID               string                 `json:"parent_id,omitempty"`
    AncestorIDs            []string               `json:"ancestor_ids,omitempty"`
}

func (c *MapleFileClient) CreateCollection(req CreateCollectionRequest) (*CreateCollectionResponse, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    httpReq, err := http.NewRequest("POST", c.BaseURL+"/collections", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+c.SessionToken)

    resp, err := c.HTTPClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: %d", resp.StatusCode)
    }

    var result CreateCollectionResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}
```

---

## Notes for Developers

### Security Considerations

1. **End-to-End Encryption**: All collection names and keys are encrypted client-side. The server cannot decrypt this data.

2. **Access Control**: The API enforces strict permission checks. Users can only access collections they own or have been explicitly granted access to.

3. **Session Management**: Ensure proper session token handling and renewal.

### Performance Optimization

1. **Pagination**: Use the sync endpoints with cursors for large datasets.

2. **Caching**: Implement client-side caching for collection hierarchies to reduce API calls.

3. **Batch Operations**: Consider batching multiple collection operations when possible.

### Data Consistency

1. **Version Control**: Always include version numbers in update requests to prevent conflicts.

2. **State Management**: Handle collection states properly (active, deleted, archived).

3. **Hierarchical Integrity**: Maintain proper parent-child relationships when moving collections.

This API documentation provides complete coverage of all collection management operations with detailed examples for React Native, React.js, and Go implementations.
