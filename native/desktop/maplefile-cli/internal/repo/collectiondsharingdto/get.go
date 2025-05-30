// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondsharingdto/get.go
package collectiondsharingdto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// GetCollectionWithMembersFromCloud retrieves collection information including member list
func (r *collectionSharingDTORepository) GetCollectionWithMembersFromCloud(ctx context.Context, collectionID primitive.ObjectID) (*collectiondto.CollectionDTO, error) {
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
	var collectionDTO struct {
		ID             primitive.ObjectID `json:"id"`
		OwnerID        primitive.ObjectID `json:"owner_id"`
		EncryptedName  string             `json:"encrypted_name"`
		CollectionType string             `json:"collection_type"`
		Members        []struct {
			ID              primitive.ObjectID `json:"id"`
			RecipientID     primitive.ObjectID `json:"recipient_id"`
			RecipientEmail  string             `json:"recipient_email"`
			PermissionLevel string             `json:"permission_level"`
			GrantedByID     primitive.ObjectID `json:"granted_by_id"`
			IsInherited     bool               `json:"is_inherited"`
			CreatedAt       string             `json:"created_at"`
		} `json:"members"`
		CreatedAt  string `json:"created_at"`
		ModifiedAt string `json:"modified_at"`
	}

	if err := json.Unmarshal(body, &collectionDTO); err != nil {
		r.logger.Error("❌ Failed to parse collection response", zap.Error(err))
		return nil, errors.NewAppError("failed to parse response", err)
	}

	// Convert to domain model
	domainCollection := &collectiondto.CollectionDTO{
		ID:             collectionDTO.ID,
		OwnerID:        collectionDTO.OwnerID,
		EncryptedName:  collectionDTO.EncryptedName,
		CollectionType: collectionDTO.CollectionType,
		Members:        make([]*collectiondto.CollectionMembershipDTO, len(collectionDTO.Members)),
	}

	// Convert members
	for i, member := range collectionDTO.Members {
		domainCollection.Members[i] = &collectiondto.CollectionMembershipDTO{
			ID:              member.ID,
			CollectionID:    collectionDTO.ID,
			RecipientID:     member.RecipientID,
			RecipientEmail:  member.RecipientEmail,
			GrantedByID:     member.GrantedByID,
			PermissionLevel: member.PermissionLevel,
			IsInherited:     member.IsInherited,
		}
	}

	r.logger.Info("✅ Successfully retrieved collection with members",
		zap.String("collectionID", collectionID.Hex()),
		zap.Int("memberCount", len(domainCollection.Members)))

	return domainCollection, nil
}
