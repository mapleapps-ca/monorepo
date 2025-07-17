# MapleFile File API Documentation

## Overview

The MapleFile API provides comprehensive file management capabilities with end-to-end encryption, collection-based organization, and presigned URL support for direct cloud storage operations. This document covers all file-related endpoints and their usage.

## Base Information

- **Base URL**: `/maplefile/api/v1`
- **Authentication**: Required for all endpoints (Bearer token or session-based)
- **Content-Type**: `application/json`
- **Response Format**: JSON

## Authentication

All endpoints require authentication. The authenticated user's ID is extracted from the session context and used for permission checks and ownership validation.

## Common Data Types

### UUID Format
All IDs use UUID format: `550e8400-e29b-41d4-a716-446655440000`

### Timestamps
All timestamps are in RFC3339 format: `2023-12-01T15:30:00Z`

### File States
- `pending`: File metadata created but upload not completed
- `active`: File fully uploaded and available
- `deleted`: File soft-deleted (tombstoned)
- `archived`: File archived but still accessible

### EncryptedFileKey Structure
```json
{
  "ciphertext": "base64-encoded-encrypted-key",
  "nonce": "base64-encoded-nonce",
  "key_version": 1,
  "rotated_at": "2023-12-01T15:30:00Z",
  "previous_keys": []
}
```

---

## API Endpoints

## 1. Create Pending File

Creates a new file record in "pending" state and returns presigned URLs for upload.

### Request
- **Method**: `POST`
- **Path**: `/maplefile/api/v1/files/pending`
- **Content-Type**: `application/json`

### Request Body
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "collection_id": "550e8400-e29b-41d4-a716-446655440001",
  "encrypted_metadata": "base64-encoded-encrypted-metadata",
  "encrypted_file_key": {
    "ciphertext": "base64-encoded-encrypted-key",
    "nonce": "base64-encoded-nonce",
    "key_version": 1
  },
  "encryption_version": "v1.0",
  "encrypted_hash": "base64-encoded-encrypted-hash",
  "expected_file_size_in_bytes": 1048576,
  "expected_thumbnail_size_in_bytes": 8192
}
```

### Response
```json
{
  "file": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "collection_id": "550e8400-e29b-41d4-a716-446655440001",
    "owner_id": "550e8400-e29b-41d4-a716-446655440002",
    "encrypted_metadata": "base64-encoded-encrypted-metadata",
    "encrypted_file_key": {
      "ciphertext": "base64-encoded-encrypted-key",
      "nonce": "base64-encoded-nonce",
      "key_version": 1
    },
    "encryption_version": "v1.0",
    "encrypted_hash": "base64-encoded-encrypted-hash",
    "encrypted_file_size_in_bytes": 1048576,
    "encrypted_thumbnail_size_in_bytes": 8192,
    "created_at": "2023-12-01T15:30:00Z",
    "modified_at": "2023-12-01T15:30:00Z",
    "version": 1,
    "state": "pending",
    "tombstone_version": 0,
    "tombstone_expiry": "0001-01-01T00:00:00Z"
  },
  "presigned_upload_url": "https://s3.amazonaws.com/bucket/path?signed-params",
  "presigned_thumbnail_url": "https://s3.amazonaws.com/bucket/path_thumb?signed-params",
  "upload_url_expiration_time": "2023-12-01T16:30:00Z",
  "success": true,
  "message": "Pending file created successfully. Use the presigned URL to upload your file."
}
```

---

## 12. List File Sync Data

Returns file sync data with cursor-based pagination for synchronization purposes. This endpoint allows clients to efficiently synchronize their local file state with the server.

### Request
- **Method**: `GET`
- **Path**: `/maplefile/api/v1/sync/files`

### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cursor` | string | No | Base64-encoded cursor for pagination |
| `limit` | number | No | Maximum items per page (default: 5000, max: 10000) |

### Example Request
```
GET /maplefile/api/v1/sync/files?limit=1000&cursor=eyJsYXN0X21vZGlmaWVkIjoiMjAyMy0xMi0wMVQxNTozMDowMFoiLCJsYXN0X2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIn0=
```

### Response
```json
{
  "files": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "collection_id": "550e8400-e29b-41d4-a716-446655440001",
      "version": 3,
      "modified_at": "2023-12-01T15:35:00Z",
      "state": "active",
      "tombstone_version": 0,
      "tombstone_expiry": "0001-01-01T00:00:00Z",
      "encrypted_file_size_in_bytes": 1048576
    }
  ],
  "next_cursor": "eyJsYXN0X21vZGlmaWVkIjoiMjAyMy0xMi0wMVQxMzoxMDowMFoiLCJsYXN0X2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDA1In0=",
  "has_more": true
}
```

---

## 13. List Recent Files

Returns the most recently modified files with cursor-based pagination. This endpoint is optimized for displaying recent file activity and only returns active files from collections the user has access to.

### Request
- **Method**: `GET`
- **Path**: `/maplefile/api/v1/files/recent`

