// cloud/backend/internal/maplefile/repo/collection/get.go
package collection

import (
	"context"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) Get(ctx context.Context, id gocql.UUID) (*dom_collection.Collection, error) {
	return nil, nil
}

// Add method to get collection regardless of state
func (impl collectionRepositoryImpl) GetWithAnyState(ctx context.Context, id gocql.UUID) (*dom_collection.Collection, error) {
	return nil, nil
}

func (impl collectionRepositoryImpl) GetAllByUserID(ctx context.Context, ownerID gocql.UUID) ([]*dom_collection.Collection, error) {
	return nil, nil
}

func (impl collectionRepositoryImpl) GetCollectionsSharedWithUser(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error) {
	return nil, nil
}

func (impl collectionRepositoryImpl) FindByParent(ctx context.Context, parentID gocql.UUID) ([]*dom_collection.Collection, error) {
	return nil, nil
}

func (impl collectionRepositoryImpl) FindRootCollections(ctx context.Context, ownerID gocql.UUID) ([]*dom_collection.Collection, error) {
	return nil, nil
}

func (impl collectionRepositoryImpl) FindDescendants(ctx context.Context, collectionID gocql.UUID) ([]*dom_collection.Collection, error) {
	return nil, nil
}

func (impl collectionRepositoryImpl) GetFullHierarchy(ctx context.Context, rootID gocql.UUID) (*dom_collection.Collection, error) {
	return nil, nil
}
