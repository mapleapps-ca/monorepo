package keys

import (
	"time"

	"github.com/gocql/gocql"
)

// EncryptedHistoricalKey represents a previous version of a key
type EncryptedHistoricalKey struct {
	KeyVersion    int       `json:"key_version" bson:"key_version"`
	Ciphertext    []byte    `json:"ciphertext" bson:"ciphertext"`
	Nonce         []byte    `json:"nonce" bson:"nonce"`
	RotatedAt     time.Time `json:"rotated_at" bson:"rotated_at"`
	RotatedReason string    `json:"rotated_reason" bson:"rotated_reason"`
	// Algorithm used for this key version
	Algorithm string `json:"algorithm" bson:"algorithm"`
}

// KeyRotationPolicy defines when and how to rotate keys
type KeyRotationPolicy struct {
	MaxKeyAge           time.Duration `json:"max_key_age" bson:"max_key_age"`
	MaxKeyUsageCount    int64         `json:"max_key_usage_count" bson:"max_key_usage_count"`
	ForceRotateOnBreach bool          `json:"force_rotate_on_breach" bson:"force_rotate_on_breach"`
}

// KeyRotationRecord tracks rotation events
type KeyRotationRecord struct {
	ID            gocql.UUID `bson:"_id" json:"id"`
	EntityType    string     `bson:"entity_type" json:"entity_type"` // "user", "collection", "file"
	EntityID      gocql.UUID `bson:"entity_id" json:"entity_id"`
	FromVersion   int        `bson:"from_version" json:"from_version"`
	ToVersion     int        `bson:"to_version" json:"to_version"`
	RotatedAt     time.Time  `bson:"rotated_at" json:"rotated_at"`
	RotatedBy     gocql.UUID `bson:"rotated_by" json:"rotated_by"`
	Reason        string     `bson:"reason" json:"reason"`
	AffectedItems int64      `bson:"affected_items" json:"affected_items"`
}
