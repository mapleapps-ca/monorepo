// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/list.go
package collectiondto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

func (r *collectionDTORepository) ListFromCloud(ctx context.Context, filter collectiondto.CollectionFilter) ([]*collectiondto.CollectionDTO, error) {
	r.logger.Debug("🔍 Listing collections from cloud",
		zap.Any("filter", filter))

	// Get access token
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

	// Build URL with query parameters
	baseURL := fmt.Sprintf("%s/maplefile/api/v1/collections", serverURL)
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		r.logger.Error("❌ Failed to parse base URL", zap.String("url", baseURL), zap.Error(err))
		return nil, errors.NewAppError("failed to parse base URL", err)
	}

	// Add query parameters if filter options are provided
	queryParams := parsedURL.Query()
	if filter.ParentID != nil {
		queryParams.Set("parent_id", filter.ParentID.String())
	}
	if filter.CollectionType != "" {
		queryParams.Set("collection_type", filter.CollectionType)
	}
	parsedURL.RawQuery = queryParams.Encode()

	finalURL := parsedURL.String()
	r.logger.Debug("➡️ Making request to collections list endpoint", zap.String("url", finalURL))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", finalURL, nil)
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request", zap.String("url", finalURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Authorization", "JWT "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	r.logger.Debug("➡️ Executing HTTP request for collections list")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute HTTP request", zap.String("url", finalURL), zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	r.logger.Debug("⬅️ Received HTTP response", zap.String("status", resp.Status), zap.Int("statusCode", resp.StatusCode))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("❌ Failed to read response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		r.logger.Error("🚨 Server returned an error status code",
			zap.String("status", resp.Status),
			zap.Int("statusCode", resp.StatusCode),
			zap.ByteString("body", body))

		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				r.logger.Error("🚨 Server returned error message in response body", zap.String("message", errMsg))
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response - expecting an array of collections
	r.logger.Debug("🔍 Parsing HTTP response body into CollectionDTO array")
	var apiCollections []*CollectionResponseDTO
	if err := json.Unmarshal(body, &apiCollections); err != nil {
		r.logger.Error("❌ Failed to parse response body into CollectionDTO array",
			zap.ByteString("body", body),
			zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Convert API response to domain models
	collections := make([]*collectiondto.CollectionDTO, len(apiCollections))
	for i, apiCollection := range apiCollections {
		collection := &collectiondto.CollectionDTO{
			ID:             apiCollection.ID,
			OwnerID:        apiCollection.OwnerID,
			EncryptedName:  apiCollection.EncryptedName,
			CollectionType: apiCollection.CollectionType,
			ParentID:       apiCollection.ParentID,
			AncestorIDs:    apiCollection.AncestorIDs,
			CreatedAt:      apiCollection.CreatedAt,
			ModifiedAt:     apiCollection.ModifiedAt,
			Members:        make([]*collectiondto.CollectionMembershipDTO, len(apiCollection.Members)),
		}

		// Convert collection's own encrypted key if present
		if apiCollection.EncryptedCollectionKey != nil && len(apiCollection.EncryptedCollectionKey.Ciphertext) > 0 {
			collection.EncryptedCollectionKey = apiCollection.EncryptedCollectionKey
		}

		// Convert members
		for j, member := range apiCollection.Members {
			// Convert bytes to EncryptedCollectionKey struct for member
			var encryptedCollectionKey *keys.EncryptedCollectionKey
			if len(member.EncryptedCollectionKey) > 0 {
				// Create EncryptedCollectionKey from box_seal bytes
				encryptedCollectionKey = keys.NewEncryptedCollectionKeyFromBoxSeal(member.EncryptedCollectionKey)
			}

			collection.Members[j] = &collectiondto.CollectionMembershipDTO{
				ID:                     member.ID,
				CollectionID:           member.CollectionID,
				RecipientID:            member.RecipientID,
				RecipientEmail:         member.RecipientEmail,
				GrantedByID:            member.GrantedByID,
				PermissionLevel:        member.PermissionLevel,
				CreatedAt:              member.CreatedAt,
				IsInherited:            member.IsInherited,
				InheritedFromID:        member.InheritedFromID,
				EncryptedCollectionKey: encryptedCollectionKey,
			}
		}

		collections[i] = collection
	}

	r.logger.Info("✅ Successfully retrieved collections from cloud",
		zap.Int("count", len(collections)),
		zap.Any("filter", filter))

	return collections, nil
}
