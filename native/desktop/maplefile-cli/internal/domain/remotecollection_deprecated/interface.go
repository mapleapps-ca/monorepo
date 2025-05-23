// monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection/interface.go
package remotecollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RemoteCollectionRepository defines the interface for interacting with collections on the cloud cloud backend.
type RemoteCollectionRepository interface {
	Create(ctx context.Context, request *RemoteCreateCollectionRequest) (*RemoteCollectionResponse, error)
	Fetch(ctx context.Context, id primitive.ObjectID) (*RemoteCollection, error)
	List(ctx context.Context, filter CollectionFilter) ([]*RemoteCollection, error)
}

// CollectionFilter defines filtering options for listing collections
type CollectionFilter struct {
	ParentID   *primitive.ObjectID `json:"parent_id,omitempty"`
	Type       string              `json:"type,omitempty"`
	SyncStatus *SyncStatus         `json:"sync_status,omitempty"`
}
