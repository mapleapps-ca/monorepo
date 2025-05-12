package collection

// CollectionRepository interface defines all operations that can be performed on collections
type CollectionRepository interface {
	// Collection Management
	Create(collection *Collection) error
	Get(id string) (*Collection, error)
	GetAllByUserID(userID string) ([]*Collection, error)
	Update(collection *Collection) error
	Delete(id string) error

	// Collection Sharing
	GetCollectionsSharedWithUser(userID string) ([]*Collection, error)
	AddMember(collectionID string, membership *CollectionMembership) error
	RemoveMember(collectionID string, recipientID string) error
	UpdateMemberPermission(collectionID string, recipientID string, newPermission string) error

	// Access Checking
	CheckIfExistsByID(id string) (bool, error)
	CheckAccess(collectionID string, userID string, requiredPermission string) (bool, error)
	IsCollectionOwner(collectionID string, userID string) (bool, error)
}
