package bannedipaddress

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BannedIPAddress structure represents the blockchain transaction that
// belongs to our user in our application.
type BannedIPAddress struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"` // The user ID that this IP address belongs to.
	Value     string             `bson:"value" json:"value"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
}

type BannedIPAddressFilter struct {
	UserID         primitive.ObjectID `json:"user_id,omitempty"`
	CreatedAtStart *time.Time         `json:"created_at_start,omitempty"`
	CreatedAtEnd   *time.Time         `json:"created_at_end,omitempty"`
	Value          *string            `bson:"value" json:"value"`

	// Cursor-based pagination
	LastID        *primitive.ObjectID `json:"last_id,omitempty"`
	LastCreatedAt *time.Time          `json:"last_created_at,omitempty"`
	Limit         int64               `json:"limit"`
}

type BannedIPAddressFilterResult struct {
	BannedIPAddresses []*BannedIPAddress `json:"banned_ip_addresses"`
	HasMore           bool               `json:"has_more"`
	LastID            primitive.ObjectID `json:"last_id,omitempty"`
	LastCreatedAt     time.Time          `json:"last_created_at,omitempty"`
}
