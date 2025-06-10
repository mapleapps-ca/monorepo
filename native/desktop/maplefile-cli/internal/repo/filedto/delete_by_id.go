// monorepo/native/desktop/maplefile-cli/internal/repo/filedto/delete_by_id.go
package filedto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// DeleteByIDFromCloud deletes a FileDTO by its unique identifier from the cloud service
func (r *fileDTORepository) DeleteByIDFromCloud(ctx context.Context, id gocql.UUID) error {
	r.logger.Debug("⚙️ Deleting file from cloud", zap.String("fileID", id.String()))

	if id.String() == "" {
		return errors.NewAppError("file ID is required", nil)
	}

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get access token for authentication
	accessToken, err := r.tokenRepo.GetAccessToken(ctx)
	if err != nil {
		return errors.NewAppError("failed to get access token", err)
	}

	// Create HTTP request
	requestURL := fmt.Sprintf("%s/maplefile/api/v1/files/%s", serverURL, id.String())
	r.logger.Debug("⚙️ Making HTTP request", zap.String("url", requestURL))

	req, err := http.NewRequestWithContext(ctx, "DELETE", requestURL, nil)
	if err != nil {
		return errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("JWT %s", accessToken))

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode == http.StatusNotFound {
		r.logger.Warn("⚠️ File not found in cloud",
			zap.String("fileID", id.String()),
			zap.String("responseBody", string(body)))
		return errors.NewAppError("file not found in cloud", nil)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Error("❌ Server error during file deletion",
					zap.String("fileID", id.String()),
					zap.String("serverError", errMsg),
					zap.Int("statusCode", resp.StatusCode))
				return errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		r.logger.Error("❌ Unexpected server response during file deletion",
			zap.String("fileID", id.String()),
			zap.Int("statusCode", resp.StatusCode),
			zap.String("responseBody", string(body)))
		return errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	r.logger.Info("✅ Successfully deleted file from cloud", zap.String("fileID", id.String()))
	return nil
}
