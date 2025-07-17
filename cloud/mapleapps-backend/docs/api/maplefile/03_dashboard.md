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
      "totalFiles": 156,
      "totalFolders": 23,
      "storageUsed": {
        "value": 4.2,
        "unit": "GB"
      },
      "storageLimit": {
        "value": 15,
        "unit": "GB"
      },
      "storageUsagePercentage": 28
    },
    "storageUsageTrend": {
      "period": "Last 7 days",
      "dataPoints": [
        {
          "date": "2025-01-10",
          "usage": {
            "value": 3.8,
            "unit": "GB"
          }
        }
      ]
    },
    "recentFiles": [
      {
        "fileName": "Budget_Report_Q4.xlsx",
        "uploaded": "2 hours ago",
        "uploadedTimestamp": "2025-01-16T14:30:00Z",
        "type": "Spreadsheet",
        "size": {
          "value": 2.4,
          "unit": "MB"
        }
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

## Data Sources
The dashboard aggregates data from:
- **User Information**: From federated user service for storage quotas
- **File Count**: From file metadata repository counting user's accessible files
- **Folder Count**: From collection repository counting user's collections
- **Storage Trend**: From storage daily usage for the last 7 days
- **Recent Files**: From file metadata repository (last 5 files)

## Implementation Notes

### Architecture
- **Domain**: `internal/maplefile/domain/dashboard/`
- **Use Case**: `internal/maplefile/usecase/dashboard/`
- **Service**: `internal/maplefile/service/dashboard/`
- **HTTP Handler**: `internal/maplefile/interface/http/dashboard/`

### Dependencies
- FederatedUserGetByIDUseCase
- CountUserFilesUseCase
- CountUserCollectionsUseCase
- GetStorageDailyUsageTrendUseCase
- ListRecentFilesUseCase

### Storage Units
Storage amounts are automatically converted to human-readable format:
- Bytes (B)
- Kilobytes (KB) - 1,024 bytes
- Megabytes (MB) - 1,024 KB
- Gigabytes (GB) - 1,024 MB
- Terabytes (TB) - 1,024 GB

### File Types
Recent files are categorized based on file extensions:
- Image: jpg, jpeg, png, gif, bmp, svg
- Video: mp4, avi, mov, wmv, flv, webm
- Audio: mp3, wav, flac, aac, ogg
- PDF: pdf
- Word Document: doc, docx
- Spreadsheet: xls, xlsx
- Presentation: ppt, pptx
- Text: txt
- Archive: zip, rar, 7z, tar, gz
- Document: default category

### Time Formatting
Upload times are formatted as:
- "Just now" - less than 1 minute
- "X minutes ago" - less than 1 hour
- "X hours ago" - less than 24 hours
- "X days ago" - less than 7 days
- "Jan 2, 2006" - older than 7 days

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
