// monorepo/native/desktop/maplefile-cli/internal/domain/localcollection/model.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionRepository defines the interface for interacting with local collections
type LocalCollectionRepository interface {
	Create(ctx context.Context, collection *LocalCollection) error
	Save(ctx context.Context, collection *LocalCollection) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*LocalCollection, error)
	List(ctx context.Context, filter LocalCollectionFilter) ([]*LocalCollection, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

// LocalCollectionFilter defines filtering options for listing local collections
type LocalCollectionFilter struct {
	ParentID   *primitive.ObjectID `json:"parent_id,omitempty"`
	Type       string              `json:"type,omitempty"`
	SyncStatus *SyncStatus         `json:"sync_status,omitempty"`
}
