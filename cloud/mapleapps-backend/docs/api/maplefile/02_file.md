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
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "collection_id": "550e8400-e29b-41d4-a716-446655440001",
      "version": 5,
      "modified_at": "2023-12-01T16:00:00Z",
      "state": "deleted",
      "tombstone_version": 5,
      "tombstone_expiry": "2024-01-01T16:00:00Z",
      "encrypted_file_size_in_bytes": 2097152
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "collection_id": "550e8400-e29b-41d4-a716-446655440004",
      "version": 2,
      "modified_at": "2023-12-01T14:20:00Z",
      "state": "archived",
      "tombstone_version": 0,
      "tombstone_expiry": "0001-01-01T00:00:00Z",
      "encrypted_file_size_in_bytes": 524288
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440005",
      "collection_id": "550e8400-e29b-41d4-a716-446655440006",
      "version": 1,
      "modified_at": "2023-12-01T13:10:00Z",
      "state": "pending",
      "tombstone_version": 0,
      "tombstone_expiry": "0001-01-01T00:00:00Z",
      "encrypted_file_size_in_bytes": 0
    }
  ],
  "next_cursor": "eyJsYXN0X21vZGlmaWVkIjoiMjAyMy0xMi0wMVQxMzoxMDowMFoiLCJsYXN0X2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDA1In0=",
  "has_more": true
}
```

### Response Fields

#### FileSyncResponse
| Field | Type | Description |
|-------|------|-------------|
| `files` | array[FileSyncItem] | Array of file sync items |
| `next_cursor` | string | Cursor for next page (null if no more) |
| `has_more` | boolean | Whether more results available |

#### FileSyncItem
| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | File ID |
| `collection_id` | UUID | Collection ID |
| `version` | number | Current version number for optimistic locking |
| `modified_at` | timestamp | Last modification time (ISO 8601 format) |
| `state` | string | Current file state (`pending`, `active`, `deleted`, `archived`) |
| `tombstone_version` | number | Version when file was deleted (0 if not deleted) |
| `tombstone_expiry` | timestamp | When tombstone expires (zero time if not deleted) |
| `encrypted_file_size_in_bytes` | number | Size of encrypted file content in bytes (0 for pending files) |

### File States and Size Information

| State | Description | Size Behavior |
|-------|-------------|---------------|
| `pending` | File metadata created but upload not completed | `encrypted_file_size_in_bytes` is 0 |
| `active` | File fully uploaded and available | `encrypted_file_size_in_bytes` shows actual encrypted file size |
| `deleted` | File soft-deleted (tombstoned) | `encrypted_file_size_in_bytes` retains last known size |
| `archived` | File archived but still accessible | `encrypted_file_size_in_bytes` shows actual encrypted file size |

### Usage Examples

#### JavaScript/TypeScript
```javascript
async function syncFiles(cursor = null, limit = 1000) {
  const url = new URL('/maplefile/api/v1/sync/files', baseURL);
  if (cursor) url.searchParams.set('cursor', cursor);
  if (limit) url.searchParams.set('limit', limit.toString());

  const response = await fetch(url, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) {
    throw new Error(`Sync failed: ${response.statusText}`);
  }

  return await response.json();
}

// Process files with size tracking
function processFileSyncData(syncResponse) {
  let totalActiveSize = 0;
  let totalDeletedSize = 0;

  syncResponse.files.forEach(file => {
    console.log(`File ${file.id}: version ${file.version}, state ${file.state}, size ${file.encrypted_file_size_in_bytes} bytes`);

    switch (file.state) {
      case 'active':
      case 'archived':
        totalActiveSize += file.encrypted_file_size_in_bytes;
        break;
      case 'deleted':
        totalDeletedSize += file.encrypted_file_size_in_bytes;
        // Remove from local storage but track size for cleanup
        localStorage.removeItem(`file_${file.id}`);
        break;
      case 'pending':
        console.log(`File ${file.id} is still pending upload`);
        break;
    }

    // Update local file metadata
    if (file.state !== 'deleted') {
      localStorage.setItem(`file_${file.id}`, JSON.stringify({
        version: file.version,
        state: file.state,
        modified_at: file.modified_at,
        size: file.encrypted_file_size_in_bytes
      }));
    }
  });

  console.log(`Total active storage: ${totalActiveSize} bytes`);
  console.log(`Total deleted storage: ${totalDeletedSize} bytes`);
}

// Usage
let syncResponse = await syncFiles();
processFileSyncData(syncResponse);

