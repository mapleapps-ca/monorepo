// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/get.go
package collection

import (
	"context"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *localcollectionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*localcollection.LocalCollection, error) {
	// TODO: Impl.
	return nil, nil
}
