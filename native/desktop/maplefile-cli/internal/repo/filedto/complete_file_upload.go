// monorepo/native/desktop/maplefile-cli/internal/repo/filedto/complete_file_upload.go
package filedto

import (
	"bytes"
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

// CompleteFileUploadInCloud completes the file upload process and transitions the file to active state
func (r *fileDTORepository) CompleteFileUploadInCloud(ctx context.Context, fileID gocql.UUID, request *filedto.CompleteFileUploadRequest) (*filedto.CompleteFileUploadResponse, error) {
	r.logger.Debug("🐛 Completing file upload",
		zap.String("fileID", fileID.Hex()),
		zap.Int64("actualFileSize", request.ActualFileSizeInBytes),
		zap.Bool("uploadConfirmed", request.UploadConfirmed))

	if fileID.IsZero() {
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

	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewAppError("failed to marshal request", err)
	}

	// Create HTTP request
	requestURL := fmt.Sprintf("%s/maplefile/api/v1/files/%s/complete", serverURL, fileID.Hex())
	r.logger.Debug("🔬 Making HTTP request", zap.String("url", requestURL))

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		r.logger.Debug("❌ failed to create HTTP request",
			zap.String("fileID", fileID.Hex()),
			zap.Error(err),
		)
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("JWT %s", accessToken))

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Debug("❌ failed to connect to server",
			zap.String("fileID", fileID.Hex()),
			zap.Error(err),
		)
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Debug("❌ failed to read response",
			zap.String("fileID", fileID.Hex()),
			zap.Error(err),
		)
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Debug("⚠️ server error",
					zap.String("fileID", fileID.Hex()),
					zap.String("error", errMsg),
				)
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		r.logger.Debug("⚠️ server returned error status",
			zap.String("fileID", fileID.Hex()),
			zap.Error(err),
		)
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s | message: %s", resp.Status, string(body)), nil)
	}

	// Parse the response
	var response filedto.CompleteFileUploadResponse
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Debug("❌ failed to parse response",
			zap.String("fileID", fileID.Hex()),
			zap.Error(err),
		)
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("✅ Successfully completed file upload",
		zap.String("fileID", fileID.Hex()),
		zap.Int64("actualFileSize", response.ActualFileSize),
		zap.Bool("uploadVerified", response.UploadVerified))

	return &response, nil
}
