// monorepo/native/desktop/maplefile-cli/internal/domain/localcollection/model.go
package localcollection

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// LocalCollection represents a collection in the system (on the user's device).
type LocalCollection struct {
	// Local primary key
	ID primitive.ObjectID `json:"id"`

	// Remote reference ID - stores the ID of the corresponding remote collection
	RemoteID primitive.ObjectID `json:"remote_id,omitempty"`

	// Existing cloud fields
	OwnerID                primitive.ObjectID          `json:"owner_id"`
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	AncestorIDs            []primitive.ObjectID        `json:"ancestor_ids,omitempty"`
	EncryptedPathSegments  []string                    `json:"encrypted_path_segments,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
	CreatedAt              time.Time                   `json:"created_at"`
	ModifiedAt             time.Time                   `json:"modified_at"`

	// Fields for local display and sync status
	DecryptedName     string    `json:"-" bson:"-"`
	DecryptedPath     string    `json:"-" bson:"-"`
	LastSyncedAt      time.Time `json:"-" bson:"-"`
	IsModifiedLocally bool      `json:"-" bson:"-"`
}

// LocalCreateCollectionRequest represents the data needed to create a collection
type LocalCreateCollectionRequest struct {
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	EncryptedPathSegments  []string                    `json:"encrypted_path_segments,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key"`
}

// CollectionResponse represents the server's response when creating a collection
type LocalCollectionResponse struct {
	ID                     primitive.ObjectID          `json:"id"`
	OwnerID                primitive.ObjectID          `json:"owner_id"`
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	AncestorIDs            []primitive.ObjectID        `json:"ancestor_ids,omitempty"`
	EncryptedPathSegments  []string                    `json:"encrypted_path_segments,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
	CreatedAt              time.Time                   `json:"created_at"`
	ModifiedAt             time.Time                   `json:"modified_at"`
	Members                []LocalMembershipResponse   `json:"members"`
}

// LocalMembershipResponse represents a collection membership
type LocalMembershipResponse struct {
	ID              primitive.ObjectID `json:"id"`
	RecipientID     primitive.ObjectID `json:"recipient_id"`
	RecipientEmail  string             `json:"recipient_email"`
	PermissionLevel string             `json:"permission_level"`
	GrantedByID     primitive.ObjectID `json:"granted_by_id"`
	CollectionID    primitive.ObjectID `json:"collection_id"`
	IsInherited     bool               `json:"is_inherited"`
	InheritedFromID primitive.ObjectID `json:"inherited_from_id,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
}
