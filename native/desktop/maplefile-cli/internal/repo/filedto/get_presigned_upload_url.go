// monorepo/native/desktop/maplefile-cli/internal/repo/filedto/get_presigned_upload_url.go
package filedto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
)

// GetPresignedUploadURL generates new presigned upload URLs for an existing file
func (r *fileDTORepository) GetPresignedUploadURL(ctx context.Context, fileID primitive.ObjectID, request *filedto.GetPresignedUploadURLRequest) (*filedto.GetPresignedUploadURLResponse, error) {
	r.logger.Debug("Getting presigned upload URL",
		zap.String("fileID", fileID.Hex()),
		zap.Duration("urlDuration", request.URLDuration))

	if fileID.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Set default URL duration if not provided
	if request.URLDuration == 0 {
		request.URLDuration = 1 * time.Hour
	}

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

	// Create request body with duration in nanoseconds as expected by the server
	requestBody := map[string]interface{}{
		"url_duration": fmt.Sprintf("%d", request.URLDuration.Nanoseconds()),
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, errors.NewAppError("failed to marshal request", err)
	}

	// Create HTTP request
	requestURL := fmt.Sprintf("%s/maplefile/api/v1/files/%s/upload-url", serverURL, fileID.Hex())
	r.logger.Debug("Making HTTP request", zap.String("url", requestURL))

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

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
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	var response filedto.GetPresignedUploadURLResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("Successfully obtained presigned upload URLs",
		zap.String("fileID", fileID.Hex()),
		zap.Time("urlExpiration", response.UploadURLExpirationTime))

	return &response, nil
}
