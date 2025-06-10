// monorepo/native/desktop/maplefile-cli/internal/repo/collection/get.go
package collection

import (
	"context"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

func (r *collectionRepository) GetByID(ctx context.Context, id gocql.UUID) (*collection.Collection, error) {
	// Generate key for this collection
	key := r.generateKey(id.String())

	// Get from database
	collBytes, err := r.dbClient.Get(key)
	if err != nil {
		r.logger.Error("Failed to retrieve collection from local storage",
			zap.String("key", key),
			zap.Error(err))
		return nil, errors.NewAppError("failed to retrieve collection from local storage", err)
	}

	// Check if collection was found
	if collBytes == nil {
		return nil, nil
	}

	// Deserialize the collection
	collection, err := collection.NewFromDeserialized(collBytes)
	if err != nil {
		r.logger.Error("Failed to deserialize collection", zap.Error(err))
		return nil, errors.NewAppError("failed to deserialize collection", err)
	}

	return collection, nil
}
