// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/download.go
package collectiondto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
)

func (r *collectionDTORepository) GetFromCloudByID(ctx context.Context, id primitive.ObjectID) (*collectiondto.CollectionDTO, error) {
	accessToken, err := r.tokenRepository.GetAccessToken(ctx)
	if err != nil {
		r.logger.Error("Failed to get access token", zap.Error(err))
		return nil, errors.NewAppError("failed to get access token", err)
	}

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("Failed to get cloud provider address", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Defensive programming
	if id.IsZero() {
		r.logger.Error("id is required")
		return nil, errors.NewAppError("id is required", nil)
	}

	// Create HTTP request
	fetchURL := fmt.Sprintf("%s/maplefile/api/v1/collections/%s", serverURL, id.Hex())
	req, err := http.NewRequestWithContext(ctx, "GET", fetchURL, nil)
	if err != nil {
		r.logger.Error("Failed to create HTTP request", zap.String("url", fetchURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("Failed to execute HTTP request", zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("Failed to read response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		r.logger.Error("Server returned an error status code",
			zap.String("status", resp.Status),
			zap.Int("statusCode", resp.StatusCode))
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	var response collectiondto.CollectionDTO
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("Failed to parse response body", zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	r.logger.Info("Successfully fetched collection from cloud server",
		zap.String("collectionID", id.Hex()))
	return &response, nil
}
