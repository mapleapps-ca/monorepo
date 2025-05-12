// github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/repo/federateduser/list.go
package federateduser

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	dom_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/federateduser"
)

// buildCountMatchStage creates the match stage for the aggregation pipeline
func (s *userStorerImpl) buildCountMatchStage(filter *dom_user.FederatedUserFilter) bson.M {
	match := bson.M{}

	if filter.Status != 0 {
		match["status"] = filter.Status
	}

	if filter.Role != 0 {
		match["role"] = filter.Role
	}

	// Date range filtering
	if filter.CreatedAtStart != nil || filter.CreatedAtEnd != nil {
		createdAtFilter := bson.M{}
		if filter.CreatedAtStart != nil {
			createdAtFilter["$gte"] = filter.CreatedAtStart
		}
		if filter.CreatedAtEnd != nil {
			createdAtFilter["$lte"] = filter.CreatedAtEnd
		}
		match["created_at"] = createdAtFilter
	}

	// Text search for name or email
	if filter.Name != nil && *filter.Name != "" {
		match["$or"] = []bson.M{
			{"name": bson.M{"$regex": *filter.Name, "$options": "i"}},
			{"lexical_name": bson.M{"$regex": *filter.Name, "$options": "i"}},
		}
	}

	if filter.Email != nil && *filter.Email != "" {
		match["email"] = bson.M{"$regex": *filter.Email, "$options": "i"}
	}

	// More comprehensive search across multiple fields
	if filter.SearchTerm != nil && *filter.SearchTerm != "" {
		searchRegex := bson.M{"$regex": *filter.SearchTerm, "$options": "i"}
		match["$or"] = []bson.M{
			{"name": searchRegex},
			{"lexical_name": searchRegex},
			{"email": searchRegex},
			{"phone": searchRegex},
			{"country": searchRegex},
			{"city": searchRegex},
		}
	}

	return match
}

// hasActiveFilters checks if any filters besides tenant_id are active
func (s *userStorerImpl) hasActiveFilters(filter *dom_user.FederatedUserFilter) bool {
	return filter.Name != nil ||
		filter.Email != nil ||
		filter.SearchTerm != nil ||
		filter.Status != 0 ||
		filter.Role != 0 ||
		filter.CreatedAtStart != nil ||
		filter.CreatedAtEnd != nil
}

func (s *userStorerImpl) CountByFilter(ctx context.Context, filter *dom_user.FederatedUserFilter) (uint64, error) {
	if filter == nil {
		return 0, errors.New("filter cannot be nil")
	}

	// For exact counts with filters
	if s.hasActiveFilters(filter) {
		pipeline := []bson.D{
			{{Key: "$match", Value: s.buildCountMatchStage(filter)}},
			{{Key: "$count", Value: "total"}},
		}

		// Use aggregation with allowDiskUse for large datasets
		opts := options.Aggregate().SetAllowDiskUse(true)
		cursor, err := s.Collection.Aggregate(ctx, pipeline, opts)
		if err != nil {
			return 0, err
		}
		defer cursor.Close(ctx)

		// Decode the result
		var results []struct {
			Total int64 `bson:"total"`
		}
		if err := cursor.All(ctx, &results); err != nil {
			return 0, err
		}

		if len(results) == 0 {
			return 0, nil
		}

		return uint64(results[0].Total), nil
	}

	// For unfiltered counts (much faster for basic filtering)
	countOpts := options.Count()
	match := bson.M{}

	count, err := s.Collection.CountDocuments(ctx, match, countOpts)
	if err != nil {
		return 0, err
	}

	return uint64(count), nil
}

