package mongodb

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ObjectID structure (24-character hex string)
// - 4 bytes: Unix timestamp (creation time)
// - 5 bytes: Random unique value per machine/process
// - 3 bytes: Counter (incrementing)
// This structure allows for distributed generation of unique identifiers.

// IsValidObjectID checks if the given string is a valid 24-character hexadecimal
// representation that can be parsed into a MongoDB ObjectID.
func IsValidObjectID(id string) bool {
	// primitive.ObjectIDFromHex returns an error if the string is not
	// a valid 24-character hexadecimal string.
	_, err := primitive.ObjectIDFromHex(id)
	return err == nil
}

// ValidateClientObjectID performs validation checks on a primitive.ObjectID,
// typically used for IDs received from external sources like clients.
// It checks:
//  1. If the ObjectID is the zero value.
//  2. Optionally, if the timestamp embedded in the ObjectID falls within
//     a reasonable time range (not too far in the future, not too old).
//
// These timestamp checks are application-specific and assume the ID was
// generated recently.
func ValidateClientObjectID(id primitive.ObjectID) error {
	// Check if zero value ObjectID
	if id.IsZero() {
		return errors.New("objectID cannot be the zero value")
	}

	// Check timestamp bounds (optional, application-specific check)
	// This assumes the ObjectID was generated recently.
	timestamp := id.Timestamp()
	now := time.Now()

	// Allow for a small margin for future timestamps due to clock sync issues or server time differences
	futureThreshold := now.Add(time.Hour)
	if timestamp.After(futureThreshold) {
		return errors.New("objectID timestamp is too far in the future")
	}

	// Define an acceptable age limit for ObjectIDs. 1 year is an example.
	// Adjust this based on your application's data retention policy or expected ID lifespan.
	pastThreshold := now.Add(-365 * 24 * time.Hour) // 1 year ago
	if timestamp.Before(pastThreshold) {
		return errors.New("objectID timestamp is too old")
	}

	return nil
}
