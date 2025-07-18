# Dashboard API Endpoint

## Overview
The Dashboard API endpoint provides aggregated data for the MapleFile dashboard, including file counts, storage usage, storage trends, and recent files.

## Endpoint
```
GET /maplefile/api/v1/dashboard
```

## Authentication
Requires valid JWT token in the Authorization header:
```
Authorization: JWT <token>
```

## Response Structure

### Success Response (200 OK)
```json
{
  "dashboard": {
    "summary": {
      "total_files": 156,
      "total_folders": 23,
      "storage_used": {
        "value": 4.2,
        "unit": "GB"
      },
      "storage_limit": {
        "value": 15,
        "unit": "GB"
      },
      "storage_usage_percentage": 28
    },
    "storage_usage_trend": {
      "period": "Last 7 days",
      "data_points": [
        {
          "date": "2025-01-10",
          "usage": {
            "value": 3.8,
            "unit": "GB"
          }
        },
        {
          "date": "2025-01-11",
          "usage": {
            "value": 3.9,
            "unit": "GB"
          }
        },
        {
          "date": "2025-01-12",
          "usage": {
            "value": 4.0,
            "unit": "GB"
          }
        },
        {
          "date": "2025-01-13",
          "usage": {
            "value": 4.1,
            "unit": "GB"
          }
        },
        {
          "date": "2025-01-14",
          "usage": {
            "value": 4.2,
            "unit": "GB"
          }
        }
      ]
    },
    "recent_files": [
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
        "encrypted_file_size_in_bytes": 2457600,
        "encrypted_thumbnail_size_in_bytes": 8192,
        "created_at": "2025-01-16T12:30:00Z",
        "modified_at": "2025-01-16T14:30:00Z",
        "version": 1,
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
        "encrypted_file_size_in_bytes": 1048576,
        "encrypted_thumbnail_size_in_bytes": 4096,
        "created_at": "2025-01-16T10:15:00Z",
        "modified_at": "2025-01-16T13:45:00Z",
        "version": 2,
        "state": "active"
      }
    ]
  },
  "success": true,
  "message": "Dashboard data retrieved successfully"
}
```

### Error Response (401 Unauthorized)
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

### Error Response (500 Internal Server Error)
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to retrieve dashboard data"
  }
}
```

## Response Field Descriptions

### Dashboard Object
| Field | Type | Description |
|-------|------|-------------|
| `dashboard` | object | Main dashboard data container |
| `success` | boolean | Whether the request was successful |
| `message` | string | Human-readable response message |

### Summary Object
| Field | Type | Description |
|-------|------|-------------|
| `total_files` | integer | Total number of active files accessible to the user |
| `total_folders` | integer | Total number of collections (folders) owned by the user |
| `storage_used` | StorageAmount | Current storage usage |
| `storage_limit` | StorageAmount | Maximum storage allowed for the user |
| `storage_usage_percentage` | integer | Percentage of storage used (0-100) |

### StorageAmount Object
| Field | Type | Description |
|-------|------|-------------|
| `value` | float | Numeric value of the storage amount |
| `unit` | string | Unit of measurement ("B", "KB", "MB", "GB", "TB") |

### StorageUsageTrend Object
| Field | Type | Description |
|-------|------|-------------|
| `period` | string | Time period description (e.g., "Last 7 days") |
| `data_points` | array[DataPoint] | Array of daily storage usage data points |

### DataPoint Object
| Field | Type | Description |
|-------|------|-------------|
| `date` | string | Date in YYYY-MM-DD format |
| `usage` | StorageAmount | Storage usage for that date |

### Recent Files Array
The `recent_files` array contains up to 5 of the most recently modified files. Each file object includes:

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique file identifier |
| `collection_id` | UUID | ID of the collection containing this file |
| `owner_id` | UUID | ID of the file owner |
| `encrypted_metadata` | string | Base64-encoded encrypted file metadata |
| `encrypted_file_key` | object | Encrypted file encryption key |
| `encryption_version` | string | Version of encryption scheme used |
| `encrypted_hash` | string | Hash of encrypted file content |
| `encrypted_file_size_in_bytes` | integer | Size of encrypted file in bytes |
| `encrypted_thumbnail_size_in_bytes` | integer | Size of encrypted thumbnail in bytes |
| `created_at` | string | File creation timestamp (ISO 8601) |
| `modified_at` | string | Last modification timestamp (ISO 8601) |
| `version` | integer | File version number |
| `state` | string | File state (always "active" for recent files) |

### EncryptedFileKey Object
| Field | Type | Description |
|-------|------|-------------|
| `ciphertext` | string | Base64-encoded encrypted key |
| `nonce` | string | Base64-encoded nonce used for encryption |
| `key_version` | integer | Version of the key encryption |

## Data Sources
The dashboard aggregates data from:
- **User Information**: From federated user service for storage quotas and limits
- **File Count**: From file metadata repository counting user's accessible files
- **Folder Count**: From collection repository counting user's owned collections only
- **Storage Trend**: From storage daily usage for the last 7 days
- **Recent Files**: From file metadata repository (last 5 modified files)

## Implementation Notes

### Architecture
- **Service**: `internal/maplefile/service/dashboard/`
- **DTOs**: `internal/maplefile/service/dashboard/dto.go`
- **HTTP Handler**: `internal/maplefile/interface/http/dashboard/`

### Dependencies
- FederatedUserGetByIDUseCase
- CountUserFilesUseCase
- CountUserCollectionsUseCase
- GetStorageDailyUsageTrendUseCase
- ListRecentFilesService

### Storage Units
Storage amounts are automatically converted to human-readable format:
- Bytes (B)
- Kilobytes (KB) - 1,024 bytes
- Megabytes (MB) - 1,024 KB
- Gigabytes (GB) - 1,024 MB
- Terabytes (TB) - 1,024 GB

### Recent Files
Recent files are retrieved using the same service as the dedicated recent files endpoint, ensuring consistency across the API. Files are:
- Limited to the 5 most recently modified
- Only include active files from accessible collections
- Ordered by `modified_at` descending
- Fully encrypted with only metadata visible to the backend

### JSON Format
All field names use snake_case formatting:
- ✅ `total_files`, `storage_usage_trend`, `recent_files`
- ❌ `totalFiles`, `storageUsageTrend`, `recentFiles`

## Testing

### Using curl:
```bash
curl -X GET \
  http://localhost:8080/maplefile/api/v1/dashboard \
  -H "Authorization: JWT your_jwt_token_here" \
  -H "Content-Type: application/json"
```

### Using HTTPie:
```bash
http GET localhost:8080/maplefile/api/v1/dashboard \
  Authorization:"JWT your_jwt_token_here"
```

## Error Handling
The endpoint includes comprehensive error handling:
- Authentication validation
- User existence verification
- Graceful degradation for non-critical data (trends, recent files)
- Proper HTTP status codes
- Structured error responses

### Common Issues and Solutions

#### Storage Usage Percentage Always Zero
- Check if `storage_used_bytes` and `storage_limit_bytes` are properly set
- Verify that file uploads/deletions are updating storage usage tracking
- Ensure user has a storage limit configured

#### Incorrect Folder Count
- Verify that collection counting only includes owned collections, not shared ones
- Check the `access_type` filtering in the collection count query

#### Empty Recent Files
- Confirm that recent files service is working independently
- Check that user has uploaded files and they are in "active" state
- Verify collection access permissions

## Security Notes
- Only returns data for authenticated users
- File content remains encrypted; only metadata is accessible
- Respects collection-level permissions
- No sensitive file content is exposed through this endpoint
- All storage calculations respect user's accessible collections only
