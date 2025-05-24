// monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection/model.go
package remotecollection

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// RemoteCollection represents a collection in the system (on the cloud backend).
type RemoteCollection struct {
	// Existing cloud fields
	ID                     primitive.ObjectID          `json:"id"`
	OwnerID                primitive.ObjectID          `json:"owner_id"`
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	AncestorIDs            []primitive.ObjectID        `json:"ancestor_ids,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
	CreatedAt              time.Time                   `json:"created_at"`
	ModifiedAt             time.Time                   `json:"modified_at"`

	// New fields for local decrypted data
	DecryptedName     string    `json:"-" bson:"-"`
	DecryptedPath     string    `json:"-" bson:"-"`
	LastSyncedAt      time.Time `json:"-" bson:"-"`
	IsModifiedLocally bool      `json:"-" bson:"-"`
}

// RemoteCreateCollectionRequest represents the data needed to create a collection
type RemoteCreateCollectionRequest struct {
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key"`
}

// RemoteCollectionResponse represents the server's response when creating a collection
type RemoteCollectionResponse struct {
	ID                     primitive.ObjectID          `json:"id"`
	OwnerID                primitive.ObjectID          `json:"owner_id"`
	EncryptedName          string                      `json:"encrypted_name"`
	Type                   string                      `json:"type"`
	ParentID               primitive.ObjectID          `json:"parent_id,omitempty"`
	AncestorIDs            []primitive.ObjectID        `json:"ancestor_ids,omitempty"`
	EncryptedCollectionKey keys.EncryptedCollectionKey `json:"encrypted_collection_key,omitempty"`
	CreatedAt              time.Time                   `json:"created_at"`
	ModifiedAt             time.Time                   `json:"modified_at"`
	Members                []RemoteMembershipResponse  `json:"members"`
}

// RemoteMembershipResponse represents a collection membership
type RemoteMembershipResponse struct {
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
