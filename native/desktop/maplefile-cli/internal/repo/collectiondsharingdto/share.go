// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondsharingdto/share.go
package collectiondsharingdto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
)

// ShareCollectionInCloud shares a collection with another user
func (r *collectionSharingDTORepository) ShareCollectionInCloud(ctx context.Context, request *collectionsharingdto.ShareCollectionRequestDTO) (*collectionsharingdto.ShareCollectionResponseDTO, error) {
	accessToken, err := r.tokenRepository.GetAccessToken(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get access token", zap.Error(err))
		return nil, errors.NewAppError("failed to get access token", err)
	}

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("❌ Failed to get cloud provider address", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Prepare request body according to API spec
	requestBody := map[string]interface{}{
		"recipient_id":             request.RecipientID.Hex(),
		"recipient_email":          request.RecipientEmail,
		"permission_level":         request.PermissionLevel,
		"encrypted_collection_key": request.EncryptedCollectionKey,
		"share_with_descendants":   request.ShareWithDescendants,
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		r.logger.Error("❌ Failed to marshal share request", zap.Error(err))
		return nil, errors.NewAppError("failed to marshal request", err)
	}

	// Create HTTP request
	shareURL := fmt.Sprintf("%s/maplefile/api/v1/collections/%s/share", serverURL, request.CollectionID.Hex())
	r.logger.Info("➡️ Making HTTP request to share collection",
		zap.String("method", "POST"),
		zap.String("url", shareURL),
		zap.String("recipient", request.RecipientEmail))

	req, err := http.NewRequestWithContext(ctx, "POST", shareURL, bytes.NewBuffer(jsonData))
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request", zap.String("url", shareURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute share request", zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("❌ Failed to read response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		r.logger.Error("🚨 Server returned error status",
			zap.String("status", resp.Status),
			zap.ByteString("body", body))

		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse successful response
	var response collectionsharingdto.ShareCollectionResponseDTO
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("❌ Failed to parse share response", zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("✅ Successfully shared collection",
		zap.String("collectionID", request.CollectionID.Hex()),
		zap.String("recipient", request.RecipientEmail),
		zap.String("permission", request.PermissionLevel))

	return &response, nil
}
