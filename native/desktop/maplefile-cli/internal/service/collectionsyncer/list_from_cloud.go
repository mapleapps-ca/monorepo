// monorepo/native/desktop/maplefile-cli/internal/service/collectionsyncer/list_from_cloud.go
package collectionsyncer

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// ListFromCloudInput represents the input for listing collections from cloud
type ListFromCloudInput struct {
	ParentID       *gocql.UUID `json:"parent_id,omitempty"`
	CollectionType string      `json:"collection_type,omitempty"`
}

// ListFromCloudOutput represents the result of listing collections from cloud
type ListFromCloudOutput struct {
	Collections []*collection.Collection `json:"collections"`
	Count       int                      `json:"count"`
}

// ListFromCloudService defines the interface for listing collections from cloud
type ListFromCloudService interface {
	ListFromCloud(ctx context.Context, input *ListFromCloudInput, userPassword string) (*ListFromCloudOutput, error)
}

// listFromCloudService implements the ListFromCloudService interface
type listFromCloudService struct {
	logger                          *zap.Logger
	getUserByIsLoggedInUseCase      uc_user.GetByIsLoggedInUseCase
	listCollectionsFromCloudUseCase uc_collectiondto.ListCollectionsFromCloudUseCase
	collectionDecryptionService     svc_collectioncrypto.CollectionDecryptionService
}

// NewListFromCloudService creates a new service for listing collections from cloud
func NewListFromCloudService(
	logger *zap.Logger,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	listCollectionsFromCloudUseCase uc_collectiondto.ListCollectionsFromCloudUseCase,
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
) ListFromCloudService {
	logger = logger.Named("ListFromCloudService")
	return &listFromCloudService{
		logger:                          logger,
		getUserByIsLoggedInUseCase:      getUserByIsLoggedInUseCase,
		listCollectionsFromCloudUseCase: listCollectionsFromCloudUseCase,
		collectionDecryptionService:     collectionDecryptionService,
	}
}

// ListFromCloud retrieves collections from the cloud and decrypts them
func (s *listFromCloudService) ListFromCloud(ctx context.Context, input *ListFromCloudInput, userPassword string) (*ListFromCloudOutput, error) {
	//
	// STEP 1: Validate inputs
	//

	if input == nil {
		s.logger.Error("❌ input is required")
		return nil, errors.NewAppError("input is required", nil)
	}

	if userPassword == "" {
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
	}

	// Validate collection type if provided
	if input.CollectionType != "" {
		validTypes := map[string]bool{
			collection.CollectionTypeFolder: true,
			collection.CollectionTypeAlbum:  true,
		}
		if !validTypes[input.CollectionType] {
			s.logger.Error("❌ invalid collection type", zap.String("type", input.CollectionType))
			return nil, errors.NewAppError("collection type must be either 'folder' or 'album'", nil)
		}
	}

	//
	// STEP 2: Get user data for decryption
	//

	userData, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("❌ failed to get authenticated user", zap.Error(err))
		return nil, errors.NewAppError("failed to get user data", err)
	}

	if userData == nil {
		s.logger.Error("❌ authenticated user not found")
		return nil, errors.NewAppError("authenticated user not found; please login first", nil)
	}

	//
	// STEP 3: Get collections from cloud
	//

	filter := collectiondto.CollectionFilter{
		ParentID:       input.ParentID,
		CollectionType: input.CollectionType,
	}

	s.logger.Debug("☁️ Getting collections from cloud",
		zap.Any("filter", filter))

	cloudCollections, err := s.listCollectionsFromCloudUseCase.Execute(ctx, filter)
	if err != nil {
		s.logger.Error("❌ failed to get collections from cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to get collections from cloud", err)
	}

	//
	// STEP 4: Decrypt collections and convert to local format using crypto service
	//

	output := &ListFromCloudOutput{
		Collections: make([]*collection.Collection, 0, len(cloudCollections)),
		Count:       0,
	}

	// Decrypt collections
	for _, cloudCollection := range cloudCollections {
		localCollection, err := s.convertAndDecryptCollection(ctx, cloudCollection, userData, userPassword)
		if err != nil {
			s.logger.Warn("⚠️ failed to decrypt collection, skipping",
				zap.String("collection_id", cloudCollection.ID.String()),
				zap.Error(err))
			continue
		}
		output.Collections = append(output.Collections, localCollection)
	}

	output.Count = len(output.Collections)

	s.logger.Info("✅ Successfully retrieved and decrypted collections from cloud using crypto service",
		zap.Int("count", output.Count),
		zap.Int("total_cloud_collections", len(cloudCollections)))

	return output, nil
}

func (s *listFromCloudService) convertAndDecryptCollection(ctx context.Context, cloudCollection *collectiondto.CollectionDTO, userData *dom_user.User, userPassword string) (*collection.Collection, error) {
	// Convert to local collection format first (for crypto service compatibility)
	localCollection := &collection.Collection{
		ID:                     cloudCollection.ID,
		OwnerID:                cloudCollection.OwnerID,
		EncryptedName:          cloudCollection.EncryptedName,
		CollectionType:         cloudCollection.CollectionType,
		EncryptedCollectionKey: cloudCollection.EncryptedCollectionKey,
		ParentID:               cloudCollection.ParentID,
		AncestorIDs:            cloudCollection.AncestorIDs,
		Children:               make([]*collection.Collection, 0),
		CreatedAt:              cloudCollection.CreatedAt,
		CreatedByUserID:        cloudCollection.CreatedByUserID,
		ModifiedAt:             cloudCollection.ModifiedAt,
		ModifiedByUserID:       cloudCollection.ModifiedByUserID,
		Version:                cloudCollection.Version,
		SyncStatus:             collection.SyncStatusSynced, // Assume synced since it came from cloud
	}

	// Set state from DTO, defaulting to active if not set
	if cloudCollection.State != "" {
		localCollection.State = cloudCollection.State
	} else {
		localCollection.State = collection.CollectionStateActive
	}

	// Validate encrypted collection key
	if cloudCollection.EncryptedCollectionKey == nil {
		return nil, errors.NewAppError("collection has no encrypted key", nil)
	}

	// Use collection decryption service for key chain decryption
	collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, userData, localCollection, userPassword)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}
	defer crypto.ClearBytes(collectionKey)

	// Use collection decryption service for data decryption
	decryptedName, err := s.collectionDecryptionService.ExecuteDecryptData(ctx, cloudCollection.EncryptedName, collectionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection name", err)
	}

	// Set decrypted data
	localCollection.Name = decryptedName

	// Convert members if any
	if cloudCollection.Members != nil {
		localCollection.Members = make([]*collection.CollectionMembership, len(cloudCollection.Members))
		for i, cloudMember := range cloudCollection.Members {
			localCollection.Members[i] = &collection.CollectionMembership{
				ID:                     cloudMember.ID,
				CollectionID:           cloudMember.CollectionID,
				RecipientID:            cloudMember.RecipientID,
				RecipientEmail:         cloudMember.RecipientEmail,
				GrantedByID:            cloudMember.GrantedByID,
				EncryptedCollectionKey: cloudMember.EncryptedCollectionKey,
				PermissionLevel:        cloudMember.PermissionLevel,
				CreatedAt:              cloudMember.CreatedAt,
				IsInherited:            cloudMember.IsInherited,
				InheritedFromID:        cloudMember.InheritedFromID,
			}
		}
	} else {
		localCollection.Members = make([]*collection.CollectionMembership, 0)
	}

	return localCollection, nil
}
