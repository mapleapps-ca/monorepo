// monorepo/native/desktop/maplefile-cli/internal/repo/remotecollection/list.go
package remotecollection

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
)

func (r *collectionRepository) List(ctx context.Context, filter collection.CollectionFilter) ([]*collection.RemoteCollection, error) {
	r.logger.Debug("Listing collections from cloud server", zap.Any("filter", filter))

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("Failed to get cloud provider address", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get authenticated user's access token
	accessToken, err := r.getAccessToken(ctx)
	if err != nil {
		r.logger.Error("Failed to get access token", zap.Error(err))
		return nil, err
	}

	// Create the collection endpoint URL
	baseURL := fmt.Sprintf("%s/maplefile/api/v1/collections", serverURL)

	// Add query parameters
	query := url.Values{}
	if filter.ParentID != nil && !filter.ParentID.IsZero() {
		query.Add("parent_id", filter.ParentID.Hex())
	}
	if filter.Type != "" {
		query.Add("type", filter.Type)
	}

	// Add query string to URL if we have parameters
	requestURL := baseURL
	if len(query) > 0 {
		requestURL = fmt.Sprintf("%s?%s", baseURL, query.Encode())
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		r.logger.Error("Failed to create HTTP request", zap.String("url", requestURL), zap.Error(err))
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
	var response struct {
		Collections []collection.RemoteCollectionResponse `json:"collections"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("Failed to parse response body", zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Convert to RemoteCollection objects
	collections := make([]*collection.RemoteCollection, 0, len(response.Collections))
	for _, resp := range response.Collections {
		remoteColl := &collection.RemoteCollection{
			ID:                     resp.ID,
			OwnerID:                resp.OwnerID,
			EncryptedName:          resp.EncryptedName,
			Type:                   resp.Type,
			ParentID:               resp.ParentID,
			AncestorIDs:            resp.AncestorIDs,
			EncryptedPathSegments:  resp.EncryptedPathSegments,
			EncryptedCollectionKey: resp.EncryptedCollectionKey,
			CreatedAt:              resp.CreatedAt,
			ModifiedAt:             resp.ModifiedAt,
			// Local fields are not set yet since they need to be decrypted
		}
		collections = append(collections, remoteColl)
	}

	r.logger.Info("Successfully listed collections from cloud server",
		zap.Int("count", len(collections)))
	return collections, nil
}
