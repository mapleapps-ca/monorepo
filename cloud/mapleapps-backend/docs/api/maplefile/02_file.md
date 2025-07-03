# MapleFile File Sync API Documentation

## List File Sync Data

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
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "collection_id": "550e8400-e29b-41d4-a716-446655440004",
      "version": 2,
      "modified_at": "2023-12-01T14:20:00Z",
      "state": "archived",
      "tombstone_version": 0,
      "tombstone_expiry": "0001-01-01T00:00:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440005",
      "collection_id": "550e8400-e29b-41d4-a716-446655440006",
      "version": 1,
      "modified_at": "2023-12-01T13:10:00Z",
      "state": "pending",
      "tombstone_version": 0,
      "tombstone_expiry": "0001-01-01T00:00:00Z"
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

### File States

| State | Description |
|-------|-------------|
| `pending` | File metadata created but upload not completed |
| `active` | File fully uploaded and available |
| `deleted` | File soft-deleted (tombstoned) |
| `archived` | File archived but still accessible |

### Sync Logic

1. **Initial Sync**: Call without cursor to get all files
2. **Incremental Sync**: Use `next_cursor` from previous response
3. **Version Tracking**: Use `version` field for conflict resolution
4. **Tombstone Handling**:
   - Files with `state: "deleted"` should be removed locally
   - `tombstone_version` indicates when deletion occurred
   - `tombstone_expiry` shows when tombstone will be permanently removed

### Cursor Format

The cursor is a base64-encoded JSON object containing:
```json
{
  "last_modified": "2023-12-01T15:30:00Z",
  "last_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Error Responses
- **400 Bad Request**: Invalid cursor format or parameters
- **401 Unauthorized**: Authentication required
- **500 Internal Server Error**: Server error

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

// Initial sync
let syncResponse = await syncFiles();
console.log(`Synced ${syncResponse.files.length} files`);

// Process files
syncResponse.files.forEach(file => {
  console.log(`File ${file.id}: version ${file.version}, state ${file.state}`);

  if (file.state === 'deleted') {
    // Remove from local storage
    localStorage.removeItem(`file_${file.id}`);
  } else {
    // Update local file metadata
    localStorage.setItem(`file_${file.id}`, JSON.stringify({
      version: file.version,
      state: file.state,
      modified_at: file.modified_at
    }));
  }
});

// Continue syncing if more data available
while (syncResponse.has_more) {
  syncResponse = await syncFiles(syncResponse.next_cursor);
  // Process additional files...
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
    ID               string    `json:"id"`
    CollectionID     string    `json:"collection_id"`
    Version          uint64    `json:"version"`
    ModifiedAt       time.Time `json:"modified_at"`
    State            string    `json:"state"`
    TombstoneVersion uint64    `json:"tombstone_version"`
    TombstoneExpiry  time.Time `json:"tombstone_expiry"`
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

// Usage
func (c *FileSyncClient) PerformFullSync() error {
    var cursor *string

    for {
        resp, err := c.SyncFiles(cursor, 1000)
        if err != nil {
            return err
        }

        // Process files
        for _, file := range resp.Files {
            fmt.Printf("File %s: version %d, state %s\n",
                file.ID, file.Version, file.State)

            if file.State == "deleted" {
                // Handle file deletion
                c.handleFileDeleted(file.ID)
            } else {
                // Update local file state
                c.updateLocalFile(file)
            }
        }

        if !resp.HasMore {
            break
        }
        cursor = resp.NextCursor
    }

    return nil
}
```

### Performance Considerations

1. **Pagination**: Use appropriate `limit` values (1000-5000 recommended)
2. **Cursor Storage**: Store cursors locally for resumable sync
3. **State Filtering**: Handle different file states appropriately
4. **Tombstone Cleanup**: Remove expired tombstones from local storage
5. **Version Conflicts**: Use version numbers for conflict resolution

### Security Notes

- Only returns files from collections the user has access to
- Automatically filters based on user permissions
- Cursor tokens are signed and expire after a reasonable time
- Rate limiting applies to prevent abuse
