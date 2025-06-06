package bannedipaddress

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	dom_banip "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/bannedipaddress"
)

// hasActiveFilters checks if any filters besides tenant_id are active
func (impl bannedIPAddressImpl) hasActiveFilters(filter *dom_banip.BannedIPAddressFilter) bool {
	return !filter.UserID.IsZero() ||
		filter.CreatedAtStart != nil ||
		filter.CreatedAtEnd != nil
}

// buildCountMatchStage creates the match stage for the aggregation pipeline
func (impl bannedIPAddressImpl) buildCountMatchStage(filter *dom_banip.BannedIPAddressFilter) bson.M {
	match := bson.M{}

	if !filter.UserID.IsZero() {
		match["user_id"] = filter.UserID
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

	return match
}

func (impl bannedIPAddressImpl) CountByFilter(ctx context.Context, filter *dom_banip.BannedIPAddressFilter) (uint64, error) {
	if filter == nil {
		return 0, errors.New("filter cannot be nil")
	}

	// For exact counts with filters
	if impl.hasActiveFilters(filter) {
		pipeline := []bson.D{
			{{Key: "$match", Value: impl.buildCountMatchStage(filter)}},
			{{Key: "$count", Value: "total"}},
		}

		// Use aggregation with allowDiskUse for large datasets
		opts := options.Aggregate().SetAllowDiskUse(true)
		cursor, err := impl.Collection.Aggregate(ctx, pipeline, opts)
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

	// For unfiltered counts (much faster for basic tenant-only filtering)
	countOpts := options.Count().SetHint("created_at_-1")
	match := bson.M{}

	count, err := impl.Collection.CountDocuments(ctx, match, countOpts)
	if err != nil {
		return 0, err
	}

	return uint64(count), nil
}

func (impl bannedIPAddressImpl) buildMatchStage(filter *dom_banip.BannedIPAddressFilter) bson.M {
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

	if !filter.UserID.IsZero() {
		match["user_id"] = filter.UserID
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

	// Text search for name
	if filter.Value != nil && *filter.Value != "" {
		match["$text"] = bson.M{"$search": *filter.Value}
	}

	return match
}

func (impl bannedIPAddressImpl) ListByFilter(ctx context.Context, filter *dom_banip.BannedIPAddressFilter) (*dom_banip.BannedIPAddressFilterResult, error) {
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
	var bannedIPAddresss []*dom_banip.BannedIPAddress
	if err := cursor.All(ctx, &bannedIPAddresss); err != nil {
		return nil, err
	}

	// Handle empty results case
	if len(bannedIPAddresss) == 0 {
		// impl.Logger.Debug("Empty list", zap.Any("filter", filter))
		return &dom_banip.BannedIPAddressFilterResult{
			BannedIPAddresses: make([]*dom_banip.BannedIPAddress, 0),
			HasMore:           false,
		}, nil
	}

	// Check if there are more results
	hasMore := false
	if len(bannedIPAddresss) > int(filter.Limit) {
		hasMore = true
		bannedIPAddresss = bannedIPAddresss[:len(bannedIPAddresss)-1]
	}

	// Get last document info for next page
	lastDoc := bannedIPAddresss[len(bannedIPAddresss)-1]

	return &dom_banip.BannedIPAddressFilterResult{
		BannedIPAddresses: bannedIPAddresss,
		HasMore:           hasMore,
		LastID:            lastDoc.ID,
		LastCreatedAt:     lastDoc.CreatedAt,
	}, nil
}

func (impl bannedIPAddressImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := impl.Collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		impl.Logger.Error("database failed deletion error",
			zap.Any("error", err))
		return err
	}
	return nil
}

func (impl bannedIPAddressImpl) ListAllValues(ctx context.Context) ([]string, error) {
	// Create an empty collection to hold our results
	var results []dom_banip.BannedIPAddress

	// Set up our find options to only get the "value" field
	opts := options.Find().SetProjection(bson.D{{Key: "value", Value: 1}})

	// Execute the find operation
	cursor, err := impl.Collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode all documents into our struct
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Extract just the values into a string slice
	values := make([]string, len(results))
	for i, result := range results {
		values[i] = result.Value
	}

	return values, nil
}
