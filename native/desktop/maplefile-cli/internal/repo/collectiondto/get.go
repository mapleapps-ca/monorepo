// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/get.go
package collectiondto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

type CollectionResponseDTO struct {
	ID                     gocql.UUID                   `json:"id"`
	OwnerID                gocql.UUID                   `json:"owner_id"`
	EncryptedName          string                       `json:"encrypted_name"`
	CollectionType         string                       `json:"collection_type"`
	ParentID               gocql.UUID                   `json:"parent_id,omitempty"`
	AncestorIDs            []gocql.UUID                 `json:"ancestor_ids,omitempty"`
	EncryptedCollectionKey *keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
	Children               []*CollectionResponseDTO     `json:"children,omitempty"`
	CreatedAt              time.Time                    `json:"created_at"`
	ModifiedAt             time.Time                    `json:"modified_at"`
	Members                []MembershipResponseDTO      `json:"members"`
}

type MembershipResponseDTO struct {
	ID             gocql.UUID `bson:"_id" json:"id"`
	CollectionID   gocql.UUID `bson:"collection_id" json:"collection_id"`     // ID of the collection (redundant but helpful for queries)
	RecipientID    gocql.UUID `bson:"recipient_id" json:"recipient_id"`       // User receiving access
	RecipientEmail string     `bson:"recipient_email" json:"recipient_email"` // Email for display purposes
	GrantedByID    gocql.UUID `bson:"granted_by_id" json:"granted_by_id"`     // User who shared the collection

	// Collection key encrypted with recipient's public key using box_seal. This matches the box_seal format which doesn't need a separate nonce.
	EncryptedCollectionKey []byte `bson:"encrypted_collection_key" json:"encrypted_collection_key"`

	// Access details
	PermissionLevel string    `bson:"permission_level" json:"permission_level"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`

	// Sharing origin tracking
	IsInherited     bool       `bson:"is_inherited" json:"is_inherited"`                               // Tracks whether access was granted directly or inherited from a parent
	InheritedFromID gocql.UUID `bson:"inherited_from_id,omitempty" json:"inherited_from_id,omitempty"` // InheritedFromID identifies which parent collection granted this access
}

func (r *collectionDTORepository) GetFromCloudByID(ctx context.Context, id gocql.UUID) (*collectiondto.CollectionDTO, error) {
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
	var apiResponse *CollectionResponseDTO

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		r.logger.Error("🚨 Failed to parse response body", zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Parse timestamps
	// createdAt, _ := time.Parse(time.RFC3339, apiResponse.CreatedAt)
	// modifiedAt, _ := time.Parse(time.RFC3339, apiResponse.ModifiedAt)
	// tombstoneExpiry, _ := time.Parse(time.RFC3339, apiResponse.TombstoneExpiry)

	// Convert API response to domain model
	response := &collectiondto.CollectionDTO{
		ID:             apiResponse.ID,
		OwnerID:        apiResponse.OwnerID,
		EncryptedName:  apiResponse.EncryptedName,
		CollectionType: apiResponse.CollectionType,
		ParentID:       apiResponse.ParentID,
		AncestorIDs:    apiResponse.AncestorIDs,
		// CreatedAt:        createdAt,
		// CreatedByUserID:  apiResponse.CreatedByUserID,
		// ModifiedAt:       modifiedAt,
		// ModifiedByUserID: apiResponse.ModifiedByUserID,
		// Version:          apiResponse.Version,
		// State:            apiResponse.State,
		// TombstoneVersion: apiResponse.TombstoneVersion,
		// TombstoneExpiry:  tombstoneExpiry,
		Members: make([]*collectiondto.CollectionMembershipDTO, len(apiResponse.Members)),
	}

	// Convert collection's own encrypted key if present
	if len(apiResponse.EncryptedCollectionKey.Ciphertext) > 0 {
		response.EncryptedCollectionKey = &keys.EncryptedCollectionKey{
			Ciphertext: apiResponse.EncryptedCollectionKey.Ciphertext,
			Nonce:      apiResponse.EncryptedCollectionKey.Nonce,
			KeyVersion: 1, // Default version for existing keys
			// RotatedAt:    &createdAt,
			PreviousKeys: []keys.EncryptedHistoricalKey{},
		}
	}

	r.logger.Debug("🔎 Looking at API response",
		zap.Any("apiResponse", apiResponse),
	)

	// Convert members
	for i, member := range apiResponse.Members {
		// memberCreatedAt, _ := time.Parse(time.RFC3339, member.CreatedAt)

		r.logger.Debug("🔎 Looking at member",
			zap.Any("member", member),
		)

		// Convert bytes to EncryptedCollectionKey struct for member
		var encryptedCollectionKey *keys.EncryptedCollectionKey
		if len(member.EncryptedCollectionKey) > 0 {
			// Create EncryptedCollectionKey from box_seal bytes
			encryptedCollectionKey = keys.NewEncryptedCollectionKeyFromBoxSeal(member.EncryptedCollectionKey)
		}

		response.Members[i] = &collectiondto.CollectionMembershipDTO{
			ID:              member.ID,
			CollectionID:    member.CollectionID,
			RecipientID:     member.RecipientID,
			RecipientEmail:  member.RecipientEmail,
			GrantedByID:     member.GrantedByID,
			PermissionLevel: member.PermissionLevel,
			// CreatedAt:              memberCreatedAt,
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
