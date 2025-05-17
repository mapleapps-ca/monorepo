// monorepo/native/desktop/maplefile-cli/internal/repo/remotecollection/list.go
package remotecollection

import (
	"context"

	collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
)

func (r *collectionRepository) List(ctx context.Context, filter collection.CollectionFilter) ([]*collection.RemoteCollection, error) {
	return nil, nil
}
