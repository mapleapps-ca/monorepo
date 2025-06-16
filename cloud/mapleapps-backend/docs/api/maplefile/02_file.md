# MapleFile API Documentation

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

### Request Fields
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Client-generated unique file ID |
| `collection_id` | UUID | Yes | Collection where file belongs |
| `encrypted_metadata` | string | Yes | Base64-encoded encrypted file metadata (filename, MIME type, etc.) |
| `encrypted_file_key` | object | Yes | Encrypted file encryption key |
| `encryption_version` | string | Yes | Version of encryption scheme used |
| `encrypted_hash` | string | Yes | Hash of encrypted file content for integrity |
| `expected_file_size_in_bytes` | number | No | Expected file size in bytes |
| `expected_thumbnail_size_in_bytes` | number | No | Expected thumbnail size in bytes |

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
    "encrypted_file_size_in_bytes": 0,
    "encrypted_thumbnail_size_in_bytes": 0,
    "created_at": "2023-12-01T15:30:00Z",
    "modified_at": "2023-12-01T15:30:00Z"
  },
  "presigned_upload_url": "https://s3.amazonaws.com/bucket/path?signed-params",
  "presigned_thumbnail_url": "https://s3.amazonaws.com/bucket/path_thumb?signed-params",
  "upload_url_expiration_time": "2023-12-01T16:30:00Z",
  "success": true,
  "message": "Pending file created successfully. Use the presigned URL to upload your file."
}
```

### Error Responses
- **400 Bad Request**: Invalid request data
- **403 Forbidden**: No write access to collection
- **409 Conflict**: File ID already exists
- **500 Internal Server Error**: Server error

---

## 2. Complete File Upload

Transitions a file from "pending" to "active" state after verifying upload completion.

### Request
- **Method**: `POST`
- **Path**: `/maplefile/api/v1/files/{file_id}/complete`
- **Content-Type**: `application/json`

### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `file_id` | UUID | The file ID to complete |

### Request Body
```json
{
  "actual_file_size_in_bytes": 1048576,
  "actual_thumbnail_size_in_bytes": 8192,
  "upload_confirmed": true,
  "thumbnail_upload_confirmed": true
}
```

### Request Fields
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `actual_file_size_in_bytes` | number | No | Actual uploaded file size for validation |
| `actual_thumbnail_size_in_bytes` | number | No | Actual uploaded thumbnail size |
| `upload_confirmed` | boolean | No | Client confirmation of successful upload |
| `thumbnail_upload_confirmed` | boolean | No | Client confirmation of thumbnail upload |

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
    "modified_at": "2023-12-01T15:35:00Z"
  },
  "success": true,
  "message": "File upload completed successfully",
  "actual_file_size": 1048576,
  "actual_thumbnail_size": 8192,
  "upload_verified": true,
  "thumbnail_verified": true
}
```

### Error Responses
- **400 Bad Request**: File not in pending state or upload verification failed
- **403 Forbidden**: No write access to collection
- **404 Not Found**: File not found
- **500 Internal Server Error**: Server error

---

## 3. Get File

Retrieves file metadata by ID.

### Request
- **Method**: `GET`
- **Path**: `/maplefile/api/v1/files/{file_id}`

### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `file_id` | UUID | The file ID to retrieve |

### Response
```json
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
  "modified_at": "2023-12-01T15:35:00Z"
}
```

### Error Responses
- **400 Bad Request**: Invalid file ID format
- **403 Forbidden**: No read access to collection
- **404 Not Found**: File not found
- **500 Internal Server Error**: Server error

---

## 4. Update File

Updates file metadata.

### Request
- **Method**: `PUT`
- **Path**: `/maplefile/api/v1/files/{file_id}`
- **Content-Type**: `application/json`

### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `file_id` | UUID | The file ID to update |

### Request Body
```json
{
  "encrypted_metadata": "new-base64-encoded-encrypted-metadata",
  "encrypted_file_key": {
    "ciphertext": "new-base64-encoded-encrypted-key",
    "nonce": "new-base64-encoded-nonce",
    "key_version": 2
  },
  "encryption_version": "v1.1",
  "encrypted_hash": "new-base64-encoded-encrypted-hash",
  "version": 5
}
```

