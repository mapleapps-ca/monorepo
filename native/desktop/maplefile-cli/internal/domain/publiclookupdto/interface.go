package publiclookupdto

import (
	"context"
)

type PublicLookupDTORepository interface {
	// CreatePendingFileInCloud creates a pending file record in the cloud and returns
	// presigned URLs for uploading the file content directly to cloud storage.
	// This is Step 1 of the three-step upload process.
	GetFromCloud(ctx context.Context, request *PublicLookupRequestDTO) (*PublicLookupResponseDTO, error)
}
