// monorepo/cloud/backend/internal/maplefile/domain/collection/filter.go
package collection

import "github.com/gocql/gocql"

// CollectionFilterOptions defines the filtering options for retrieving collections
type CollectionFilterOptions struct {
	// IncludeOwned includes collections where the user is the owner
	IncludeOwned bool `json:"include_owned"`
	// IncludeShared includes collections where the user is a member (shared with them)
	IncludeShared bool `json:"include_shared"`
	// UserID is the user for whom we're filtering collections
	UserID gocql.UUID `json:"user_id"`
}

// CollectionFilterResult represents the result of a filtered collection query
type CollectionFilterResult struct {
	// OwnedCollections are collections where the user is the owner
	OwnedCollections []*Collection `json:"owned_collections"`
	// SharedCollections are collections shared with the user
	SharedCollections []*Collection `json:"shared_collections"`
	// TotalCount is the total number of collections returned
	TotalCount int `json:"total_count"`
}

// GetAllCollections returns all collections (owned + shared) in a single slice
func (r *CollectionFilterResult) GetAllCollections() []*Collection {
	allCollections := make([]*Collection, 0, len(r.OwnedCollections)+len(r.SharedCollections))
	allCollections = append(allCollections, r.OwnedCollections...)
	allCollections = append(allCollections, r.SharedCollections...)
	return allCollections
}

// IsValid checks if the filter options are valid
func (options *CollectionFilterOptions) IsValid() bool {
	// At least one filter option must be enabled
	return options.IncludeOwned || options.IncludeShared
}

// ShouldIncludeAll returns true if both owned and shared collections should be included
func (options *CollectionFilterOptions) ShouldIncludeAll() bool {
	return options.IncludeOwned && options.IncludeShared
}
