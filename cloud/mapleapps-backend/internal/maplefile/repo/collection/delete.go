// cloud/backend/internal/maplefile/repo/collection/delete.go
package collection

import (
	"context"

	"github.com/gocql/gocql"
)

func (impl collectionRepositoryImpl) SoftDelete(ctx context.Context, id gocql.UUID) error {

	return nil
}

func (impl collectionRepositoryImpl) HardDelete(ctx context.Context, id gocql.UUID) error {

	return nil
}
