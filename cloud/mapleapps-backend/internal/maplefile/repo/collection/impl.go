// cloud/mapleapps-backend/internal/maplefile/repo/collection/impl.go
package collection

import (
	"encoding/json"

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

// Helper functions for JSON serialization
func (impl *collectionRepositoryImpl) serializeMembers(members []dom_collection.CollectionMembership) (string, error) {
	if len(members) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(members)
	return string(data), err
}

func (impl *collectionRepositoryImpl) deserializeMembers(data string) ([]dom_collection.CollectionMembership, error) {
	if data == "" || data == "[]" {
		return []dom_collection.CollectionMembership{}, nil
	}
	var members []dom_collection.CollectionMembership
	err := json.Unmarshal([]byte(data), &members)
	return members, err
}

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

// calculateDepth calculates the depth of a collection in hierarchy
func (impl *collectionRepositoryImpl) calculateDepth(ancestorIDs []gocql.UUID) int {
	return len(ancestorIDs) + 1 // +1 because depth starts at 1 for direct children
}

// isValidUUID checks if UUID is not nil/empty
func (impl *collectionRepositoryImpl) isValidUUID(id gocql.UUID) bool {
	return id.String() != "00000000-0000-0000-0000-000000000000"
}

// buildAncestorDepthEntries builds entries for ancestor depth table
func (impl *collectionRepositoryImpl) buildAncestorDepthEntries(collectionID gocql.UUID, ancestorIDs []gocql.UUID) []ancestorDepthEntry {
	var entries []ancestorDepthEntry

	for i, ancestorID := range ancestorIDs {
		depth := len(ancestorIDs) - i // Reverse depth calculation
		entries = append(entries, ancestorDepthEntry{
			AncestorID:   ancestorID,
			Depth:        depth,
			CollectionID: collectionID,
		})
	}

	return entries
}

type ancestorDepthEntry struct {
	AncestorID   gocql.UUID
	Depth        int
	CollectionID gocql.UUID
}