### Query Parameters
| Parameter | Type | Required | Default | Max | Description |
|-----------|------|----------|---------|-----|-------------|
| `limit` | integer | No | 30 | 100 | Maximum number of files to return per page |
| `cursor` | string | No | - | - | Base64-encoded cursor for pagination |

### Example Request
```
GET /maplefile/api/v1/files/recent?limit=30&cursor=eyJsYXN0X21vZGlmaWVkIjoiMjAyMy0xMi0wMVQxNjo0NTowMFoiLCJsYXN0X2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIn0=
```

### Response
```json
{
  "files": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "collection_id": "550e8400-e29b-41d4-a716-446655440001",
      "owner_id": "550e8400-e29b-41d4-a716-446655440002",
      "encrypted_metadata": "base64-encoded-encrypted-metadata",
      "encrypted_file_key": {
        "ciphertext": "base64-encoded-encrypted-key",
        "nonce": "base64-encoded-nonce",
        "key_version": 1
      },
      "encryption_version": "v1.0",
      "encrypted_hash": "base64-encoded-encrypted-hash",
      "encrypted_file_size_in_bytes": 1048576,
      "encrypted_thumbnail_size_in_bytes": 8192,
      "created_at": "2023-12-01T15:30:00Z",
      "modified_at": "2023-12-01T16:45:00Z",
      "version": 3,
      "state": "active"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "collection_id": "550e8400-e29b-41d4-a716-446655440004",
      "owner_id": "550e8400-e29b-41d4-a716-446655440002",
      "encrypted_metadata": "base64-encoded-encrypted-metadata-2",
      "encrypted_file_key": {
        "ciphertext": "base64-encoded-encrypted-key-2",
        "nonce": "base64-encoded-nonce-2",
        "key_version": 1
      },
      "encryption_version": "v1.0",
      "encrypted_hash": "base64-encoded-encrypted-hash-2",
      "encrypted_file_size_in_bytes": 2097152,
      "encrypted_thumbnail_size_in_bytes": 16384,
      "created_at": "2023-12-01T14:20:00Z",
      "modified_at": "2023-12-01T16:30:00Z",
      "version": 2,
      "state": "active"
    }
  ],
  "next_cursor": "eyJsYXN0X21vZGlmaWVkIjoiMjAyMy0xMi0wMVQxNjozMDowMFoiLCJsYXN0X2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAzIn0=",
  "has_more": true,
  "total_count": 30
}
```

### Response Fields

#### ListRecentFilesResponse
| Field | Type | Description |
|-------|------|-------------|
| `files` | array[RecentFileItem] | Array of recent file objects |
| `next_cursor` | string | Cursor for next page (null if no more) |
| `has_more` | boolean | Whether more results available |
| `total_count` | integer | Number of files in current response |

#### RecentFileItem
| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | File ID |
| `collection_id` | UUID | Collection ID |
| `owner_id` | UUID | File owner's user ID |
| `encrypted_metadata` | string | Base64-encoded encrypted file metadata |
| `encrypted_file_key` | object | Encrypted file encryption key |
| `encryption_version` | string | Version of encryption scheme used |
| `encrypted_hash` | string | Hash of encrypted file content |
| `encrypted_file_size_in_bytes` | integer | Size of encrypted file in bytes |
| `encrypted_thumbnail_size_in_bytes` | integer | Size of encrypted thumbnail in bytes |
| `created_at` | timestamp | File creation time (ISO 8601 format) |
| `modified_at` | timestamp | Last modification time (ISO 8601 format) |
| `version` | integer | Current version number for optimistic locking |
| `state` | string | File state (always "active" for this endpoint) |

### Filtering and Ordering

| Aspect | Behavior |
|--------|----------|
| **Ordering** | Files ordered by `modified_at` DESC, then by `id` ASC |
| **State Filter** | Only `active` files are returned |
| **Access Control** | Only files from collections user has access to |
| **Collection Types** | Includes both owned and shared collections |

### Usage Examples

#### JavaScript/TypeScript
```javascript
async function getRecentFiles(limit = 30, cursor = null) {
  const url = new URL('/maplefile/api/v1/files/recent', baseURL);
  if (limit) url.searchParams.set('limit', limit.toString());
  if (cursor) url.searchParams.set('cursor', cursor);

  const response = await fetch(url, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) {
    throw new Error(`Failed to get recent files: ${response.statusText}`);
  }

  return await response.json();
}

// Usage - Get first page
let recentFiles = await getRecentFiles(30);
console.log(`Found ${recentFiles.files.length} recent files`);

// Process each file
recentFiles.files.forEach(file => {
  console.log(`File: ${file.id}, Modified: ${file.modified_at}, Size: ${file.encrypted_file_size_in_bytes} bytes`);
});

// Get next page if available
if (recentFiles.has_more) {
  const nextPage = await getRecentFiles(30, recentFiles.next_cursor);
  console.log(`Next page has ${nextPage.files.length} files`);
}

// Get all recent files (paginate through all)
async function getAllRecentFiles() {
  let allFiles = [];
  let cursor = null;

  do {
    const response = await getRecentFiles(50, cursor);
    allFiles.push(...response.files);
    cursor = response.next_cursor;
  } while (cursor);

  return allFiles;
}
```

