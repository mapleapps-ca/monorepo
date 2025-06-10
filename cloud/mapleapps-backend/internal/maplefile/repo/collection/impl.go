// cloud/mapleapps-backend/internal/maplefile/repo/collection/impl.go
package collection

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/keys"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

type collectionRepositoryImpl struct {
	Logger  *zap.Logger
	Session *gocql.Session
}

func NewRepository(appCfg *config.Configuration, session *gocql.Session, loggerp *zap.Logger) dom_collection.CollectionRepository {
	loggerp = loggerp.Named("CollectionRepository")

	return &collectionRepositoryImpl{
		Logger:  loggerp,
		Session: session,
	}
}

// Helper functions for simplified JSON serialization (only ancestor_ids now)
func (impl *collectionRepositoryImpl) serializeAncestorIDs(ancestorIDs []gocql.UUID) (string, error) {
	if len(ancestorIDs) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(ancestorIDs)
	return string(data), err
}

func (impl *collectionRepositoryImpl) deserializeAncestorIDs(data string) ([]gocql.UUID, error) {
	if data == "" || data == "[]" {
		return []gocql.UUID{}, nil
	}
	var ancestorIDs []gocql.UUID
	err := json.Unmarshal([]byte(data), &ancestorIDs)
	return ancestorIDs, err
}

func (impl *collectionRepositoryImpl) serializeEncryptedCollectionKey(key *keys.EncryptedCollectionKey) (string, error) {
	if key == nil {
		return "", nil
	}
	data, err := json.Marshal(key)
	return string(data), err
}

func (impl *collectionRepositoryImpl) deserializeEncryptedCollectionKey(data string) (*keys.EncryptedCollectionKey, error) {
	if data == "" {
		return nil, nil
	}
	var key keys.EncryptedCollectionKey
	err := json.Unmarshal([]byte(data), &key)
	return &key, err
}

// isValidUUID checks if UUID is not nil/empty
func (impl *collectionRepositoryImpl) isValidUUID(id gocql.UUID) bool {
	return id.String() != "00000000-0000-0000-0000-000000000000"
}

// NEW: Core helper methods for loading collections with members
func (impl *collectionRepositoryImpl) loadCollectionWithMembers(ctx context.Context, collectionID gocql.UUID, stateAware bool) (*dom_collection.Collection, error) {
	// 1. Load base collection
	collection, err := impl.getBaseCollection(ctx, collectionID, stateAware)
	if err != nil || collection == nil {
		return collection, err
	}

	// 2. Load and populate members
	members, err := impl.getCollectionMembers(ctx, collectionID)
	if err != nil {
		return nil, err
	}
	collection.Members = members

	return collection, nil
}

func (impl *collectionRepositoryImpl) getBaseCollection(ctx context.Context, id gocql.UUID, stateAware bool) (*dom_collection.Collection, error) {
	var (
		encryptedName, collectionType, encryptedKeyJSON      string
		ancestorIDsJSON                                      string
		parentID, ownerID, createdByUserID, modifiedByUserID gocql.UUID
		createdAt, modifiedAt, tombstoneExpiry               time.Time
		version, tombstoneVersion                            uint64
		state                                                string
	)

	query := `SELECT id, owner_id, encrypted_name, collection_type, encrypted_collection_key,
		parent_id, ancestor_ids, created_at, created_by_user_id, modified_at,
		modified_by_user_id, version, state, tombstone_version, tombstone_expiry
		FROM maplefile_collections_by_id WHERE id = ?`

	err := impl.Session.Query(query, id).WithContext(ctx).Scan(
		&id, &ownerID, &encryptedName, &collectionType, &encryptedKeyJSON,
		&parentID, &ancestorIDsJSON, &createdAt, &createdByUserID,
		&modifiedAt, &modifiedByUserID, &version, &state, &tombstoneVersion, &tombstoneExpiry)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Apply state filtering if state-aware mode is enabled
	if stateAware && state != dom_collection.CollectionStateActive {
		return nil, nil
	}

	// Deserialize complex fields
	ancestorIDs, err := impl.deserializeAncestorIDs(ancestorIDsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize ancestor IDs: %w", err)
	}

	encryptedKey, err := impl.deserializeEncryptedCollectionKey(encryptedKeyJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize encrypted collection key: %w", err)
	}

	collection := &dom_collection.Collection{
		ID:                     id,
		OwnerID:                ownerID,
		EncryptedName:          encryptedName,
		CollectionType:         collectionType,
		EncryptedCollectionKey: encryptedKey,
		Members:                []dom_collection.CollectionMembership{}, // Will be populated separately
		ParentID:               parentID,
		AncestorIDs:            ancestorIDs,
		CreatedAt:              createdAt,
		CreatedByUserID:        createdByUserID,
		ModifiedAt:             modifiedAt,
		ModifiedByUserID:       modifiedByUserID,
		Version:                version,
		State:                  state,
		TombstoneVersion:       tombstoneVersion,
		TombstoneExpiry:        tombstoneExpiry,
	}

	return collection, nil
}

func (impl *collectionRepositoryImpl) getCollectionMembers(ctx context.Context, collectionID gocql.UUID) ([]dom_collection.CollectionMembership, error) {
	var members []dom_collection.CollectionMembership

	query := `SELECT recipient_id, recipient_email, granted_by_id,
		encrypted_collection_key, permission_level, created_at,
		is_inherited, inherited_from_id
		FROM maplefile_collection_members WHERE collection_id = ?`

	iter := impl.Session.Query(query, collectionID).WithContext(ctx).Iter()

	var (
		recipientID, grantedByID, inheritedFromID gocql.UUID
		recipientEmail, permissionLevel           string
		encryptedCollectionKey                    []byte
		createdAt                                 time.Time
		isInherited                               bool
	)

	for iter.Scan(&recipientID, &recipientEmail, &grantedByID,
		&encryptedCollectionKey, &permissionLevel, &createdAt,
		&isInherited, &inheritedFromID) {

		member := dom_collection.CollectionMembership{
			ID:                     gocql.TimeUUID(),
			CollectionID:           collectionID,
			RecipientID:            recipientID,
			RecipientEmail:         recipientEmail,
			GrantedByID:            grantedByID,
			EncryptedCollectionKey: encryptedCollectionKey,
			PermissionLevel:        permissionLevel,
			CreatedAt:              createdAt,
			IsInherited:            isInherited,
			InheritedFromID:        inheritedFromID,
		}
		members = append(members, member)
	}

	return members, iter.Close()
}

func (impl *collectionRepositoryImpl) loadMultipleCollectionsWithMembers(ctx context.Context, collectionIDs []gocql.UUID) ([]*dom_collection.Collection, error) {
	if len(collectionIDs) == 0 {
		return []*dom_collection.Collection{}, nil
	}

	var collections []*dom_collection.Collection
	for _, id := range collectionIDs {
		collection, err := impl.loadCollectionWithMembers(ctx, id, true)
		if err != nil {
			impl.Logger.Warn("failed to load collection",
				zap.String("collection_id", id.String()),
				zap.Error(err))
			continue
		}
		if collection != nil {
			collections = append(collections, collection)
		}
	}

	return collections, nil
}

// Permission helper
func (impl *collectionRepositoryImpl) hasPermission(userPermission, requiredPermission string) bool {
	permissionLevels := map[string]int{
		dom_collection.CollectionPermissionReadOnly:  1,
		dom_collection.CollectionPermissionReadWrite: 2,
		dom_collection.CollectionPermissionAdmin:     3,
	}

	userLevel, userExists := permissionLevels[userPermission]
	requiredLevel, requiredExists := permissionLevels[requiredPermission]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}
