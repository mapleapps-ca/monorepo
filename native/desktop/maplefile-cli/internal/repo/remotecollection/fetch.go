// monorepo/native/desktop/maplefile-cli/internal/repo/remotecollection/fetch.go
package remotecollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
)

func (r *collectionRepository) Fetch(ctx context.Context, id primitive.ObjectID) (*collection.RemoteCollection, error) {
	return nil, nil
}
