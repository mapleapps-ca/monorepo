// monorepo/native/desktop/maplefile-cli/internal/domain/collection/model.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionRepository defines the interface for interacting with local collections
type CollectionRepository interface {
	Create(ctx context.Context, collection *Collection) error
	Save(ctx context.Context, collection *Collection) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Collection, error)
	List(ctx context.Context, filter CollectionFilter) ([]*Collection, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	OpenTransaction() error
	CommitTransaction() error
	DiscardTransaction()
}

// CollectionFilter defines filtering options for listing local collections
type CollectionFilter struct {
	ParentID   *primitive.ObjectID `json:"parent_id,omitempty"`
	Type       string              `json:"type,omitempty"`
	SyncStatus *SyncStatus         `json:"sync_status,omitempty"`
}
