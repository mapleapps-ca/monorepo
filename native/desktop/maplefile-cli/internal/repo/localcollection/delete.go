// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/delete.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *localcollectionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	// TODO: Impl.
	return nil
}
