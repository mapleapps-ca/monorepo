// cloud/backend/internal/maplefile/repo/collection/delete.go
package collection

import (
	"context"

	"github.com/gocql/gocql"
)

func (impl collectionRepositoryImpl) Archive(ctx context.Context, id gocql.UUID) error {
	//TODO: Impl.
	return nil
}
