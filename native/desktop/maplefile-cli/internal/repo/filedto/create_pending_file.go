// monorepo/native/desktop/maplefile-cli/internal/repo/filedto/create_pending_file.go
package filedto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
)

// CreatePendingFileInCloud creates a pending file record in the cloud and returns presigned URLs
func (r *fileDTORepository) CreatePendingFileInCloud(ctx context.Context, request *filedto.CreatePendingFileRequest) (*filedto.CreatePendingFileResponse, error) {
	r.logger.Debug("📝 Creating pending file in cloud",
		zap.String("fileID", request.ID.Hex()),
		zap.String("collectionID", request.CollectionID.Hex()),
		zap.Int64("expectedFileSize", request.ExpectedFileSizeInBytes))

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get access token for authentication
	accessToken, err := r.tokenRepo.GetAccessToken(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get access token", err)
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewAppError("failed to marshal request", err)
	}

	// Create HTTP request
	requestURL := fmt.Sprintf("%s/maplefile/api/v1/files/pending", serverURL)
	r.logger.Debug("📡 Making HTTP request", zap.String("url", requestURL))

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("JWT %s", accessToken))

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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s | reason: %s", resp.Status, string(body)), nil)
	}

	// Parse the response
	var response filedto.CreatePendingFileResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("✅ Successfully created pending file",
		zap.String("fileID", response.File.ID.Hex()),
		zap.Time("urlExpiration", response.UploadURLExpirationTime))

	return &response, nil
}