### Request Fields
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `encrypted_metadata` | string | No | Updated encrypted metadata |
| `encrypted_file_key` | object | No | Updated encrypted file key |
| `encryption_version` | string | No | Updated encryption version |
| `encrypted_hash` | string | No | Updated encrypted hash |
| `version` | number | Yes | Current version for optimistic locking |

### Response
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "collection_id": "550e8400-e29b-41d4-a716-446655440001",
  "owner_id": "550e8400-e29b-41d4-a716-446655440002",
  "encrypted_metadata": "new-base64-encoded-encrypted-metadata",
  "encrypted_file_key": {
    "ciphertext": "new-base64-encoded-encrypted-key",
    "nonce": "new-base64-encoded-nonce",
    "key_version": 2
  },
  "encryption_version": "v1.1",
  "encrypted_hash": "new-base64-encoded-encrypted-hash",
  "encrypted_file_size_in_bytes": 1048576,
  "encrypted_thumbnail_size_in_bytes": 8192,
  "created_at": "2023-12-01T15:30:00Z",
  "modified_at": "2023-12-01T16:00:00Z"
}
```

### Error Responses
- **400 Bad Request**: Invalid data or version mismatch
- **403 Forbidden**: No write access to collection
- **404 Not Found**: File not found
- **500 Internal Server Error**: Server error

---

## 5. Archive File

Archives a file by changing its state to "archived".

### Request
- **Method**: `POST`
- **Path**: `/maplefile/api/v1/files/{file_id}/archive`

### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `file_id` | UUID | The file ID to archive |

### Response
```json
{
  "success": true,
  "message": "File archived successfully"
}
```

### Error Responses
- **400 Bad Request**: Invalid state transition
- **403 Forbidden**: No write access to collection
- **404 Not Found**: File not found
- **500 Internal Server Error**: Server error

---

## 6. Restore File

Restores an archived file back to "active" state.

### Request
- **Method**: `POST`
- **Path**: `/maplefile/api/v1/files/{file_id}/restore`

### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `file_id` | UUID | The file ID to restore |

### Response
```json
{
  "success": true,
  "message": "File restored successfully"
}
```

### Error Responses
- **400 Bad Request**: Invalid state transition
- **403 Forbidden**: No write access to collection
- **404 Not Found**: File not found
- **500 Internal Server Error**: Server error

---

## 7. Soft Delete File

Soft deletes a file by marking it as deleted with a tombstone.

### Request
- **Method**: `DELETE`
- **Path**: `/maplefile/api/v1/files/{file_id}`

### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `file_id` | UUID | The file ID to delete |

### Response
```json
{
  "success": true,
  "message": "File soft-deleted successfully"
}
```

### Error Responses
- **400 Bad Request**: Invalid state transition
- **403 Forbidden**: No write access to collection
- **404 Not Found**: File not found
- **500 Internal Server Error**: Server error

---

## 8. Delete Multiple Files

Soft deletes multiple files in a single operation.

### Request
- **Method**: `DELETE`
- **Path**: `/maplefile/api/v1/files/multiple`
- **Content-Type**: `application/json`

### Request Body
```json
{
  "file_ids": [
    "550e8400-e29b-41d4-a716-446655440000",
    "550e8400-e29b-41d4-a716-446655440001",
    "550e8400-e29b-41d4-a716-446655440002"
  ]
}
```

### Request Fields
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file_ids` | array[UUID] | Yes | Array of file IDs to delete |

### Response
```json
{
  "success": true,
  "message": "Successfully deleted 2 files",
  "deleted_count": 2,
  "skipped_count": 1,
  "total_requested": 3
}
```

### Error Responses
- **400 Bad Request**: Invalid file IDs
- **500 Internal Server Error**: Server error

---

## 9. Get Presigned Upload URL

Generates presigned URLs for uploading file data directly to cloud storage.

### Request
- **Method**: `POST`
- **Path**: `/maplefile/api/v1/files/{file_id}/upload-url`
- **Content-Type**: `application/json`

### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `file_id` | UUID | The file ID to generate upload URL for |

### Request Body
```json
{
  "url_duration": "3600000000000"
}
```

### Request Fields
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url_duration` | string | No | Duration in nanoseconds (default: 1 hour) |

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
    "modified_at": "2023-12-01T15:35:00Z"
  },
  "presigned_upload_url": "https://s3.amazonaws.com/bucket/path?signed-params",
  "presigned_thumbnail_url": "https://s3.amazonaws.com/bucket/path_thumb?signed-params",
  "upload_url_expiration_time": "2023-12-01T16:35:00Z",
  "success": true,
  "message": "Presigned upload URLs generated successfully"
}
```

### Error Responses
- **400 Bad Request**: Invalid duration
- **403 Forbidden**: No write access to collection
- **404 Not Found**: File not found
- **500 Internal Server Error**: Server error

---

## 10. Get Presigned Download URL

Generates presigned URLs for downloading file data directly from cloud storage.

### Request
- **Method**: `POST`
- **Path**: `/maplefile/api/v1/files/{file_id}/download-url`
- **Content-Type**: `application/json`

### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `file_id` | UUID | The file ID to generate download URL for |

### Request Body
```json
{
  "url_duration": "3600000000000"
}
```

### Request Fields
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url_duration` | string | No | Duration in nanoseconds (default: 1 hour) |

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
    "modified_at": "2023-12-01T15:35:00Z"
  },
  "presigned_download_url": "https://s3.amazonaws.com/bucket/path?signed-params",
  "presigned_thumbnail_url": "https://s3.amazonaws.com/bucket/path_thumb?signed-params",
  "download_url_expiration_time": "2023-12-01T16:35:00Z",
  "success": true,
  "message": "Presigned download URLs generated successfully"
}
```

### Error Responses
- **400 Bad Request**: Invalid duration
- **403 Forbidden**: No read access to collection
- **404 Not Found**: File not found
- **500 Internal Server Error**: Server error

---

## 11. List Files by Collection

Lists all files in a specific collection.

### Request
- **Method**: `GET`
- **Path**: `/maplefile/api/v1/collections/{collection_id}/files`

### Path Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `collection_id` | UUID | The collection ID to list files from |

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
      "modified_at": "2023-12-01T15:35:00Z"
    }
  ]
}
```

### Error Responses
- **403 Forbidden**: No read access to collection
- **404 Not Found**: Collection not found
- **500 Internal Server Error**: Server error

---

## 12. List File Sync Data

Returns file sync data with cursor-based pagination for synchronization purposes.

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
      "tombstone_expiry": "0001-01-01T00:00:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "collection_id": "550e8400-e29b-41d4-a716-446655440001",
      "version": 5,
      "modified_at": "2023-12-01T16:00:00Z",
      "state": "deleted",
      "tombstone_version": 5,
      "tombstone_expiry": "2024-01-01T16:00:00Z"
    }
  ],
  "next_cursor": "eyJsYXN0X21vZGlmaWVkIjoiMjAyMy0xMi0wMVQxNjowMDowMFoiLCJsYXN0X2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAyIn0=",
  "has_more": true
}
```

### Response Fields
| Field | Type | Description |
|-------|------|-------------|
| `files` | array | Array of file sync items |
| `files[].id` | UUID | File ID |
| `files[].collection_id` | UUID | Collection ID |
| `files[].version` | number | Current version number |
| `files[].modified_at` | timestamp | Last modification time |
| `files[].state` | string | Current file state |
| `files[].tombstone_version` | number | Version when deleted (0 if not deleted) |
| `files[].tombstone_expiry` | timestamp | When tombstone expires |
| `next_cursor` | string | Cursor for next page (null if no more) |
| `has_more` | boolean | Whether more results available |

### Error Responses
- **400 Bad Request**: Invalid cursor format
- **500 Internal Server Error**: Server error

---

## File Upload Workflow

### Complete File Upload Process

