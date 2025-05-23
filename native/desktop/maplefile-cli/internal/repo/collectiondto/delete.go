// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/delete.go
package collectiondto

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *collectionDTORepository) DeleteInCloudByID(ctx context.Context, id primitive.ObjectID) error {
	// Stub implementation: Always return nil error.
	// In a real stub or mock, you would simulate specific errors like ErrNotFound.
	return nil
}
