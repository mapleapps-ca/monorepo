// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/list.go
package collectiondto

import (
	"context"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
)

// List lists CollectionDTOs from the cloud service based on the provided filter criteria.
// An empty filter should return all accessible CollectionDTOs.
// It returns a slice of matching CollectionDTOs or an error if the listing fails.
// The returned slice is guaranteed to be non-nil, even if no collections match the filter (it will be an empty slice).
func (s *collectionDTORepository) List(ctx context.Context, filter collectiondto.CollectionFilter) ([]*collectiondto.CollectionDTO, error) {
	// Stub implementation: Always return an empty slice and nil error.
	// In a real stub or mock, you would return a predefined list of mock objects or simulate errors.
	return []*collectiondto.CollectionDTO{}, nil
}
