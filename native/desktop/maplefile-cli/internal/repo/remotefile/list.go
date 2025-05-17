// internal/repo/remotefile/list.go
package remotefile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ListByCollection lists files within a specific collection
func (r *remoteFileRepository) ListByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*remotefile.RemoteFile, error) {
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
	listURL := fmt.Sprintf("%s/maplefile/api/v1/collections/%s/files", serverURL, collectionID.Hex())
	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
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
	var response struct {
		Files []remotefile.RemoteFileResponse `json:"files"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Convert responses to RemoteFile objects
	files := make([]*remotefile.RemoteFile, 0, len(response.Files))
	for _, fileResp := range response.Files {
		file := &remotefile.RemoteFile{
			ID:                    fileResp.ID,
			CollectionID:          fileResp.CollectionID,
			OwnerID:               fileResp.OwnerID,
			EncryptedFileID:       fileResp.EncryptedFileID,
			FileObjectKey:         fileResp.FileObjectKey,
			EncryptedSize:         fileResp.EncryptedSize,
			EncryptedOriginalSize: fileResp.EncryptedOriginalSize,
			EncryptedMetadata:     fileResp.EncryptedMetadata,
			EncryptedFileKey:      fileResp.EncryptedFileKey,
			EncryptionVersion:     fileResp.EncryptionVersion,
			EncryptedHash:         fileResp.EncryptedHash,
			ThumbnailObjectKey:    fileResp.ThumbnailObjectKey,
			CreatedAt:             fileResp.CreatedAt,
			ModifiedAt:            fileResp.ModifiedAt,
			Status:                remotefile.FileStatusAvailable, // Default to available
		}
		files = append(files, file)
	}

	return files, nil
}

// List lists remote files based on filter criteria
func (r *remoteFileRepository) List(ctx context.Context, filter remotefile.RemoteFileFilter) ([]*remotefile.RemoteFile, error) {
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

	// Create HTTP request with query parameters
	listURL := fmt.Sprintf("%s/maplefile/api/v1/files", serverURL)

	// Add query parameters based on filter
	queryParams := make([]string, 0)
	if filter.CollectionID != nil {
		queryParams = append(queryParams, fmt.Sprintf("collection_id=%s", filter.CollectionID.Hex()))
	}
	if filter.OwnerID != nil {
		queryParams = append(queryParams, fmt.Sprintf("owner_id=%s", filter.OwnerID.Hex()))
	}
	if filter.Status != nil {
		queryParams = append(queryParams, fmt.Sprintf("status=%d", *filter.Status))
	}

	// Append query parameters to URL
	if len(queryParams) > 0 {
		listURL += "?" + strings.Join(queryParams, "&")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
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
	var response struct {
		Files []remotefile.RemoteFileResponse `json:"files"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Convert responses to RemoteFile objects
	files := make([]*remotefile.RemoteFile, 0, len(response.Files))
	for _, fileResp := range response.Files {
		file := &remotefile.RemoteFile{
			ID:                    fileResp.ID,
			CollectionID:          fileResp.CollectionID,
			OwnerID:               fileResp.OwnerID,
			EncryptedFileID:       fileResp.EncryptedFileID,
			FileObjectKey:         fileResp.FileObjectKey,
			EncryptedSize:         fileResp.EncryptedSize,
			EncryptedOriginalSize: fileResp.EncryptedOriginalSize,
			EncryptedMetadata:     fileResp.EncryptedMetadata,
			EncryptedFileKey:      fileResp.EncryptedFileKey,
			EncryptionVersion:     fileResp.EncryptionVersion,
			EncryptedHash:         fileResp.EncryptedHash,
			ThumbnailObjectKey:    fileResp.ThumbnailObjectKey,
			CreatedAt:             fileResp.CreatedAt,
			ModifiedAt:            fileResp.ModifiedAt,
			Status:                remotefile.FileStatusAvailable,
		}
		files = append(files, file)
	}

	return files, nil
}
