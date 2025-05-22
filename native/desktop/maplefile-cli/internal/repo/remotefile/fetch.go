// internal/repo/remotefile/fetch.go
package remotefile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Fetch retrieves a file from the remote server by ID
func (r *remoteFileRepository) Fetch(ctx context.Context, id primitive.ObjectID) (*remotefile.RemoteFile, error) {
	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get access token
	accessToken, err := r.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	fetchURL := fmt.Sprintf("%s/maplefile/api/v1/files/%s", serverURL, id.Hex())
	req, err := http.NewRequestWithContext(ctx, "GET", fetchURL, nil)
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	var response remotefile.RemoteFileResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Convert the response to a RemoteFile
	remoteFile := &remotefile.RemoteFile{
		ID:                    response.ID,
		CollectionID:          response.CollectionID,
		OwnerID:               response.OwnerID,
		EncryptedFileID:       response.EncryptedFileID,
		FileObjectKey:         response.FileObjectKey,
		FileSize:              response.FileSize,
		EncryptedOriginalSize: response.EncryptedOriginalSize,
		EncryptedMetadata:     response.EncryptedMetadata,
		EncryptedFileKey:      response.EncryptedFileKey,
		EncryptionVersion:     response.EncryptionVersion,
		EncryptedHash:         response.EncryptedHash,
		ThumbnailObjectKey:    response.ThumbnailObjectKey,
		CreatedAt:             response.CreatedAt,
		ModifiedAt:            response.ModifiedAt,
		Status:                remotefile.FileStatusAvailable, // Default to available
	}

	return remoteFile, nil
}
