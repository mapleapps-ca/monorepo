// monorepo/native/desktop/maplefile-cli/internal/repo/collection/delete.go
package collection

import (
	"context"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

func (r *collectionRepository) Delete(ctx context.Context, id gocql.UUID) error {
	// Generate key for this collection
	key := r.generateKey(id.String())

	// Delete from database
	if err := r.dbClient.Delete(key); err != nil {
		r.logger.Error("Failed to delete collection from local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to delete collection from local storage", err)
	}

	return nil
}
