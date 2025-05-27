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
	ParentID       *primitive.ObjectID `json:"parent_id,omitempty"`
	CollectionType string              `json:"collection_type,omitempty"`
	SyncStatus     *SyncStatus         `json:"sync_status,omitempty"`
	State          *string             `json:"state,omitempty"`           // Add state filter
	IncludeDeleted bool                `json:"include_deleted,omitempty"` // Option to include deleted items
}

// StateFilter provides helper methods for common state filtering scenarios
type StateFilter struct {
	ActiveOnly     bool `json:"active_only"`
	IncludeDeleted bool `json:"include_deleted"`
	ArchivedOnly   bool `json:"archived_only"`
}

// ToCollectionFilter converts StateFilter to CollectionFilter
func (sf StateFilter) ToCollectionFilter() CollectionFilter {
	filter := CollectionFilter{}

	if sf.ActiveOnly {
		state := CollectionStateActive
		filter.State = &state
	} else if sf.ArchivedOnly {
		state := CollectionStateArchived
		filter.State = &state
	} else if !sf.IncludeDeleted {
		// Default behavior: exclude deleted unless explicitly included
		state := CollectionStateActive
		filter.State = &state
	}

	filter.IncludeDeleted = sf.IncludeDeleted
	return filter
}

// GetActiveFilter returns a filter for active collections only
func GetActiveFilter() CollectionFilter {
	state := CollectionStateActive
	return CollectionFilter{
		State: &state,
	}
}

// GetDeletedFilter returns a filter for deleted collections only
func GetDeletedFilter() CollectionFilter {
	state := CollectionStateDeleted
	return CollectionFilter{
		State: &state,
	}
}

// GetArchivedFilter returns a filter for archived collections only
func GetArchivedFilter() CollectionFilter {
	state := CollectionStateArchived
	return CollectionFilter{
		State: &state,
	}
}

// GetAllStatesFilter returns a filter that includes all states
func GetAllStatesFilter() CollectionFilter {
	return CollectionFilter{
		IncludeDeleted: true,
	}
}