// Continue syncing
while (syncResponse.has_more) {
  syncResponse = await syncFiles(syncResponse.next_cursor);
  processFileSyncData(syncResponse);
}
```

#### Go Example
```go
type FileSyncClient struct {
    baseURL string
    token   string
    client  *http.Client
}

type FileSyncResponse struct {
    Files      []FileSyncItem  `json:"files"`
    NextCursor *string         `json:"next_cursor"`
    HasMore    bool            `json:"has_more"`
}

type FileSyncItem struct {
    ID                       string    `json:"id"`
    CollectionID             string    `json:"collection_id"`
    Version                  uint64    `json:"version"`
    ModifiedAt               time.Time `json:"modified_at"`
    State                    string    `json:"state"`
    TombstoneVersion         uint64    `json:"tombstone_version"`
    TombstoneExpiry          time.Time `json:"tombstone_expiry"`
    EncryptedFileSizeInBytes int64     `json:"encrypted_file_size_in_bytes"`
}

func (c *FileSyncClient) SyncFiles(cursor *string, limit int) (*FileSyncResponse, error) {
    url := fmt.Sprintf("%s/maplefile/api/v1/sync/files", c.baseURL)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    q := req.URL.Query()
    if cursor != nil {
        q.Add("cursor", *cursor)
    }
    if limit > 0 {
        q.Add("limit", strconv.Itoa(limit))
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
        return nil, fmt.Errorf("sync failed with status %d", resp.StatusCode)
    }

    var syncResp FileSyncResponse
    if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
        return nil, err
    }

    return &syncResp, nil
}

// Usage with size tracking
func (c *FileSyncClient) PerformFullSyncWithSizeTracking() error {
    var cursor *string
    var totalActiveSize int64
    var totalDeletedSize int64

    for {
        resp, err := c.SyncFiles(cursor, 1000)
        if err != nil {
            return err
        }

        // Process files with size tracking
        for _, file := range resp.Files {
            fmt.Printf("File %s: version %d, state %s, size %d bytes\n",
                file.ID, file.Version, file.State, file.EncryptedFileSizeInBytes)

            switch file.State {
            case "active", "archived":
                totalActiveSize += file.EncryptedFileSizeInBytes
                c.updateLocalFile(file)
            case "deleted":
                totalDeletedSize += file.EncryptedFileSizeInBytes
                c.handleFileDeleted(file.ID)
            case "pending":
                fmt.Printf("File %s is still pending upload (size will be 0)\n", file.ID)
                c.updateLocalFile(file)
            }
        }

        if !resp.HasMore {
            break
        }
        cursor = resp.NextCursor
    }

    fmt.Printf("Sync complete - Active storage: %d bytes, Deleted storage: %d bytes\n",
               totalActiveSize, totalDeletedSize)
    return nil
}

func (c *FileSyncClient) updateLocalFile(file FileSyncItem) {
    // Update local file state with size information
    localFile := LocalFileMetadata{
        Version:   file.Version,
        State:     file.State,
        ModifiedAt: file.ModifiedAt,
        Size:      file.EncryptedFileSizeInBytes,
    }
    // Store in local database...
}

func (c *FileSyncClient) handleFileDeleted(fileID string) {
    // Remove from local storage but potentially track for billing
    fmt.Printf("Removing deleted file %s from local storage\n", fileID)
    // Remove from local database...
}
```

### Storage and Billing Considerations

The `encrypted_file_size_in_bytes` field provides important information for:

1. **Storage Accounting**: Track total storage usage across all user files
2. **Billing Calculations**: Accurate billing based on actual encrypted storage consumption
3. **Quota Management**: Enforce storage limits per user or organization
4. **Bandwidth Estimation**: Estimate transfer costs for sync operations
5. **Cleanup Planning**: Identify large files for deletion or archiving

### Performance Notes

1. **Size-Based Sync**: Use file sizes to prioritize sync order (smaller files first)
2. **Bandwidth Management**: Throttle sync based on cumulative file sizes
3. **Storage Optimization**: Identify files that may benefit from re-encryption with better compression
4. **Cache Management**: Use file sizes for local cache eviction policies

### Security Notes

- `encrypted_file_size_in_bytes` represents the size of the encrypted content, not the original file size
- This field is not sensitive information and can be used for storage management
- The actual file content remains encrypted and inaccessible without proper decryption keys
- Size information helps with storage quotas without revealing file content details