func (impl *userStorerImpl) buildMatchStage(filter *dom_user.FederatedUserFilter) bson.M {
	match := bson.M{}

	// Handle cursor-based pagination
	if filter.LastID != nil && filter.LastCreatedAt != nil {
		match["$or"] = []bson.M{
			{
				"created_at": bson.M{"$lt": filter.LastCreatedAt},
			},
			{
				"created_at": filter.LastCreatedAt,
				"_id":        bson.M{"$lt": filter.LastID},
			},
		}
	}

	// Add other filters
	if filter.Status != 0 {
		match["status"] = filter.Status
	}

	if filter.Role != 0 {
		match["role"] = filter.Role
	}

	if filter.CreatedAtStart != nil || filter.CreatedAtEnd != nil {
		createdAtFilter := bson.M{}
		if filter.CreatedAtStart != nil {
			createdAtFilter["$gte"] = filter.CreatedAtStart
		}
		if filter.CreatedAtEnd != nil {
			createdAtFilter["$lte"] = filter.CreatedAtEnd
		}
		match["created_at"] = createdAtFilter
	}

	// Handle name search
	if filter.Name != nil && *filter.Name != "" {
		match["$or"] = []bson.M{
			{"name": bson.M{"$regex": *filter.Name, "$options": "i"}},
			{"lexical_name": bson.M{"$regex": *filter.Name, "$options": "i"}},
		}
	}

	// Handle email search
	if filter.Email != nil && *filter.Email != "" {
		match["email"] = bson.M{"$regex": *filter.Email, "$options": "i"}
	}

	// More comprehensive search across multiple fields
	if filter.SearchTerm != nil && *filter.SearchTerm != "" {
		searchRegex := bson.M{"$regex": *filter.SearchTerm, "$options": "i"}
		match["$or"] = []bson.M{
			{"name": searchRegex},
			{"lexical_name": searchRegex},
			{"email": searchRegex},
			{"phone": searchRegex},
			{"country": searchRegex},
			{"city": searchRegex},
		}
	}

	return match
}

func (impl *userStorerImpl) ListByFilter(ctx context.Context, filter *dom_user.FederatedUserFilter) (*dom_user.FederatedUserFilterResult, error) {
	if filter == nil {
		return nil, errors.New("filter cannot be nil")
	}

	// Default limit if not specified
	if filter.Limit <= 0 {
		filter.Limit = 100
	}

	// Request one more document than needed to determine if there are more results
	limit := filter.Limit + 1

	// Build the aggregation pipeline
	pipeline := make([]bson.D, 0)

	// Match stage - initial filtering
	matchStage := bson.D{{"$match", impl.buildMatchStage(filter)}}
	pipeline = append(pipeline, matchStage)

	// Sort stage
	sortStage := bson.D{{"$sort", bson.D{
		{"created_at", -1},
		{"_id", -1},
	}}}
	pipeline = append(pipeline, sortStage)

	// Limit stage
	limitStage := bson.D{{"$limit", limit}}
	pipeline = append(pipeline, limitStage)

	// Execute aggregation
	opts := options.Aggregate().SetAllowDiskUse(true)
	cursor, err := impl.Collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode results
	var users []*dom_user.FederatedUser
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	// Handle empty results case
	if len(users) == 0 {
		// For debugging purposes only.
		impl.Logger.Debug("Empty list", zap.Any("filter", filter))
		return &dom_user.FederatedUserFilterResult{
			Users:   make([]*dom_user.FederatedUser, 0),
			HasMore: false,
		}, nil
	}

	// Check if there are more results
	hasMore := false
	if len(users) > int(filter.Limit) {
		hasMore = true
		users = users[:len(users)-1]
	}

	// Get last document info for next page
	lastDoc := users[len(users)-1]

	// Get total count
	totalCount, err := impl.CountByFilter(ctx, filter)
	if err != nil {
		impl.Logger.Warn("Failed to get total count", zap.Any("error", err))
		// Continue without total count
	}

	return &dom_user.FederatedUserFilterResult{
		Users:         users,
		HasMore:       hasMore,
		LastID:        lastDoc.ID,
		LastCreatedAt: lastDoc.CreatedAt,
		TotalCount:    totalCount,
	}, nil
}