1. **Create Pending File**
   ```
   POST /maplefile/api/v1/files/pending
   ```
   - Creates file metadata in "pending" state
   - Returns presigned upload URLs

2. **Upload File Data**
   ```
   PUT {presigned_upload_url}
   Content-Type: application/octet-stream
   {encrypted_file_data}
   ```
   - Upload encrypted file data directly to S3
   - Upload thumbnail if applicable

3. **Complete Upload**
   ```
   POST /maplefile/api/v1/files/{file_id}/complete
   ```
   - Verifies upload completion
   - Transitions file to "active" state

### File Download Process

1. **Get Presigned Download URL**
   ```
   POST /maplefile/api/v1/files/{file_id}/download-url
   ```
   - Returns presigned download URLs

2. **Download File Data**
   ```
   GET {presigned_download_url}
   ```
   - Download encrypted file data directly from S3

---

## Error Handling

### Standard Error Response Format
```json
{
  "error": {
    "message": "Error description",
    "field_errors": {
      "field_name": "Field-specific error message"
    }
  }
}
```

### HTTP Status Codes
- **200 OK**: Success
- **400 Bad Request**: Invalid request data
- **401 Unauthorized**: Authentication required
- **403 Forbidden**: Permission denied
- **404 Not Found**: Resource not found
- **409 Conflict**: Resource conflict
- **500 Internal Server Error**: Server error

---

## Security Considerations

### End-to-End Encryption
- All file content is encrypted client-side before upload
- Server only stores encrypted data and metadata
- Encryption keys are encrypted with user's master key

### Access Control
- Files belong to collections with permission-based access
- Users must have appropriate permissions to read/write files
- Owner can always access their files

### Presigned URLs
- Limited time validity (default 1 hour, max 24 hours)
- Direct upload/download to/from S3 without server mediation
- URLs expire automatically for security

### File States and Tombstones
- Soft deletion preserves data with tombstone expiry
- State transitions are validated server-side
- Version tracking prevents concurrent modification conflicts

---

## Rate Limiting

The API implements standard rate limiting. Clients should implement exponential backoff for 429 responses and respect rate limit headers in responses.

---

## SDKs and Examples

### React Native Example
```javascript
// Create pending file
const createFile = async (fileData) => {
  const response = await fetch('/maplefile/api/v1/files/pending', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authToken}`
    },
    body: JSON.stringify({
      id: generateUUID(),
      collection_id: collectionId,
      encrypted_metadata: encryptedMetadata,
      encrypted_file_key: encryptedKey,
      encryption_version: 'v1.0',
      encrypted_hash: encryptedHash,
      expected_file_size_in_bytes: fileSize
    })
  });

  return response.json();
};

// Upload file using presigned URL
const uploadFile = async (presignedUrl, fileData) => {
  const response = await fetch(presignedUrl, {
    method: 'PUT',
    body: fileData,
    headers: {
      'Content-Type': 'application/octet-stream'
    }
  });

  return response.ok;
};

// Complete upload
const completeUpload = async (fileId) => {
  const response = await fetch(`/maplefile/api/v1/files/${fileId}/complete`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authToken}`
    },
    body: JSON.stringify({
      upload_confirmed: true
    })
  });

  return response.json();
};
```

### Go Example
```go
type FileClient struct {
    baseURL string
    token   string
    client  *http.Client
}

func (c *FileClient) CreatePendingFile(req CreatePendingFileRequest) (*CreatePendingFileResponse, error) {
    body, _ := json.Marshal(req)

    httpReq, _ := http.NewRequest("POST", c.baseURL+"/files/pending", bytes.NewBuffer(body))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+c.token)

    resp, err := c.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result CreatePendingFileResponse
    json.NewDecoder(resp.Body).Decode(&result)

    return &result, nil
}

func (c *FileClient) UploadFile(presignedURL string, data []byte) error {
    req, _ := http.NewRequest("PUT", presignedURL, bytes.NewBuffer(data))
    req.Header.Set("Content-Type", "application/octet-stream")

    resp, err := c.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return fmt.Errorf("upload failed with status %d", resp.StatusCode)
    }

    return nil
}
```
