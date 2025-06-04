// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/get.go
package collectiondto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

func (r *collectionDTORepository) GetFromCloudByID(ctx context.Context, id primitive.ObjectID) (*collectiondto.CollectionDTO, error) {
	accessToken, err := r.tokenRepository.GetAccessToken(ctx)
	if err != nil {
		r.logger.Error("🚨 Failed to get access token", zap.Error(err))
		return nil, errors.NewAppError("failed to get access token", err)
	}

	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		r.logger.Error("🚨 Failed to get cloud provider address", zap.Error(err))
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Defensive programming
	if id.IsZero() {
		r.logger.Error("🚨 id is required")
		return nil, errors.NewAppError("id is required", nil)
	}

	// Create HTTP request
	fetchURL := fmt.Sprintf("%s/maplefile/api/v1/collections/%s", serverURL, id.Hex())
	req, err := http.NewRequestWithContext(ctx, "GET", fetchURL, nil)
	if err != nil {
		r.logger.Error("🚨 Failed to create HTTP request", zap.String("url", fetchURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	// Set headers
	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("🚨 Failed to execute HTTP request", zap.Error(err))
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("🚨 Failed to read response body", zap.Error(err))
		return nil, errors.NewAppError("failed to read response", err)
	}

	r.logger.Debug("🔍 Raw collection response from API",
		zap.ByteString("responseBody", body))

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		r.logger.Error("🚨 Server returned an error status code",
			zap.String("status", resp.Status),
			zap.Int("statusCode", resp.StatusCode))
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse response using intermediate struct to handle byte arrays from API
	var apiResponse struct {
		ID             primitive.ObjectID `json:"id"`
		OwnerID        primitive.ObjectID `json:"owner_id"`
		EncryptedName  string             `json:"encrypted_name"`
		CollectionType string             `json:"collection_type"`
		// Handle collection's own encrypted key
		EncryptedCollectionKey struct {
			Ciphertext []byte `json:"ciphertext"`
			Nonce      []byte `json:"nonce"`
		} `json:"encrypted_collection_key"`
		Members []struct {
			ID              primitive.ObjectID `json:"id"`
			CollectionID    primitive.ObjectID `json:"collection_id"`
			RecipientID     primitive.ObjectID `json:"recipient_id"`
			RecipientEmail  string             `json:"recipient_email"`
			GrantedByID     primitive.ObjectID `json:"granted_by_id"`
			PermissionLevel string             `json:"permission_level"`
			CreatedAt       string             `json:"created_at"`
			IsInherited     bool               `json:"is_inherited"`
			InheritedFromID primitive.ObjectID `json:"inherited_from_id,omitempty"`
			// Handle member's encrypted collection key as bytes
			EncryptedCollectionKey []byte `json:"encrypted_collection_key"`
		} `json:"members"`
		ParentID         primitive.ObjectID   `json:"parent_id,omitempty"`
		AncestorIDs      []primitive.ObjectID `json:"ancestor_ids,omitempty"`
		CreatedAt        string               `json:"created_at"`
		CreatedByUserID  primitive.ObjectID   `json:"created_by_user_id"`
		ModifiedAt       string               `json:"modified_at"`
		ModifiedByUserID primitive.ObjectID   `json:"modified_by_user_id"`
		Version          uint64               `json:"version"`
		State            string               `json:"state"`
		TombstoneVersion uint64               `json:"tombstone_version"`
		TombstoneExpiry  string               `json:"tombstone_expiry"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		r.logger.Error("🚨 Failed to parse response body", zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Parse timestamps
	createdAt, _ := time.Parse(time.RFC3339, apiResponse.CreatedAt)
	modifiedAt, _ := time.Parse(time.RFC3339, apiResponse.ModifiedAt)
	tombstoneExpiry, _ := time.Parse(time.RFC3339, apiResponse.TombstoneExpiry)

	// Convert API response to domain model
	response := &collectiondto.CollectionDTO{
		ID:               apiResponse.ID,
		OwnerID:          apiResponse.OwnerID,
		EncryptedName:    apiResponse.EncryptedName,
		CollectionType:   apiResponse.CollectionType,
		ParentID:         apiResponse.ParentID,
		AncestorIDs:      apiResponse.AncestorIDs,
		CreatedAt:        createdAt,
		CreatedByUserID:  apiResponse.CreatedByUserID,
		ModifiedAt:       modifiedAt,
		ModifiedByUserID: apiResponse.ModifiedByUserID,
		Version:          apiResponse.Version,
		State:            apiResponse.State,
		TombstoneVersion: apiResponse.TombstoneVersion,
		TombstoneExpiry:  tombstoneExpiry,
		Members:          make([]*collectiondto.CollectionMembershipDTO, len(apiResponse.Members)),
	}

	// Convert collection's own encrypted key if present
	if len(apiResponse.EncryptedCollectionKey.Ciphertext) > 0 {
		response.EncryptedCollectionKey = &keys.EncryptedCollectionKey{
			Ciphertext:   apiResponse.EncryptedCollectionKey.Ciphertext,
			Nonce:        apiResponse.EncryptedCollectionKey.Nonce,
			KeyVersion:   1, // Default version for existing keys
			RotatedAt:    &createdAt,
			PreviousKeys: []keys.EncryptedHistoricalKey{},
		}
	}

	// Convert members
	for i, member := range apiResponse.Members {
		memberCreatedAt, _ := time.Parse(time.RFC3339, member.CreatedAt)

		// Convert bytes to EncryptedCollectionKey struct for member
		var encryptedCollectionKey *keys.EncryptedCollectionKey
		if len(member.EncryptedCollectionKey) > 0 {
			// Create EncryptedCollectionKey from box_seal bytes
			encryptedCollectionKey = keys.NewEncryptedCollectionKeyFromBoxSeal(member.EncryptedCollectionKey)
		}

		response.Members[i] = &collectiondto.CollectionMembershipDTO{
			ID:                     member.ID,
			CollectionID:           member.CollectionID,
			RecipientID:            member.RecipientID,
			RecipientEmail:         member.RecipientEmail,
			GrantedByID:            member.GrantedByID,
			PermissionLevel:        member.PermissionLevel,
			CreatedAt:              memberCreatedAt,
			IsInherited:            member.IsInherited,
			InheritedFromID:        member.InheritedFromID,
			EncryptedCollectionKey: encryptedCollectionKey,
		}
	}

	r.logger.Debug("🔍 Parsed collection members",
		zap.Int("memberCount", len(response.Members)))

	for i, member := range response.Members {
		encryptedKeyLength := 0
		if member.EncryptedCollectionKey != nil {
			encryptedKeyLength = len(member.EncryptedCollectionKey.ToBoxSealBytes())
		}
		r.logger.Debug("🔍 Member details from API",
			zap.Int("memberIndex", i),
			zap.String("memberID", member.ID.Hex()),
			zap.String("recipientEmail", member.RecipientEmail),
			zap.Int("encryptedKeyLength", encryptedKeyLength))
	}

	r.logger.Info("✨ Successfully fetched collection from cloud server",
		zap.String("collectionID", id.Hex()))
	return response, nil
}
