// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/upload.go
package collectiondto

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
)

func (s *collectionDTORepository) Create(ctx context.Context, collection *collectiondto.CollectionDTO) (*primitive.ObjectID, error) {
	// Stub implementation: Always return a zero ObjectID and nil error.
	// In a real stub or mock, you would implement test-specific logic.
	var zeroID primitive.ObjectID
	return &zeroID, nil
}
