// internal/repo/collection/collection.go
package collection

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
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// collectionRepository implements the collection.CollectionRepository interface
type collectionRepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	userRepo      user.Repository
	httpClient    *http.Client
}

// NewCollectionRepository creates a new repository for collection operations
func NewCollectionRepository(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
) collection.CollectionRepository {
	return &collectionRepository{
		logger:        logger,
		configService: configService,
		userRepo:      userRepo,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateCollection creates a new collection by calling the server API
func (r *collectionRepository) CreateCollection(ctx context.Context, request *collection.CreateCollectionRequest) (*collection.CollectionResponse, error) {
	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get authenticated user's email
	email, err := r.configService.GetEmail(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get authenticated user", err)
	}

	// Get user data to retrieve auth token
	userData, err := r.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.NewAppError("failed to retrieve user data", err)
	}

	if userData == nil {
		return nil, errors.NewAppError("user not found; please login first", nil)
	}

	// Check if access token is valid
	if userData.AccessToken == "" || time.Now().After(userData.AccessTokenExpiryTime) {
		return nil, errors.NewAppError("authentication token has expired; please refresh token", nil)
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewAppError("failed to marshal request", err)
	}

	// Create HTTP request
	createURL := fmt.Sprintf("%s/maplefile/api/v1/collections", serverURL)
	r.logger.Debug("Making HTTP request to create collection",
		zap.String("url", createURL),
		zap.String("email", email))

	req, err := http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userData.AccessToken)

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
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	var response collection.CollectionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.NewAppError("failed to parse response", err)
	}

	return &response, nil
}
