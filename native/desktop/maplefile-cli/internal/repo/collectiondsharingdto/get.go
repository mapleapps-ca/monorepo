// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondsharingdto/get.go
package collectiondsharingdto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"go.uber.org/zap"
)

// GetCollectionWithMembersFromCloud retrieves collection information including member list
func (r *collectionSharingDTORepository) GetCollectionWithMembersFromCloud(ctx context.Context, collectionID gocql.UUID) (*collectiondto.CollectionDTO, error) {
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

	// Create HTTP request - this uses the existing get collection endpoint which includes members
	getURL := fmt.Sprintf("%s/maplefile/api/v1/collections/%s", serverURL, collectionID.Hex())
	req, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
	if err != nil {
		r.logger.Error("❌ Failed to create HTTP request", zap.String("url", getURL), zap.Error(err))
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error("❌ Failed to execute get collection request", zap.Error(err))
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

	// Parse response - this will be a CollectionDTO format from the API
	var collectionResponse struct {
		ID             gocql.UUID `json:"id"`
		OwnerID        gocql.UUID `json:"owner_id"`
		EncryptedName  string     `json:"encrypted_name"`
		CollectionType string     `json:"collection_type"`
		Members        []struct {
			ID              gocql.UUID `json:"id"`
			RecipientID     gocql.UUID `json:"recipient_id"`
			RecipientEmail  string     `json:"recipient_email"`
			PermissionLevel string     `json:"permission_level"`
			GrantedByID     gocql.UUID `json:"granted_by_id"`
			IsInherited     bool       `json:"is_inherited"`
			InheritedFromID gocql.UUID `json:"inherited_from_id,omitempty"`
			CreatedAt       string     `json:"created_at"`
			// Handle encrypted collection key as bytes from API
			EncryptedCollectionKey []byte `json:"encrypted_collection_key"`
		} `json:"members"`
		CreatedAt  string `json:"created_at"`
		ModifiedAt string `json:"modified_at"`
	}

	if err := json.Unmarshal(body, &collectionResponse); err != nil {
		r.logger.Error("❌ Failed to parse collection response", zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Parse timestamps
	createdAt, _ := time.Parse(time.RFC3339, collectionResponse.CreatedAt)
	modifiedAt, _ := time.Parse(time.RFC3339, collectionResponse.ModifiedAt)

	// Convert to domain model
	domainCollection := &collectiondto.CollectionDTO{
		ID:             collectionResponse.ID,
		OwnerID:        collectionResponse.OwnerID,
		EncryptedName:  collectionResponse.EncryptedName,
		CollectionType: collectionResponse.CollectionType,
		CreatedAt:      createdAt,
		ModifiedAt:     modifiedAt,
		Members:        make([]*collectiondto.CollectionMembershipDTO, len(collectionResponse.Members)),
	}

	// Convert members
	for i, member := range collectionResponse.Members {
		memberCreatedAt, _ := time.Parse(time.RFC3339, member.CreatedAt)

		// Convert bytes to EncryptedCollectionKey struct
		var encryptedCollectionKey *keys.EncryptedCollectionKey
		if len(member.EncryptedCollectionKey) > 0 {
			// Create EncryptedCollectionKey from box_seal bytes
			encryptedCollectionKey = keys.NewEncryptedCollectionKeyFromBoxSeal(member.EncryptedCollectionKey)
		}

		domainCollection.Members[i] = &collectiondto.CollectionMembershipDTO{
			ID:                     member.ID,
			CollectionID:           collectionResponse.ID,
			RecipientID:            member.RecipientID,
			RecipientEmail:         member.RecipientEmail,
			GrantedByID:            member.GrantedByID,
			PermissionLevel:        member.PermissionLevel,
			IsInherited:            member.IsInherited,
			InheritedFromID:        member.InheritedFromID,
			CreatedAt:              memberCreatedAt,
			EncryptedCollectionKey: encryptedCollectionKey,
		}
	}

	r.logger.Info("✅ Successfully retrieved collection with members",
		zap.String("collectionID", collectionID.Hex()),
		zap.Int("memberCount", len(domainCollection.Members)))

	return domainCollection, nil
}
