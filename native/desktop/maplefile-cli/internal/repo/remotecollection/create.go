// monorepo/native/desktop/maplefile-cli/internal/repo/remotecollection/create.go
package remotecollection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
)

// CreateRemoteCollection creates a new collection by calling the server API
func (r *collectionRepository) Create(ctx context.Context, request *collection.RemoteCreateCollectionRequest) (*collection.RemoteCollectionResponse, error) {
	r.logger.Debug("Starting CreateRemoteCollection process")

	// Get server URL from configuration
	r.logger.Debug("Getting cloud provider address from config")
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("Failed to get cloud provider address from config", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}
	r.logger.Debug("Successfully retrieved cloud provider address", zap.String("serverURL", serverURL))

	// Get authenticated user's email
	r.logger.Debug("Getting authenticated user email from config")
	email, err := r.configService.GetEmail(ctx)
	if err != nil {
		r.logger.Error("Failed to get authenticated user email from config", zap.Error(err))
		return nil, errors.NewAppError("failed to get authenticated user", err)
	}
	r.logger.Debug("Successfully retrieved authenticated user email", zap.String("email", email))

	// Get user data to retrieve auth token
	r.logger.Debug("Getting user data by email", zap.String("email", email))
	userData, err := r.userRepo.GetByEmail(ctx, email)
	if err != nil {
		r.logger.Error("Failed to retrieve user data from repository", zap.String("email", email), zap.Error(err))
		return nil, errors.NewAppError("failed to retrieve user data", err)
	}

	if userData == nil {
		r.logger.Error("User data not found for email", zap.String("email", email))
		return nil, errors.NewAppError("user not found; please login first", nil)
	}
	r.logger.Debug("Successfully retrieved user data")

	// Check if access token is valid
	r.logger.Debug("Checking if access token is valid")
	if userData.AccessToken == "" || time.Now().After(userData.AccessTokenExpiryTime) {
		r.logger.Info("Access token is invalid or expired, attempting to refresh")
		if err := r.refreshTokenIfNeeded(ctx, userData); err != nil {
			r.logger.Error("Failed to refresh authentication token", zap.String("email", email), zap.Error(err))
			return nil, errors.NewAppError("authentication token has expired; please refresh token", nil)
		}
		r.logger.Info("Successfully refreshed authentication token")
	} else {
		r.logger.Debug("Access token is valid")
	}

	// Convert request to JSON
	r.logger.Debug("Marshalling create collection request to JSON")
	jsonData, err := json.Marshal(request)
	if err != nil {
		r.logger.Error("Failed to marshal request to JSON", zap.Any("request", request), zap.Error(err))
		return nil, errors.NewAppError("failed to marshal request", err)
	}
	r.logger.Debug("Successfully marshalled request to JSON")

	// Create HTTP request
	createURL := fmt.Sprintf("%s/maplefile/api/v1/collections", serverURL)
	r.logger.Info("Making HTTP request to create collection",
		zap.String("method", "POST"),
		zap.String("url", createURL),
		zap.String("email", email))

	req, err := http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		r.logger.Error("Failed to create HTTP request for creating collection", zap.String("url", createURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "JWT "+userData.AccessToken) // Log this carefully if sensitive
	r.logger.Debug("HTTP request headers set")

	// Execute the request
	r.logger.Debug("Executing HTTP request to create collection")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("Failed to execute HTTP request to create collection", zap.String("url", createURL), zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()
	r.logger.Info("Received HTTP response", zap.String("status", resp.Status), zap.Int("statusCode", resp.StatusCode))

	// Read response body
	r.logger.Debug("Reading HTTP response body")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("Failed to read HTTP response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}
	r.logger.Debug("Successfully read HTTP response body", zap.ByteString("body", body)) // Be careful logging raw body if sensitive

	// Check for error status codes
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		r.logger.Error("Server returned an error status code", zap.String("status", resp.Status), zap.Int("statusCode", resp.StatusCode), zap.ByteString("body", body))
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Error("Server returned error message in response body", zap.String("message", errMsg))
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}
	r.logger.Debug("HTTP response status is successful")

	// Parse the response
	r.logger.Debug("Parsing HTTP response body into RemoteCollectionResponse")
	var response collection.RemoteCollectionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("Failed to parse response body into RemoteCollectionResponse", zap.ByteString("body", body), zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}
	r.logger.Debug("Successfully parsed HTTP response body")

	r.logger.Info("Successfully created collection", zap.Any("response", response))
	return &response, nil
}
