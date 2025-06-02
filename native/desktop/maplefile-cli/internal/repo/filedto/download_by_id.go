// native/desktop/maplefile-cli/internal/repo/filedto/download_by_id.go
package filedto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
)

// DownloadByIDFromCloud downloads a FileDTO by its unique identifier from the cloud service
func (r *fileDTORepository) DownloadByIDFromCloud(ctx context.Context, id primitive.ObjectID) (*filedto.FileDTO, error) {
	r.logger.Debug("⬇️ Downloading file metadata from cloud", zap.String("fileID", id.Hex()))

	if id.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
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

	// Create HTTP request
	requestURL := fmt.Sprintf("%s/maplefile/api/v1/files/%s", serverURL, id.Hex())
	r.logger.Debug("🌐 Making HTTP request", zap.String("url", requestURL))

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
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
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NewAppError("file not found", nil)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// FIXED: Try parsing as direct FileDTO first, then as wrapped response
	var fileDTO *filedto.FileDTO

	// First, try parsing as direct FileDTO (most likely case)
	if err := json.Unmarshal(body, &fileDTO); err == nil && fileDTO != nil && !fileDTO.ID.IsZero() {
		r.logger.Info("✅ Successfully downloaded file metadata from cloud (direct format)",
			zap.String("fileID", id.Hex()),
			zap.String("collectionID", fileDTO.CollectionID.Hex()))
		return fileDTO, nil
	}

	// If that fails, try parsing as wrapped response
	var fileResponse struct {
		File *filedto.FileDTO `json:"file"`
	}
	if err := json.Unmarshal(body, &fileResponse); err == nil && fileResponse.File != nil {
		r.logger.Info("✅ Successfully downloaded file metadata from cloud (wrapped format)",
			zap.String("fileID", id.Hex()),
			zap.String("collectionID", fileResponse.File.CollectionID.Hex()))
		return fileResponse.File, nil
	}

	// If both parsing attempts fail, log the raw response for debugging
	r.logger.Error("❌ Failed to parse file response in any expected format",
		zap.String("fileID", id.Hex()),
		zap.String("rawResponse", string(body)))

	return nil, errors.NewAppError("failed to parse file response - unexpected format", nil)
}