#### Go Example
```go
type RecentFilesClient struct {
    baseURL string
    token   string
    client  *http.Client
}

type RecentFilesResponse struct {
    Files      []RecentFileItem `json:"files"`
    NextCursor *string          `json:"next_cursor"`
    HasMore    bool             `json:"has_more"`
    TotalCount int              `json:"total_count"`
}

type RecentFileItem struct {
    ID                       string    `json:"id"`
    CollectionID             string    `json:"collection_id"`
    OwnerID                  string    `json:"owner_id"`
    EncryptedMetadata        string    `json:"encrypted_metadata"`
    EncryptedFileKey         FileKey   `json:"encrypted_file_key"`
    EncryptionVersion        string    `json:"encryption_version"`
    EncryptedHash            string    `json:"encrypted_hash"`
    EncryptedFileSizeInBytes int64     `json:"encrypted_file_size_in_bytes"`
    CreatedAt                time.Time `json:"created_at"`
    ModifiedAt               time.Time `json:"modified_at"`
    Version                  uint64    `json:"version"`
    State                    string    `json:"state"`
}

func (c *RecentFilesClient) GetRecentFiles(limit int, cursor *string) (*RecentFilesResponse, error) {
    url := fmt.Sprintf("%s/maplefile/api/v1/files/recent", c.baseURL)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    q := req.URL.Query()
    if limit > 0 {
        q.Add("limit", strconv.Itoa(limit))
    }
    if cursor != nil {
        q.Add("cursor", *cursor)
    }
    req.URL.RawQuery = q.Encode()

    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
    }

    var result RecentFilesResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

// Usage example
func (c *RecentFilesClient) ListAllRecentFiles() ([]RecentFileItem, error) {
    var allFiles []RecentFileItem
    var cursor *string

    for {
        resp, err := c.GetRecentFiles(50, cursor)
        if err != nil {
            return nil, err
        }

        allFiles = append(allFiles, resp.Files...)

        if !resp.HasMore {
            break
        }
        cursor = resp.NextCursor
    }

    return allFiles, nil
}
```

#### Python Example
```python
import requests
import json
from typing import Optional, List, Dict, Any

class RecentFilesClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.token = token
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        })

    def get_recent_files(self, limit: int = 30, cursor: Optional[str] = None) -> Dict[str, Any]:
        url = f"{self.base_url}/maplefile/api/v1/files/recent"
        params = {'limit': limit}
        if cursor:
            params['cursor'] = cursor

        response = self.session.get(url, params=params)
        response.raise_for_status()
        return response.json()

    def get_all_recent_files(self) -> List[Dict[str, Any]]:
        all_files = []
        cursor = None

        while True:
            response = self.get_recent_files(limit=50, cursor=cursor)
            all_files.extend(response['files'])

            if not response['has_more']:
                break
            cursor = response['next_cursor']

        return all_files

# Usage
client = RecentFilesClient('https://api.example.com', 'your-token')
recent_files = client.get_recent_files(30)
print(f"Found {len(recent_files['files'])} recent files")

for file in recent_files['files']:
    print(f"File: {file['id']}, Modified: {file['modified_at']}")
```

### Performance Notes

1. **Optimized Querying**: Uses Cassandra table optimized for user-based file access
2. **Efficient Pagination**: Cursor-based pagination ensures consistent results
3. **Access Control**: Pre-filters accessible collections to minimize database queries
4. **Reasonable Limits**: Default limit of 30, maximum of 100 to prevent large result sets
5. **Index Usage**: Leverages existing database indices for optimal performance

### Security Notes

- Only returns files from collections the authenticated user has access to
- Filters out non-active files (deleted, archived, pending) for security
- All file content remains encrypted; only metadata is accessible
- Respects collection-level permissions (owned and shared collections)
- No sensitive file content is exposed through this endpoint

### Error Handling

#### 400 Bad Request
```json
{
  "error": {
    "limit": "Limit cannot exceed 100"
  }
}
```

#### 401 Unauthorized
```json
{
  "error": {
    "message": "Authentication required"
  }
}
```

#### 404 Not Found
```json
{
  "error": {
    "message": "Recent files not found"
  }
}
```

#### 500 Internal Server Error
```json
{
  "error": {
    "message": "Internal server error"
  }
}
```

### Use Cases

1. **File Activity Dashboard**: Display recently modified files in a user interface
2. **Quick File Access**: Allow users to quickly access files they've been working on
3. **Sync Optimization**: Help clients identify recently changed files for synchronization
4. **Activity Monitoring**: Track file modification patterns for analytics
5. **Mobile Apps**: Efficiently load recent files on mobile devices with limited bandwidth
