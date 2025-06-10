// monorepo/native/desktop/maplefile-cli/internal/service/collection/get_filtered.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// GetFilteredInput represents the input for getting filtered collections
type GetFilteredInput struct {
	IncludeOwned  bool `json:"include_owned"`
	IncludeShared bool `json:"include_shared"`
}

// GetFilteredOutput represents the result of getting filtered collections
type GetFilteredOutput struct {
	OwnedCollections  []*collection.Collection `json:"owned_collections"`
	SharedCollections []*collection.Collection `json:"shared_collections"`
	TotalCount        int                      `json:"total_count"`
}

// GetFilteredService defines the interface for getting filtered collections
type GetFilteredService interface {
	GetFiltered(ctx context.Context, input *GetFilteredInput, userPassword string) (*GetFilteredOutput, error)
}

// getFilteredService implements the GetFilteredService interface
type getFilteredService struct {
	logger                                 *zap.Logger
	getUserByIsLoggedInUseCase             uc_user.GetByIsLoggedInUseCase
	getFilteredCollectionsFromCloudUseCase uc_collectiondto.GetFilteredCollectionsFromCloudUseCase
	collectionDecryptionService            svc_collectioncrypto.CollectionDecryptionService
}

// NewGetFilteredService creates a new service for getting filtered collections
func NewGetFilteredService(
	logger *zap.Logger,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getFilteredCollectionsFromCloudUseCase uc_collectiondto.GetFilteredCollectionsFromCloudUseCase,
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
) GetFilteredService {
	logger = logger.Named("GetFilteredService")
	return &getFilteredService{
		logger:                                 logger,
		getUserByIsLoggedInUseCase:             getUserByIsLoggedInUseCase,
		getFilteredCollectionsFromCloudUseCase: getFilteredCollectionsFromCloudUseCase,
		collectionDecryptionService:            collectionDecryptionService,
	}
}

// GetFiltered retrieves filtered collections from the cloud and decrypts them
func (s *getFilteredService) GetFiltered(ctx context.Context, input *GetFilteredInput, userPassword string) (*GetFilteredOutput, error) {
	//
	// STEP 1: Validate inputs
	//

	if input == nil {
		s.logger.Error("❌ input is required")
		return nil, errors.NewAppError("input is required", nil)
	}

	if !input.IncludeOwned && !input.IncludeShared {
		s.logger.Error("❌ at least one filter option must be enabled")
		return nil, errors.NewAppError("at least one filter option (include_owned or include_shared) must be enabled", nil)
	}

	if userPassword == "" {
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
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
	// STEP 3: Get filtered collections from cloud
	//

	request := &collectiondto.GetFilteredCollectionsRequest{
		IncludeOwned:  input.IncludeOwned,
		IncludeShared: input.IncludeShared,
	}

	s.logger.Debug("☁️ Getting filtered collections from cloud",
		zap.Bool("include_owned", input.IncludeOwned),
		zap.Bool("include_shared", input.IncludeShared))

	cloudResponse, err := s.getFilteredCollectionsFromCloudUseCase.Execute(ctx, request)
	if err != nil {
		s.logger.Error("❌ failed to get filtered collections from cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to get filtered collections from cloud", err)
	}

	//
	// STEP 4: Decrypt collections and convert to local format using crypto service
	//

	output := &GetFilteredOutput{
		OwnedCollections:  make([]*collection.Collection, 0, len(cloudResponse.OwnedCollections)),
		SharedCollections: make([]*collection.Collection, 0, len(cloudResponse.SharedCollections)),
		TotalCount:        cloudResponse.TotalCount,
	}

	// Decrypt owned collections
	for _, cloudCollection := range cloudResponse.OwnedCollections {
		localCollection, err := s.convertAndDecryptCollection(ctx, cloudCollection, userData, userPassword)
		if err != nil {
			s.logger.Warn("⚠️ failed to decrypt owned collection, skipping",
				zap.String("collection_id", cloudCollection.ID.String()),
				zap.Error(err))
			continue
		}
		output.OwnedCollections = append(output.OwnedCollections, localCollection)
	}

	// Decrypt shared collections
	for _, cloudCollection := range cloudResponse.SharedCollections {
		localCollection, err := s.convertAndDecryptCollection(ctx, cloudCollection, userData, userPassword)
		if err != nil {
			s.logger.Warn("⚠️ failed to decrypt shared collection, skipping",
				zap.String("collection_id", cloudCollection.ID.String()),
				zap.Error(err))
			continue
		}
		output.SharedCollections = append(output.SharedCollections, localCollection)
	}

	s.logger.Info("✅ Successfully retrieved and decrypted filtered collections using crypto service",
		zap.Int("owned_count", len(output.OwnedCollections)),
		zap.Int("shared_count", len(output.SharedCollections)),
		zap.Int("total_count", output.TotalCount))

	return output, nil
}

func (s *getFilteredService) convertAndDecryptCollection(ctx context.Context, cloudCollection *collectiondto.CollectionDTO, userData *dom_user.User, userPassword string) (*collection.Collection, error) {
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

	// Validate encrypted collection key
	if cloudCollection.EncryptedCollectionKey == nil {
		return nil, errors.NewAppError("collection has no encrypted key", nil)
	}

	// ✅ REPLACED: Use collection decryption service instead of manual key chain decryption
	collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, userData, localCollection, userPassword)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}
	defer crypto.ClearBytes(collectionKey)

	// ✅ REPLACED: Use collection decryption service for data decryption instead of manual implementation
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

// ❌ REMOVED: Manual crypto helper methods - replaced with crypto services
// - deriveKeyEncryptionKey() -> handled by collectionDecryptionService.ExecuteDecryptCollectionKeyChain()
// - decryptMasterKey() -> handled by collectionDecryptionService.ExecuteDecryptCollectionKeyChain()
// - decryptCollectionName() -> collectionDecryptionService.ExecuteDecryptData()
