// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/list.go
package collection

import (
	"context"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

func (r *localcollectionRepository) List(ctx context.Context, filter localcollection.LocalCollectionFilter) ([]*localcollection.LocalCollection, error) {
	// TODO: Impl.
	return nil, nil
}
