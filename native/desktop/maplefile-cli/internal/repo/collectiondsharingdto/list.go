// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondsharingdto/list.go
package collectiondsharingdto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
)

// ListSharedCollectionsFromCloud gets all collections that have been shared with the authenticated user
func (r *collectionSharingDTORepository) ListSharedCollectionsFromCloud(ctx context.Context) ([]*collectiondto.CollectionDTO, error) {
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

	// Create HTTP request
	listURL := fmt.Sprintf("%s/maplefile/api/v1/collections/shared", serverURL)
	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request", zap.String("url", listURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute list shared collections request", zap.Error(err))
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
	if resp.StatusCode != http.StatusOK {
		r.logger.Error("🚨 Server returned error status",
			zap.String("status", resp.Status),
			zap.ByteString("body", body))
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse response
	var response struct {
		Collections []struct {
			ID             gocql.UUID `json:"id"`
			OwnerID        gocql.UUID `json:"owner_id"`
			EncryptedName  string     `json:"encrypted_name"`
			CollectionType string     `json:"collection_type"`
			CreatedAt      string     `json:"created_at"`
			ModifiedAt     string     `json:"modified_at"`
		} `json:"collections"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		r.logger.Error("❌ Failed to parse shared collections response", zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Convert to domain models
	collections := make([]*collectiondto.CollectionDTO, len(response.Collections))
	for i, coll := range response.Collections {
		collections[i] = &collectiondto.CollectionDTO{
			ID:             coll.ID,
			OwnerID:        coll.OwnerID,
			EncryptedName:  coll.EncryptedName,
			CollectionType: coll.CollectionType,
			// Name:           "[Encrypted]", // Would need decryption logic here //TODO: IMPL.
			// SyncStatus:     collectiondto.SyncStatusSynced, //TODO: IMPL.
		}
	}

	r.logger.Info("✅ Successfully retrieved shared collections",
		zap.Int("count", len(collections)))

	return collections, nil
}
